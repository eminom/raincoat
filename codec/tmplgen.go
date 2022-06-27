package codec

import (
	"fmt"
	"html/template"
	"io"
	"log"
)

const (
	Dorado_CdmaCountPerCluster  = 4
	Dorado_SdmaCountPerCluster  = 12
	Dorado_SipCountPerCluster   = 12
	Dorado_CqmCountPerCluster   = 3
	Dorado_GsyncCountPerCluster = 3
)

type ArchTarget struct {
	CdmaPerC    int
	SdmaPerC    int
	SipPerC     int
	CqmPerC     int
	GsyncPerC   int
	ClusterPerD int
}

func NewDoradoArchTarget() ArchTarget {
	return ArchTarget{
		CdmaPerC:    4,
		SdmaPerC:    12,
		SipPerC:     12,
		CqmPerC:     3,
		GsyncPerC:   3,
		ClusterPerD: 2,
	}
}

func MakeCollectDispatch(
	archTarget ArchTarget,
	masterSeq []DpfEngineT) map[string]*MidRec {
	dispatch := map[string]*MidRec{
		ENGINE_CDMA:  NewMidRec("cdma", archTarget.CdmaPerC),
		ENGINE_SDMA:  NewMidRec("sdma", archTarget.SdmaPerC),
		ENGINE_SIP:   NewMidRec("sip", archTarget.SipPerC),
		ENGINE_CQM:   NewMidRec("cqm", archTarget.CqmPerC),
		ENGINE_GSYNC: NewMidRec("gsync", archTarget.GsyncPerC),
	}
	for _, v := range masterSeq {
		if dict, ok := dispatch[v.EngType]; ok {
			dict.PickAt(v.ClusterID, v.EngineId, v.UniqueEngIdx())
		}
	}

	for _, disp := range dispatch {
		disp.Sumup()
	}
	return dispatch
}

const midInfoSrcTmpl = `
// Automatically generated
func initMidInfo(mi *MidInfoMan) {
	{{range $EngineName, $Distr := .}}
	mi.{{$EngineName}} = []int {
		{{range $Seq:=$Distr.ValueSeq}} {{FormatEntry $EngineName $Seq}}
		{{end}}
	} {{end}}
} // done initMidInfo
`

func genDictForDorado(midCodec []DpfEngineT, out io.Writer) {
	// Prepares
	dispatch := MakeCollectDispatch(NewDoradoArchTarget(), doradoDpfTy)

	sourceTextTmpl := template.Must(
		template.New("master-id-map").Funcs(
			template.FuncMap{
				//idx int
				"FormatEntry": func(name string, mid int) string {
					disp, ok := dispatch[name]
					if !ok {
						log.Fatalf("unexpected for [%v]", name)
					}
					idx, ok := disp.midToIndex[mid]
					if !ok {
						log.Fatalf("failed for %v over %v", idx, name)
					}
					cid, eid := disp.decode(idx)
					return fmt.Sprintf("%v, // %v, %v", mid, cid, eid)
				}}).Parse(midInfoSrcTmpl))
	sourceTextTmpl.Execute(out, dispatch)

}
