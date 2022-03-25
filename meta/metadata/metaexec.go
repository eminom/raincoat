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

type PacketIdInfoMap struct {
	PktIdToSubIdx map[int]int // Start and End packet id share the same sub index
	PktIdToName   map[int]string
}

func (es ExecScope) GetPacketToSubIdxMap() PacketIdInfoMap {
	// op id to packet id seq
	opIdToPktSeq := make(map[int][]int)
	for pktId, opId := range es.pktIdToOp {
		opIdToPktSeq[opId] = append(opIdToPktSeq[opId], pktId)
	}

	for opId := range opIdToPktSeq {
		sort.Ints(opIdToPktSeq[opId])
	}

	// Setup a map, op id to its name sequence
	// With validation
	opIdToSubOpNameSeq := make(map[int][]string)
	for opId, subSeq := range es.subOpMap {
		slaveOpNameMap := make(map[int]string)
		for _, subOp := range subSeq {
			if name, ok := slaveOpNameMap[subOp.SlaveOpId]; ok {
				if name != subOp.SubOpName {
					fmt.Fprintf(os.Stderr,
						"# ERROR: inconsistent sub op information: %s vs %s, op id(%v)\n",
						name, subOp.SubOpName,
						subOp.MasterOpId)
				}
			} else {
				slaveOpNameMap[subOp.SlaveOpId] = subOp.SubOpName
			}
		}
		var subIndexVec []int
		for subIdx := range slaveOpNameMap {
			subIndexVec = append(subIndexVec, subIdx)
		}
		sort.Ints(subIndexVec)
		var valid = false
		if len(subIndexVec) > 0 {
			valid = true
			if subIndexVec[0] != 0 {
				fmt.Fprintf(os.Stderr, "# ERROR: inconsistent sub op indexing start\n")
				valid = false
			}
			if subIndexVec[len(subIndexVec)-1] != len(subIndexVec)-1 {
				fmt.Fprintf(os.Stderr, "# ERROR: inconsistent sub op indexing end\n")
				valid = false
			}
		}

		if valid {
			for _, idx := range subIndexVec {
				opIdToSubOpNameSeq[opId] =
					append(opIdToSubOpNameSeq[opId],
						slaveOpNameMap[idx])
			}
		}
	}

	pktIdToSubIdx := make(map[int]int)
	pktIdToNameStr := make(map[int]string)
	// for packet id to name:  some entry may be missing if the verification fails

	for opId, pktIdSeq := range opIdToPktSeq {
		subIdx := 0
		for _, pktId := range pktIdSeq {
			// start op and end op are in pairs
			// But we dont assume that if start op packet id is odd(or event)
			pktIdToSubIdx[pktId] = subIdx / 2
			subIdx++
		}

		if subOpNameSeq, ok := opIdToSubOpNameSeq[opId]; ok &&
			len(subOpNameSeq) == subIdx/2 {
			for _, pktId := range pktIdSeq {
				// First map packet id to its sub indexing within the op scope
				subIdx := pktIdToSubIdx[pktId]
				pktIdToNameStr[pktId] = subOpNameSeq[subIdx]
			}
		} else {
			fmt.Fprintf(os.Stderr, "# error: No packet id to name map genereated\n")
		}

	}
	return PacketIdInfoMap{
		pktIdToSubIdx,
		pktIdToNameStr,
	}
}
