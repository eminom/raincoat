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
}

func NewExecScope(execUuid uint64,
	pktIdToOp map[int]int, opMap map[int]DtuOp, dmaMap DmaInfoMap) *ExecScope {
	return &ExecScope{
		execUuid,
		pktIdToOp,
		opMap,
		dmaMap,
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

func (es ExecScope) DumpDtuopToOstream(fout *os.File) {
	var opIdVec []int
	for opId := range es.opMap {
		opIdVec = append(opIdVec, opId)
	}
	sort.Ints(opIdVec)
	for _, opId := range opIdVec {
		dtuOp := es.opMap[opId]
		fmt.Fprintf(fout, "%v %v kind=(%v), fusion_kind=(%v), mod_name=(%v)\n",
			opId, dtuOp.OpName,
			dtuOp.Kind,
			dtuOp.FusionKind,
			dtuOp.ModuleName,
		)
		fmt.Fprintf(fout, "##\n%v\n", dtuOp.MetaString)
	}
}

func (es ExecScope) DumpDtuOpToFile() {
	filename := es.getDumpFileName("opmeta")
	fout, err := os.Create(filename)
	if err != nil {
		panic(fmt.Errorf("could not open %v for: %v", filename, err))
	}
	defer fout.Close()
	es.DumpDtuopToOstream(fout)
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
