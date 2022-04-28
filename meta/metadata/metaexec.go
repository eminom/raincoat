package metadata

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"git.enflame.cn/hai.bai/dmaster/efintf/efconst"
)

var (
	ErrValidPacketIdNoOp = errors.New("known packet but no op")
	ErrInvalidPacketId   = errors.New("invalid packet id")
)

type SuperExecScope interface {
	GetPacketToSubIdxMap() PacketIdInfoMap
	GetPacketToSubIdxMapV2() PacketIdInfoMap
}

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

	subInfo := es.GetPacketToSubIdxMap()

	for _, pktId := range pktIdVec {
		opId := es.pktIdToOp[pktId]

		if name, ok := subInfo.PktIdToName[pktId]; ok {
			fmt.Fprintf(fout, "%v %v %v\n", pktId, opId, name)
		} else {
			fmt.Fprintf(fout, "%v %v\n", pktId, opId)
		}
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
	typeSet := es.dmaMap.GetTypeSet()
	for dmaEngTy := range typeSet {
		filename := es.getDumpFileName(dmaEngTy + "_memcpy")
		fout, err := os.Create(filename)
		if err != nil {
			panic(fmt.Errorf("could not open %v for: %v", filename, err))
		}
		es.dmaMap.FilterByEngineType(dmaEngTy).DumpToOstream(fout)
		fout.Close()
	}
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
	PktIdToName   map[int]map[int]string
}

func (es ExecScope) GetOpToPktInfoCollection() (
	opIdToPktSeq map[int][]int,
	opIdToName map[int]string,
	opIdToSubOpNameSeq map[int][]map[int]string) {
	// op id to packet id seq
	opIdToPktSeq = make(map[int][]int)
	for pktId, opId := range es.pktIdToOp {
		opIdToPktSeq[opId] = append(opIdToPktSeq[opId], pktId)
	}

	opIdToName = make(map[int]string)
	for opId, opInfo := range es.opMap {
		opIdToName[opId] = opInfo.OpName
	}

	for opId := range opIdToPktSeq {
		sort.Ints(opIdToPktSeq[opId])
	}

	// Setup a map, op id to its name sequence
	// With validation
	opIdToSubOpNameSeq = make(map[int][]map[int]string)

	// sub index   0,                    1,                2
	// sub name    ---------------------------------------------------
	//               |
	//               |----> factor__0
	//               |----> factor__1
	//               |----> factor__2
	//               |----> factor__3

	for opId, subSeq := range es.subOpMap {
		// slaveOpNameMap: slave-index to thread-id to name
		slaveOpNameMap := make(map[int]map[int]string)
		getOrNewForSub := func(subOpId int) map[int]string {
			var dc map[int]string
			var ok bool
			dc, ok = slaveOpNameMap[subOpId]
			if !ok {
				dc = make(map[int]string)
				slaveOpNameMap[subOpId] = dc
			}
			return dc
		}
		for _, subOp := range subSeq {
			var subDict map[int]string
			var ok bool
			if subDict, ok = slaveOpNameMap[subOp.SlaveOpId]; ok {
				if _, ok = subDict[subOp.Tid]; ok {
					fmt.Fprintf(os.Stderr,
						"# ERROR: duplicate tid(%v) at op id(%v), sub index(%v)\n",
						subOp.Tid, subOp.MasterOpId, subOp.SlaveOpId)
				}
			}
			if !ok {
				getOrNewForSub(subOp.SlaveOpId)[subOp.Tid] = subOp.SubOpName
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
				// idx is the sub index
				// Now we need to correlct tid to sub name
				// (for there are different sub op name for different tid)
				opIdToSubOpNameSeq[opId] =
					append(opIdToSubOpNameSeq[opId],
						slaveOpNameMap[idx])
			}
		}
	}
	return
}

func SynsthesisPktInfoDict(execUuid uint64,
	opIdToPktSeq map[int][]int,
	opNameDict map[int]string,
	opIdToSubOpNameSeq map[int][]map[int]string) PacketIdInfoMap {
	// The return values of this function
	pktIdToSubIdx := make(map[int]int)
	pktIdToNameStr := make(map[int]map[int]string)
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
			if !efconst.IsInvalidOpId(opId) {
				fmt.Fprintf(os.Stderr,
					"# error: (%v)No packet id to name map genereated for op id(%v %v)\n",
					fmt.Sprintf("0x%016x", execUuid)[:10],
					opId,
					opNameDict[opId])
			}
		}

	}
	return PacketIdInfoMap{
		pktIdToSubIdx,
		pktIdToNameStr,
	}
}

func (es ExecScope) GetPacketToSubIdxMap() PacketIdInfoMap {
	opIdToPktSeq, opNameDict, opIdToSubOpNameSeq := es.GetOpToPktInfoCollection()
	return SynsthesisPktInfoDict(es.GetExecUuid(),
		opIdToPktSeq, opNameDict, opIdToSubOpNameSeq)
}

func (es ExecScope) GetPacketToSubIdxMapV2() PacketIdInfoMap {
	return es.GetPacketToSubIdxMap()
}
