package meta

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

type metaInspector struct {
	er *ExecRaw
}

func NewMetaInspector(er *ExecRaw) metaInspector {
	return metaInspector{er: er}
}

func TestExecRaw(er *ExecRaw) {
	inspect := NewMetaInspector(er)
	inspect.DoBasicState()
}

func (m *metaInspector) DoBasicState() {
	allPacketIds := make(map[int]map[uint64]bool)
	for _, es := range m.er.bundle {
		es.IteratePktToOp(func(pktId, opId int) {
			if _, ok := allPacketIds[pktId]; !ok {
				allPacketIds[pktId] = make(map[uint64]bool)
			}
			allPacketIds[pktId][es.GetExecUuid()] = true
		})
	}
	fetchAny := func(dc map[uint64]bool) uint64 {
		for one := range dc {
			return one
		}
		assert.Assert(false, "must not be true")
		return 0
	}

	fmt.Printf("# DoBasicState\n")
	uniquePktDict := make(map[int]uint64)
	duplicatedPktDict := make(map[int]bool)
	for pkt, dc := range allPacketIds {
		if len(dc) == 1 {
			uniquePktDict[pkt] = fetchAny(dc)
		} else {
			duplicatedPktDict[pkt] = true
		}
	}

	fmt.Printf("  %v packets are unique across this session\n",
		len(uniquePktDict))
	fmt.Printf("  %v packets are not unique\n", len(duplicatedPktDict))
	fmt.Println()
}
