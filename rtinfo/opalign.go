package rtinfo

import (
	"sort"

	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

// Return value is a copy
func GenerateKerenlActSeq(
	kernelActs []rtdata.KernelActivity,
	sot SubOpTracker,
	subOpQuerier efintf.QuerySubOp,
) []rtdata.KernelActivity {
	// Sub ops
	var newSipBusies []rtdata.KernelActivity
	for _, kernelAct := range kernelActs {
		act := kernelAct
		if act.RtInfo.TaskId > 0 {
			tid := act.RtInfo.TaskId
			opId, subIdx := sot.LocateOpId(
				tid,
				act.StartCycle(),
				act.EndCycle(),
			)
			if opId >= 0 && subIdx >= 0 {
				realName, ok := subOpQuerier.QuerySubOpName(tid, opId, subIdx)
				if ok {
					act.RtInfo.Update(subIdx, realName, opId)
					newSipBusies = append(newSipBusies, act)
				}
			}
		}
	}
	sort.Sort(rtdata.KernelActivityVec(newSipBusies))
	return newSipBusies
}
