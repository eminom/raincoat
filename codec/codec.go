package codec

import (
	"fmt"
	"io"
	"sort"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

/*
Descriptive in C++

typedef struct {
  int32_t cluster_id;
  int32_t master_id_hi;
  int32_t master_id_lo;
  int32_t engine_id;
  DPF_ENGINE_T engine_type;
} dpf_id_t;
*/

type EngineType int

const (
	ENGINE_SIP       = "SIP"
	ENGINE_SDMA      = "SDMA"
	ENGINE_CDMA      = "CDMA"
	ENGINE_CQM       = "CQM"
	ENGINE_GSYNC     = "GSYNC"
	ENGINE_SIP_LITE  = "SIP_LITE"
	ENGINE_SDMA_LITE = "SDMA_LITE"
	ENGINE_CDMA_LITE = "ENGINE_CDMA_LITE"
	ENGINE_PCIE      = "PCIE"
	ENGINE_TS        = "TS"
	ENGINE_ODMA      = "ODMA"
	ENGINE_HCVG      = "HCVG"
	ENGINE_VDEC      = "VDEC"
	ENGINE_UNKNOWN   = "UNKNOWN"
)

type DpfEngineT struct {
	ClusterID int
	MasterHi  int
	MasterLo  int
	EngineId  int
	EngType   string
}

func (d DpfEngineT) UniqueEngIdx() int {
	return d.MasterLo | (d.MasterHi << 5)
}

type EngineTypeIndexMap struct {
	fromIdxToTypeCode map[int]EngineTypeCode
}

func newEngineTypeIndexMap(infos []DpfEngineT) EngineTypeIndexMap {
	fromIdxToTypeCode := make(map[int]EngineTypeCode)
	for _, info := range infos {
		u := info.UniqueEngIdx()
		if _, ok := fromIdxToTypeCode[u]; ok {
			panic("invalid engine index value: duplicated")
		}
		fromIdxToTypeCode[u] = ToEngineTypeCode(info.EngType)
	}
	return EngineTypeIndexMap{
		fromIdxToTypeCode: fromIdxToTypeCode,
	}
}

func (e EngineTypeIndexMap) Lookup(engineUniqIdx int) EngineTypeCode {
	if nameStr, ok := e.fromIdxToTypeCode[engineUniqIdx]; ok {
		return nameStr
	}
	return EngCat_UNKNOWN
}

var (
	doradoDpfTy []DpfEngineT
	pavoDpfTy   []DpfEngineT

	doradoEngIdxMap EngineTypeIndexMap
	pavoEngIdxMap   EngineTypeIndexMap
)

func init() {
	doradoDpfTy = []DpfEngineT{
		{0, 0, 0, 0, ENGINE_SIP},
		{0, 0, 1, 0, ENGINE_SDMA},
		{0, 0, 2, 1, ENGINE_SIP},
		{0, 0, 3, 1, ENGINE_SDMA},
		{0, 0, 4, 2, ENGINE_SIP},
		{0, 0, 5, 2, ENGINE_SDMA},
		{0, 0, 6, 3, ENGINE_SIP},
		{0, 0, 7, 3, ENGINE_SDMA},
		{0, 0, 8, 0, ENGINE_CQM},
		{0, 0, 13, 0, ENGINE_GSYNC},
		{0, 1, 0, 4, ENGINE_SIP},
		{0, 1, 1, 4, ENGINE_SDMA},
		{0, 1, 2, 5, ENGINE_SIP},
		{0, 1, 3, 5, ENGINE_SDMA},
		{0, 1, 4, 6, ENGINE_SIP},
		{0, 1, 5, 6, ENGINE_SDMA},
		{0, 1, 6, 7, ENGINE_SIP},
		{0, 1, 7, 7, ENGINE_SDMA},
		{0, 1, 8, 1, ENGINE_CQM},
		{0, 1, 13, 1, ENGINE_GSYNC},
		{0, 2, 0, 8, ENGINE_SIP},
		{0, 2, 1, 8, ENGINE_SDMA},
		{0, 2, 2, 9, ENGINE_SIP},
		{0, 2, 3, 9, ENGINE_SDMA},
		{0, 2, 4, 10, ENGINE_SIP},
		{0, 2, 5, 10, ENGINE_SDMA},
		{0, 2, 6, 11, ENGINE_SIP},
		{0, 2, 7, 11, ENGINE_SDMA},
		{0, 2, 8, 2, ENGINE_CQM},
		{0, 2, 13, 2, ENGINE_GSYNC},

		{0, 3, 0, 0, ENGINE_CDMA},
		{0, 4, 0, 1, ENGINE_CDMA},
		{0, 5, 0, 2, ENGINE_CDMA},
		{0, 6, 0, 3, ENGINE_CDMA},
		{0, 7, 0, 0, ENGINE_SIP_LITE},
		{0, 7, 1, 0, ENGINE_SDMA_LITE},

		{1, 8, 0, 0, ENGINE_SIP},
		{1, 8, 1, 0, ENGINE_SDMA},
		{1, 8, 2, 1, ENGINE_SIP},
		{1, 8, 3, 1, ENGINE_SDMA},
		{1, 8, 4, 2, ENGINE_SIP},
		{1, 8, 5, 2, ENGINE_SDMA},
		{1, 8, 6, 3, ENGINE_SIP},
		{1, 8, 7, 3, ENGINE_SDMA},
		{1, 8, 8, 0, ENGINE_CQM},
		{1, 8, 13, 0, ENGINE_GSYNC},
		{1, 9, 0, 4, ENGINE_SIP},
		{1, 9, 1, 4, ENGINE_SDMA},
		{1, 9, 2, 5, ENGINE_SIP},
		{1, 9, 3, 5, ENGINE_SDMA},
		{1, 9, 4, 6, ENGINE_SIP},
		{1, 9, 5, 6, ENGINE_SDMA},
		{1, 9, 6, 7, ENGINE_SIP},
		{1, 9, 7, 7, ENGINE_SDMA},
		{1, 9, 8, 1, ENGINE_CQM},
		{1, 9, 13, 1, ENGINE_GSYNC},
		{1, 10, 0, 8, ENGINE_SIP},
		{1, 10, 1, 8, ENGINE_SDMA},
		{1, 10, 2, 9, ENGINE_SIP},
		{1, 10, 3, 9, ENGINE_SDMA},
		{1, 10, 4, 10, ENGINE_SIP},
		{1, 10, 5, 10, ENGINE_SDMA},
		{1, 10, 6, 11, ENGINE_SIP},
		{1, 10, 7, 11, ENGINE_SDMA},
		{1, 10, 8, 2, ENGINE_CQM},
		{1, 10, 13, 2, ENGINE_GSYNC},

		{1, 11, 0, 0, ENGINE_CDMA},
		{1, 12, 0, 1, ENGINE_CDMA},
		{1, 13, 0, 2, ENGINE_CDMA},
		{1, 14, 0, 3, ENGINE_CDMA},
		{1, 15, 0, 0, ENGINE_SIP_LITE},
		{1, 15, 1, 0, ENGINE_SDMA_LITE},

		{2, 22, 0, 0, ENGINE_PCIE},
		{2, 24, 0, 0, ENGINE_TS},
		{2, 25, 0, 0, ENGINE_ODMA},
		{2, 27, 0, 0, ENGINE_HCVG},
		{2, 28, 0, 1, ENGINE_HCVG},
		{2, 29, 0, 0, ENGINE_VDEC},

		// make sure this is the last one
		{-1, -1, -1, -1, ENGINE_UNKNOWN},
	}

	pavoDpfTy = []DpfEngineT{

		/* cid,	mid_hi,	mid_lo,	eid,	engine_type */
		{0, 0, 0, 0, ENGINE_SIP},
		{0, 0, 1, 0, ENGINE_SDMA},
		{0, 0, 2, 1, ENGINE_SIP},
		{0, 0, 3, 1, ENGINE_SDMA},
		{0, 0, 4, 2, ENGINE_SIP},
		{0, 0, 5, 2, ENGINE_SDMA},
		{0, 0, 6, 3, ENGINE_SIP},
		{0, 0, 7, 3, ENGINE_SDMA},
		{0, 0, 8, 4, ENGINE_SIP},
		{0, 0, 9, 4, ENGINE_SDMA},
		{0, 0, 10, 5, ENGINE_SIP},
		{0, 0, 11, 5, ENGINE_SDMA},
		{0, 0, 12, 6, ENGINE_SIP},
		{0, 0, 13, 6, ENGINE_SDMA},
		{0, 0, 14, 0, ENGINE_CDMA},
		{0, 0, 15, 1, ENGINE_CDMA},
		{0, 0, 16, 2, ENGINE_CDMA},
		{0, 0, 17, 3, ENGINE_CDMA},
		{0, 0, 18, 0, ENGINE_CQM},
		{0, 0, 22, 0, ENGINE_GSYNC},

		{1, 1, 0, 0, ENGINE_SIP},
		{1, 1, 1, 0, ENGINE_SDMA},
		{1, 1, 2, 1, ENGINE_SIP},
		{1, 1, 3, 1, ENGINE_SDMA},
		{1, 1, 4, 2, ENGINE_SIP},
		{1, 1, 5, 2, ENGINE_SDMA},
		{1, 1, 6, 3, ENGINE_SIP},
		{1, 1, 7, 3, ENGINE_SDMA},
		{1, 1, 8, 4, ENGINE_SIP},
		{1, 1, 9, 4, ENGINE_SDMA},
		{1, 1, 10, 5, ENGINE_SIP},
		{1, 1, 11, 5, ENGINE_SDMA},
		{1, 1, 12, 6, ENGINE_SIP},
		{1, 1, 13, 6, ENGINE_SDMA},
		{1, 1, 14, 0, ENGINE_CDMA},
		{1, 1, 15, 1, ENGINE_CDMA},
		{1, 1, 16, 2, ENGINE_CDMA},
		{1, 1, 17, 3, ENGINE_CDMA},
		{1, 1, 18, 0, ENGINE_CQM},
		{1, 1, 22, 0, ENGINE_GSYNC},

		{2, 2, 0, 0, ENGINE_SIP},
		{2, 2, 1, 0, ENGINE_SDMA},
		{2, 2, 2, 1, ENGINE_SIP},
		{2, 2, 3, 1, ENGINE_SDMA},
		{2, 2, 4, 2, ENGINE_SIP},
		{2, 2, 5, 2, ENGINE_SDMA},
		{2, 2, 6, 3, ENGINE_SIP},
		{2, 2, 7, 3, ENGINE_SDMA},
		{2, 2, 8, 4, ENGINE_SIP},
		{2, 2, 9, 4, ENGINE_SDMA},
		{2, 2, 10, 5, ENGINE_SIP},
		{2, 2, 11, 5, ENGINE_SDMA},
		{2, 2, 12, 6, ENGINE_SIP},
		{2, 2, 13, 6, ENGINE_SDMA},
		{2, 2, 14, 0, ENGINE_CDMA},
		{2, 2, 15, 1, ENGINE_CDMA},
		{2, 2, 16, 2, ENGINE_CDMA},
		{2, 2, 17, 3, ENGINE_CDMA},
		{2, 2, 18, 0, ENGINE_CQM},
		{2, 2, 22, 0, ENGINE_GSYNC},

		{3, 3, 0, 0, ENGINE_SIP},
		{3, 3, 1, 0, ENGINE_SDMA},
		{3, 3, 2, 1, ENGINE_SIP},
		{3, 3, 3, 1, ENGINE_SDMA},
		{3, 3, 4, 2, ENGINE_SIP},
		{3, 3, 5, 2, ENGINE_SDMA},
		{3, 3, 6, 3, ENGINE_SIP},
		{3, 3, 7, 3, ENGINE_SDMA},
		{3, 3, 8, 4, ENGINE_SIP},
		{3, 3, 9, 4, ENGINE_SDMA},
		{3, 3, 10, 5, ENGINE_SIP},
		{3, 3, 11, 5, ENGINE_SDMA},
		{3, 3, 12, 6, ENGINE_SIP},
		{3, 3, 13, 6, ENGINE_SDMA},
		{3, 3, 14, 0, ENGINE_CDMA},
		{3, 3, 15, 1, ENGINE_CDMA},
		{3, 3, 16, 2, ENGINE_CDMA},
		{3, 3, 17, 3, ENGINE_CDMA},
		{3, 3, 18, 0, ENGINE_CQM},
		{3, 3, 22, 0, ENGINE_GSYNC},

		{6, 4, 0, 0, ENGINE_PCIE},
		{6, 6, 0, 0, ENGINE_TS},

		{4, 15, 0, 0, ENGINE_SIP_LITE},
		{4, 15, 1, 0, ENGINE_SDMA_LITE},
		{4, 15, 2, 0, ENGINE_CDMA_LITE},
		{4, 15, 3, 1, ENGINE_CDMA_LITE},

		{5, 16, 0, 0, ENGINE_SIP_LITE},
		{5, 16, 1, 0, ENGINE_SDMA_LITE},
		{5, 16, 2, 0, ENGINE_CDMA_LITE},
		{5, 16, 3, 1, ENGINE_CDMA_LITE},

		// make sure this is the last one
		{-1, -1, -1, -1, ENGINE_UNKNOWN},
	}

	// Build from this descriptive lang
	doradoEngIdxMap = newEngineTypeIndexMap(doradoDpfTy)
	pavoEngIdxMap = newEngineTypeIndexMap(pavoDpfTy)
}

type MidRec struct {
	name      string
	dc        map[int]int // from 0 to last one
	toIdxFunc func(int, int) int
	decode    func(int) (int, int)

	midToIndex map[int]int
	ValueSeq   []int
}

func genIndexerFunc(enginePerCluster int) func(int, int) int {
	return func(cid, eid int) int {
		return cid*enginePerCluster + eid
	}
}

func genDecoder(enginePerCluster int) func(int) (int, int) {
	return func(idx int) (int, int) {
		cid := idx / enginePerCluster
		eid := idx % enginePerCluster
		return cid, eid
	}
}

func NewMidRec(name string, eCount int) *MidRec {
	return &MidRec{
		name:      name,
		dc:        make(map[int]int),
		toIdxFunc: genIndexerFunc(eCount),
		decode:    genDecoder(eCount),
	}
}

func (m *MidRec) PickAt(cid, eid int, mid int) {
	idx := m.toIdxFunc(cid, eid)
	if _, ok := m.dc[idx]; ok {
		panic(fmt.Errorf("duplicate entry for %v(%v,%v)", m.name, cid, eid))
	}
	m.dc[idx] = mid
}

func (m *MidRec) Sumup() {
	var seq []int
	midToIdx := make(map[int]int)
	for k, mid := range m.dc {
		seq = append(seq, k)
		midToIdx[mid] = k
	}
	sort.Ints(seq)
	var vals []int
	for _, v := range seq {
		vals = append(vals, m.dc[v])
	}
	m.ValueSeq = vals
	m.midToIndex = midToIdx
}

type EngElement struct {
	idx int
}

type EngElements []EngElement

func (e EngElements) Len() int {
	return len(e)
}
func (e EngElements) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e EngElements) Less(i, j int) bool {
	return e[i].idx < e[j].idx
}

func GenDictForDorado(out io.Writer) {
	genDictForDorado(doradoDpfTy, out)
}

type MidCheckout interface {
	CheckoutFor(string)
}

type DoradoMidCheckout struct{}

func (DoradoMidCheckout) CheckoutFor(name string) {
	for _, ent := range doradoDpfTy {
		if ent.EngType == name {
			fmt.Printf("Dorado: %v: %v\n", name, ent.UniqueEngIdx())
		}
	}
}

type PavoMidCheckout struct{}

func (PavoMidCheckout) CheckoutFor(name string) {
	for _, ent := range pavoDpfTy {
		if ent.EngType == name {
			fmt.Printf("Pavo: %v: %v\n", name, ent.UniqueEngIdx())
		}
	}
}

func getEngInfo(lo, hi int, dpfInfo []DpfEngineT) (int, int, int) {
	for _, item := range dpfInfo {
		if item.MasterLo == lo && item.MasterHi == hi {
			return item.EngineId, item.UniqueEngIdx(), item.ClusterID
		}
	}
	return -1, -1, -1
}

func GetEngInfoDorado(lo, hi int64) (int, int, int) {
	return getEngInfo(int(lo), int(hi), doradoDpfTy)
}

func GetEngInfoPavo(lo, hi int64) (int, int, int) {
	return getEngInfo(int(lo), int(hi), pavoDpfTy)
}

type DecodeMaster struct {
	Arch            string
	decoder         func(int64, int64) (int, int, int)
	engIdxToNameMap EngineTypeIndexMap
}

func NewDecodeMaster(arch string) *DecodeMaster {
	var decoder func(int64, int64) (int, int, int)
	var engIdxNameMap EngineTypeIndexMap
	switch arch {
	case "pavo":
		decoder = GetEngInfoPavo
		engIdxNameMap = pavoEngIdxMap
	case "dorado":
		decoder = GetEngInfoDorado
		engIdxNameMap = doradoEngIdxMap
	default:
		return nil
	}
	return &DecodeMaster{
		Arch:            arch,
		decoder:         decoder,
		engIdxToNameMap: engIdxNameMap,
	}
}

func (md *DecodeMaster) EngUniqueIndexToTypeName(engineUniqIdx int) EngineTypeCode {
	return md.engIdxToNameMap.Lookup(engineUniqIdx)
}

func (md *DecodeMaster) DecodeMasterValue(masterVal int) (EngineTypeCode, int, int) {
	lo, hi := int64(masterVal&0x1f), int64((masterVal>>5)&0x1f)
	engineIdx, mVal, clusterId := md.decoder(lo, hi)
	assert.Assert(mVal == masterVal, "Yes")
	typeStr := md.EngUniqueIndexToTypeName(masterVal)
	return typeStr, engineIdx, clusterId
}

// Format V1: flag = 0
//   master_id_ : 10;
//   reserved1_ : 2;
//   context_id_ : 4;
//   reserved2_ : 16;
func (md *DecodeMaster) GetEngineInfo(val uint32) (
	engineIdx int,
	engineUniqueIndex int,
	ctxIdx int,
	clusterId int,
	ok bool,
) {
	engineIdx, engineUniqueIndex, ctxIdx, clusterId = -1, -1, -1, -1
	reserved_1 := (val >> 10) & 3
	reserved_2 := (val >> 16) & 0xFFFF
	if reserved_1 != 0 || reserved_2 != 0 {
		_ = 101 // Does not really matter: 2021-12-29
	}
	ctxIdx = int((val >> 12) & 0xF)
	lo, hi := int64(val&0x1f), int64(((val >> 5) & 0x1f))
	engineIdx, engineUniqueIndex, clusterId = md.decoder(lo, hi)
	ok = engineUniqueIndex >= 0
	return
}

// Format V2: flag = 1
// master_id_ : 10;
// reserved_ : 22;
func (md *DecodeMaster) GetEngineInfoV2(val uint32) (engineIdx int,
	engineUniqueIndex int, clusterID int, ok bool) {
	engineIdx, engineUniqueIndex, clusterID = -1, -1, -1
	reserved := (val >> 10)
	if reserved != 0 {
		_ = 101 // Does not really matter: 2021-12-29
	}
	lo, hi := int64(val&0x1f), int64(((val >> 5) & 0x1f))
	engineIdx, engineUniqueIndex, clusterID = md.decoder(lo, hi)
	ok = engineUniqueIndex >= 0
	return
}

type EngineTypeCode int

const (
	EngCat_SIP EngineTypeCode = iota + 100
	EngCat_SDMA
	EngCat_CDMA
	EngCat_CQM
	EngCat_GSYNC
	EngCat_SIP_LITE
	EngCat_SDMA_LITE
	EngCat_CDMA_LITE
	EngCat_PCIE
	EngCat_TS
	EngCat_ODMA
	EngCat_HCVG
	EngCat_VDEC
	EngCat_UNKNOWN
)

func EngineTypeCodeFor(walk func(typeCode EngineTypeCode)) {
	for _, v := range []EngineTypeCode{
		EngCat_TS,
		EngCat_PCIE,
		EngCat_GSYNC,
		EngCat_CQM,
		EngCat_CDMA,
		EngCat_SDMA,
		EngCat_SIP,
	} {
		walk(v)
	}
}

func ToEngineTypeCode(engingDesc string) EngineTypeCode {
	switch engingDesc {
	case ENGINE_SIP:
		return EngCat_SIP
	case ENGINE_SDMA:
		return EngCat_SDMA
	case ENGINE_CDMA:
		return EngCat_CDMA
	case ENGINE_CQM:
		return EngCat_CQM
	case ENGINE_GSYNC:
		return EngCat_GSYNC
	case ENGINE_SIP_LITE:
		return EngCat_SIP_LITE
	case ENGINE_SDMA_LITE:
		return EngCat_SDMA_LITE
	case ENGINE_CDMA_LITE:
		return EngCat_CDMA_LITE
	case ENGINE_PCIE:
		return EngCat_PCIE
	case ENGINE_TS:
		return EngCat_TS
	case ENGINE_ODMA:
		return EngCat_ODMA
	case ENGINE_HCVG:
		return EngCat_HCVG
	case ENGINE_VDEC:
		return EngCat_VDEC
	}
	return EngCat_UNKNOWN
}

func (e EngineTypeCode) String() string {
	switch e {
	case EngCat_SIP:
		return ENGINE_SIP
	case EngCat_SDMA:
		return ENGINE_SDMA
	case EngCat_CDMA:
		return ENGINE_CDMA
	case EngCat_CQM:
		return ENGINE_CQM
	case EngCat_GSYNC:
		return ENGINE_GSYNC
	case EngCat_SIP_LITE:
		return ENGINE_SIP_LITE
	case EngCat_SDMA_LITE:
		return ENGINE_SDMA_LITE
	case EngCat_CDMA_LITE:
		return ENGINE_CDMA_LITE
	case EngCat_PCIE:
		return ENGINE_PCIE
	case EngCat_TS:
		return ENGINE_TS
	case EngCat_ODMA:
		return ENGINE_ODMA
	case EngCat_HCVG:
		return ENGINE_HCVG
	case EngCat_VDEC:
		return ENGINE_VDEC
	}
	return ENGINE_UNKNOWN
}
