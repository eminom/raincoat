package rtinfo

import (
	"fmt"
	"os"
	"sort"

	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type KernelSubSeq struct {
	seq []*rtdata.KernelActivity
}

type TaskToEngines struct {
	traces map[int]*KernelSubSeq
}

func NewTaskToEngines() TaskToEngines {
	return TaskToEngines{
		traces: make(map[int]*KernelSubSeq),
	}
}

func (tte *TaskToEngines) AddTrace(
	kernelAct rtdata.KernelActivity) {
	masterId := kernelAct.Start.MasterIdValue()

	var trace *KernelSubSeq
	var ok bool
	if trace, ok = tte.traces[masterId]; !ok {
		trace = new(KernelSubSeq)
		tte.traces[masterId] = trace
	}
	// Use a copy
	cloned := kernelAct
	trace.seq = append(trace.seq, &cloned)
}

type SubOpState struct {
	states map[int]map[int]int
}

func NewSubOpState() SubOpState {
	return SubOpState{
		states: make(map[int]map[int]int),
	}
}

func (subS *SubOpState) AddTaskDict(taskId int) {
	if _, ok := subS.states[taskId]; !ok {
		subS.states[taskId] = make(map[int]int)
	}
}

func (subS *SubOpState) GetNextSubId(task int, masterId int, opId int) int {
	var targetState map[int]int
	var existed bool
	if targetState, existed = subS.states[task]; !existed {
		return -1
	}
	// sub index is distributed in per engine style
	// master value is 10-bit width
	key := masterId + (opId << 10)
	newSeqIdValue := targetState[key] // 0 for non-exist
	targetState[key]++
	return newSeqIdValue
}

// Return value is a copy
func GenerateKerenlActSeq(
	kernelActs []rtdata.KernelActivity,
	ops []rtdata.OpActivity,
) []rtdata.KernelActivity {

	// divided by task id, to engineval-opid s
	subOpState := NewSubOpState()
	chans := make(map[int][]rtdata.OpActivity)
	kernelTask := make(map[int]TaskToEngines)
	for _, op := range ops {
		taskId := op.GetTask().TaskID
		if _, ok := kernelTask[taskId]; !ok {
			kernelTask[taskId] = NewTaskToEngines()
		}
		subOpState.AddTaskDict(taskId)
		chans[taskId] = append(chans[taskId], op)
	}

	// distribute
	for _, kernelAct := range kernelActs {
		if kernelAct.RtInfo.TaskId > 0 {
			belongToTask := kernelAct.RtInfo.TaskId
			if tte, ok := kernelTask[belongToTask]; ok {
				tte.AddTrace(kernelAct)
			}
		}
	}

	var newSipBusies []rtdata.KernelActivity

	// distribution is done.
	// process task by task, engine streak by engine streak
	for taskId, tte := range kernelTask {
		opSeq := chans[taskId]
		// the value >= valueOf[idx]
		locateLowerBoundForOp := func(cycle uint64) int {
			lo, hi := 0, len(opSeq)
			for lo < hi {
				mid := (lo + hi) >> 1
				if cycle > opSeq[mid].StartCycle() {
					lo = mid + 1
				} else {
					hi = mid
				}
			}
			// lo if the lowerBound
			rvIdx := lo
			if rvIdx >= len(opSeq) {
				rvIdx--
			}
			if rvIdx >= 0 && rvIdx < len(opSeq) {
				if cycle < opSeq[rvIdx].StartCycle() {
					rvIdx--
				}
			}
			return rvIdx
		}
		for _, sipSeq := range tte.traces {
			for _, act := range sipSeq.seq {
				act.RtInfo.OpId = -1 // else
				idx := locateLowerBoundForOp(act.StartCycle())
				if idx < len(opSeq) && idx >= 0 {
					//
					dtuOpMeta := opSeq[idx].GetOp()
					subIdx := subOpState.GetNextSubId(taskId,
						act.Start.MasterIdValue(),
						dtuOpMeta.OpId)
					act.RtInfo.SubIdx = subIdx
					act.RtInfo.Name = dtuOpMeta.OpName
					act.RtInfo.OpId = dtuOpMeta.OpId

					// append to result collection
					newSipBusies = append(newSipBusies, *act)
				} else {
					fmt.Fprintf(os.Stderr, "Not found for %+v\n", act)
				}
			}
		}
	} // Assume that there is no missing
	sort.Sort(rtdata.KernelActivityVec(newSipBusies))
	return newSipBusies
}
