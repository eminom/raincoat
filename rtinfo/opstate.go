package rtinfo

import (
	"fmt"
	"sort"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type OpXState struct {
	opLane    map[int][]rtdata.OpActivity
	converged []rtdata.OpActivity
}

func NewOpXState() *OpXState {
	return &OpXState{
		opLane: make(map[int][]rtdata.OpActivity),
	}
}

func (xs *OpXState) AddOp(op rtdata.OpActivity) {
	// And it is guaratee that op carries meta info
	chanIdx := mapTaskIdContextId(op.GetTask().TaskID, op.ContextId())
	if arr, ok := xs.opLane[chanIdx]; ok && arr[len(arr)-1].Eq(op) {
		arr[len(arr)-1].DpfAct.CombineCycle(op.DpfAct)
	} else {
		// The first one.
		xs.opLane[chanIdx] = append(xs.opLane[chanIdx], op)
	}
}

func (xs *OpXState) FinalizeOps() []rtdata.OpActivity {
	if len(xs.opLane) > 0 {
		idxSet := make(map[int]bool)
		for idVal := range xs.opLane {
			idxSet[idVal] = true
		}
		for idVal := range idxSet {
			xs.collectAtChannel(idVal)
		}
	}
	assert.Assert(len(xs.opLane) == 0, "must be purged")
	sort.Sort(rtdata.OpActivityVector(xs.converged))
	fmt.Printf("# %v Dtu Op(s) have been collected\n", len(xs.converged))
	return xs.converged
}

func (xs *OpXState) CombineOps(taskId int, ctxId int) {
	xs.collectAtChannel(mapTaskIdContextId(taskId, ctxId))
}

// If step ends is reached
func (xs *OpXState) collectAtChannel(idVal int) {
	if ops, ok := xs.opLane[idVal]; ok {
		sort.Sort(rtdata.OpActivityVector(ops))
		xs.converged = append(xs.converged, ops...)
		delete(xs.opLane, idVal) // purge
	}
}

func mapTaskIdContextId(taskId, contextId int) int {
	return (taskId << 4) + contextId
}
