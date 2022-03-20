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
	// task-id to op-id to entity-index to sub-op seq
	collections := make(map[int]map[int]map[int][]rtdata.KernelActivity)
	addKernelAct := func(taskId int, opId int, kernelAct rtdata.KernelActivity) {
		var taskScope map[int]map[int][]rtdata.KernelActivity
		var ok bool
		taskScope, ok = collections[taskId]
		if !ok {
			taskScope = make(map[int]map[int][]rtdata.KernelActivity)
			collections[taskId] = taskScope
		}
		var opScope map[int][]rtdata.KernelActivity
		opScope, ok = taskScope[opId]
		if !ok {
			opScope = make(map[int][]rtdata.KernelActivity)
			taskScope[opId] = opScope
		}
		entityIdx := kernelAct.GetEngineIndex()
		opScope[entityIdx] = append(opScope[entityIdx], kernelAct)
	}

	// distribute
	for _, kernelAct := range kernelActs {
		if kernelAct.RtInfo.TaskId > 0 {
			tid := kernelAct.RtInfo.TaskId
			opId := sot.LocateOpId(
				tid,
				kernelAct.StartCycle(),
				kernelAct.EndCycle(),
			)
			if opId >= 0 {
				addKernelAct(tid, opId, kernelAct)
			}
		}
	}

	// Sub ops
	var newSipBusies []rtdata.KernelActivity
	for taskId, taskScope := range collections {
		for opId, opScope := range taskScope {
			for entityIdx, entityScopeSeq := range opScope {
				for subIdx, act := range entityScopeSeq {
					realName, ok := subOpQuerier.QuerySubOpName(
						taskId, opId, entityIdx, subIdx)
					if ok {
						act.RtInfo.Update(subIdx, realName, opId)
						newSipBusies = append(newSipBusies, act)
					}
				}
			}
		}
	}

	sort.Sort(rtdata.KernelActivityVec(newSipBusies))
	return newSipBusies
}
