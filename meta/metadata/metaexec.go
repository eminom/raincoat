package metadata

import (
	"errors"
	"fmt"
	"os"
	"sort"
)

var (
	ErrValidPacketIdNoOp = errors.New("known packet but no op")
	ErrInvalidPacketId   = errors.New("invalid packet id")
)

type ExecScope struct {
	execUuid  uint64
	pktIdToOp map[int]int
	opMap     map[int]DtuOp
	dmaMap    DmaInfoMap
	subOpMap  map[int][]SubOpMeta
}

func NewExecScope(execUuid uint64,
	pktIdToOp map[int]int, opMap map[int]DtuOp,
	dmaMap DmaInfoMap,
	subOpMap map[int][]SubOpMeta) *ExecScope {
	return &ExecScope{
		execUuid,
		pktIdToOp,
		opMap,
		dmaMap,
		subOpMap,
	}
}

func (es ExecScope) getDumpFileName(name string) string {
	const (
		dumpFileSuffix = "pbdumptxt"
	)
	hi32 := es.execUuid >> 32
	return fmt.Sprintf("0x%08x_%v.%v", hi32, name, dumpFileSuffix)
}

func (es ExecScope) DumpPktOpMapToOstream(fout *os.File) {
	var pktIdVec []int
	for pktId := range es.pktIdToOp {
		pktIdVec = append(pktIdVec, pktId)
	}
	sort.Ints(pktIdVec)
	for _, pktId := range pktIdVec {
		opId := es.pktIdToOp[pktId]
		fmt.Fprintf(fout, "%v %v\n", pktId, opId)
	}
}

func (es ExecScope) DumpPktOpMapToFile() {
	filename := es.getDumpFileName("pkt2op")
	fout, err := os.Create(filename)
	if err != nil {
		panic(fmt.Errorf("could not open %v for: %v", filename, err))
	}
	defer fout.Close()
	es.DumpPktOpMapToOstream(fout)
}

func (es ExecScope) DumpDtuopToOstream(fout *os.File, brief bool) {
	var opIdVec []int
	for opId := range es.opMap {
		opIdVec = append(opIdVec, opId)
	}
	sort.Ints(opIdVec)
	for _, opId := range opIdVec {
		dtuOp := es.opMap[opId]
		if brief {
			fmt.Fprintf(fout, "%v %v\n", opId, dtuOp.OpName)
		} else {
			fmt.Fprintf(fout, "%v %v kind=(%v), fusion_kind=(%v), mod_name=(%v)\n",
				opId, dtuOp.OpName,
				dtuOp.Kind,
				dtuOp.FusionKind,
				dtuOp.ModuleName,
			)
			fmt.Fprintf(fout, "##\n%v\n", dtuOp.MetaString)
		}
	}
}

func (es ExecScope) DumpSubOpMetaToOstream(fout *os.File) {
	var masterOpIdVec []int
	for opId := range es.subOpMap {
		masterOpIdVec = append(masterOpIdVec, opId)
	}
	sort.Ints(masterOpIdVec)
	for _, masterOpId := range masterOpIdVec {
		subVec := es.subOpMap[masterOpId]
		for _, subOp := range subVec {
			fmt.Fprintf(fout, "%v %v %v %v\n",
				subOp.MasterOpId,
				subOp.SlaveOpId,
				subOp.Tid,
				subOp.SubOpName,
			)
		}
	}
}

func (es ExecScope) DumpDtuOpToFile() {
	filename := es.getDumpFileName("opmeta")
	if fout, err := os.Create(filename); err == nil {
		defer fout.Close()
		es.DumpDtuopToOstream(fout, false)
	} else {
		panic(fmt.Errorf("could not open %v for: %v", filename, err))
	}
	filename = es.getDumpFileName("opmetaseq")
	if fout, err := os.Create(filename); err == nil {
		defer fout.Close()
		es.DumpDtuopToOstream(fout, true)
	} else {
		panic(fmt.Errorf("could not open %v for: %v", filename, err))
	}
}

func (es ExecScope) DumpSubOpToFile() {
	filename := es.getDumpFileName("subopmeta")
	if fout, err := os.Create(filename); err == nil {
		defer fout.Close()
		es.DumpSubOpMetaToOstream(fout)
	} else {
		panic(fmt.Errorf("could not open %v for: %v", filename, err))
	}
}

func (es ExecScope) DumpDmaToFile() {
	filename := es.getDumpFileName("memcpy")
	fout, err := os.Create(filename)
	if err != nil {
		panic(fmt.Errorf("could not open %v for: %v", filename, err))
	}
	defer fout.Close()
	es.dmaMap.DumpToOstream(fout)
}

func (es ExecScope) FindOp(packetId int) (DtuOp, error) {
	if opId, ok := es.pktIdToOp[packetId]; ok {
		if rv, ok1 := es.opMap[opId]; ok1 {
			return rv, nil
		}
		return DtuOp{}, ErrValidPacketIdNoOp
	}
	return DtuOp{}, ErrInvalidPacketId
}

func (es ExecScope) FindDma(packetId int) (DmaOp, error) {
	if dmaOp, ok := es.dmaMap.Info[packetId]; ok {
		return dmaOp, nil
	}
	return DmaOp{}, ErrInvalidPacketId
}

func (es ExecScope) CheckOpMapStatus(opMap map[int]bool) {
	matchedCount := 0
	for opId := range opMap {
		if _, ok := es.opMap[opId]; ok {
			matchedCount++
		}
	}
	fmt.Printf(
		"  all op(from pkt-mapped) %v, matched: %v out of %v\n",
		len(opMap),
		matchedCount,
		len(es.opMap),
	)
}

func (es ExecScope) MapPacketIdToOpId(packetId int) (int, bool) {
	if opId, ok := es.pktIdToOp[packetId]; ok {
		return opId, true
	}
	return 0, false
}

func (es ExecScope) IteratePacketId(cb func(pktId int)) {
	for pktId := range es.pktIdToOp {
		cb(pktId)
	}
}

func (es ExecScope) GetExecUuid() uint64 {
	return es.execUuid
}

func (es ExecScope) IteratePktToOp(cb func(int, int)) {
	for pkt, opId := range es.pktIdToOp {
		cb(pkt, opId)
	}
}

func (es ExecScope) CopyOpIdMap() map[int]string {
	dc := make(map[int]string)
	for opId, opInfo := range es.opMap {
		dc[opId] = opInfo.OpName
	}
	return dc
}

// OpId to Sub-sequencet
func (es ExecScope) GetSubOpIndexMap() map[int][]string {
	subOpIndexMap := make(map[int][]string)
	for opId, subOpVec := range es.subOpMap {
		visited := make(map[int]bool)
		for _, subOp := range subOpVec {
			if visited[subOp.SlaveOpId] {
				continue
			}
			visited[subOp.SlaveOpId] = true
			subOpIndexMap[opId] = append(subOpIndexMap[opId], subOp.SubOpName)
		}
	}
	return subOpIndexMap
}

func (es ExecScope) GetPacketToSubIdxMap() map[int]int {
	// op id to packet id seq
	opIdToPktSeq := make(map[int][]int)
	for pktId, opId := range es.pktIdToOp {
		opIdToPktSeq[opId] = append(opIdToPktSeq[opId], pktId)
	}

	pktIdToSubIdx := make(map[int]int)
	for _, pktIdSeq := range opIdToPktSeq {
		subIdx := 0
		for _, pktId := range pktIdSeq {
			// start op and end op are in pairs
			// But we dont assume that if start op packet id is odd(or event)
			pktIdToSubIdx[pktId] = subIdx / 2
			subIdx++
		}
	}
	return pktIdToSubIdx
}
