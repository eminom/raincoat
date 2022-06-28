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

type EngineInfo struct {
	ClusterID  int
	EngineID   int
	EngineName string
}

func (e EngineInfo) ToString() string {
	return fmt.Sprintf("%v(%v,%v)", e.EngineName, e.ClusterID, e.EngineID)
}

type ArchDispatcher struct {
	dispatch map[string]*MidRec
	revMap   map[int]EngineInfo // From master id to engine info
}

func (ad ArchDispatcher) CheckOut(mid int) (string, bool) {
	item, ok := ad.revMap[mid]
	if ok {
		return item.ToString() + fmt.Sprintf(" indexing(%v)", mid), true
	}
	return "", false
}

func MakeCollectDispatch(
	archTarget ArchTarget,
	masterSeq []DpfEngineT) ArchDispatcher {
	dispatch := map[string]*MidRec{
		ENGINE_CDMA:  NewMidRec("cdma", archTarget.CdmaPerC),
		ENGINE_SDMA:  NewMidRec("sdma", archTarget.SdmaPerC),
		ENGINE_SIP:   NewMidRec("sip", archTarget.SipPerC),
		ENGINE_CQM:   NewMidRec("cqm", archTarget.CqmPerC),
		ENGINE_GSYNC: NewMidRec("gsync", archTarget.GsyncPerC),
	}
	revMap := make(map[int]EngineInfo)
	for _, v := range masterSeq {
		if dict, ok := dispatch[v.EngType]; ok {
			dict.PickAt(v.ClusterID, v.EngineId, v.UniqueEngIdx())
			revMap[v.UniqueEngIdx()] = EngineInfo{
				ClusterID:  v.ClusterID,
				EngineID:   v.EngineId,
				EngineName: v.EngType,
			}
		}
	}

	for _, disp := range dispatch {
		disp.Sumup()
	}
	return ArchDispatcher{
		dispatch: dispatch,
		revMap:   revMap,
	}
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
					disp, ok := dispatch.dispatch[name]
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
	sourceTextTmpl.Execute(out, dispatch.dispatch)
}

const kmdAffinityMapTmpl = `
// Automatically generated
int affinity_map[] = {
	{{range $Idx, $Pg := .}}{{$Pg}}, {{IndexToComment $Idx}}
	{{end}}
}
`

func genCompleteMapForDorado(midCodec []DpfEngineT, out io.Writer) {

	target := NewDoradoArchTarget()
	dispatch := MakeCollectDispatch(target, doradoDpfTy)
	srcTmpl := template.Must(
		template.New("kmdaffinity").Funcs(
			template.FuncMap{
				"IndexToComment": func(mid int) string {
					comment, ok := dispatch.CheckOut(mid)
					if ok {
						return "// " + comment
					}
					return ""
				},
			}).Parse(kmdAffinityMapTmpl))
	srcTmpl.Execute(out, genAffnityMapForDorado(target, dispatch))
}

func genAffnityMapForDorado(target ArchTarget, archDisp ArchDispatcher) []int {

	affinityMap := make([]int, MAX_MASTER_ID_VALUE)
	for i := 0; i < len(affinityMap); i++ {
		affinityMap[i] = -1
	}

	iterateOn := func(mr *MidRec, calc func(cid, eid, mval int) int) {
		for cid := 0; cid < target.ClusterPerD; cid++ {
			for eid := 0; eid < mr.EngineCountPerC; eid++ {
				mval := mr.MidFor(cid, eid)
				pgIdx := calc(cid, eid, mval)
				affinityMap[mval] = pgIdx
			}
		}
	}
	dispatch := archDisp.dispatch

	// CDMA
	// CDMA(0, 3) belongs to 000111
	// CDMA(1, 3) belongs to 111000
	// And with some distortion(introduced in June. 2022)
	iterateOn(dispatch[ENGINE_CDMA], func(cid, eid, mval int) int {
		var pgIdx int
		if eid == 3 {
			pgIdx = cid * 3
		} else {
			// TODO: (hardcode with version)
			pgIdx = cid*3 + eid
		}
		return pgIdx
	})
	iterateOn(dispatch[ENGINE_CQM], func(cid, eid, mval int) int {
		return cid*3 + eid
	})
	iterateOn(dispatch[ENGINE_GSYNC], func(cid, eid, mval int) int {
		return cid*3 + eid
	})
	// SIP(0, 0~3) belongs to 000001
	// SIP(0, 4~7) belongs to 000010
	iterateOn(dispatch[ENGINE_SIP], func(cid, eid, mval int) int {
		return cid*3 + eid/4
	})
	// SDAM, same as SIP
	iterateOn(dispatch[ENGINE_SDMA], func(cid, eid, mval int) int {
		return cid*3 + eid/4
	})
	return affinityMap
}
