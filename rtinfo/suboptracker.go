package rtinfo

import (
	"fmt"
	"math"
	"os"

	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type SubOpTracker struct {
	taskIdToOpSeq   map[int][]rtdata.OpActivity
	taskToPidSubIdx map[int]map[int]int
}

func NewSubOpTracker(
	taskIdToOpSeq map[int][]rtdata.OpActivity,
	taskToPidSubIdx map[int]map[int]int) SubOpTracker {
	return SubOpTracker{
		taskIdToOpSeq:   taskIdToOpSeq,
		taskToPidSubIdx: taskToPidSubIdx,
	}
}

// -1 for no op id
func (sot SubOpTracker) LocateOpId(taskId int,
	startCycle, endCycle uint64) (opId int, subOpIndex int) {
	opId = -1
	subOpIndex = -1
	// Must be non-descending order to do binary search
	opSeq, taskValid := sot.taskIdToOpSeq[taskId]
	if !taskValid {
		return
	}
	// binary search for op id
	lo, hi := 0, len(opSeq)
	for lo < hi {
		md := (lo + hi) >> 1
		if startCycle <= opSeq[md].StartCycle() {
			hi = md
		} else {
			// srartCycle > [md]
			lo = md + 1
		}
	}
	//First one >= target(startCycle)
	// From the aspect of timeline view:
	//      Cqm Op Start           Cqm Op Start
	//                         Sip Start
	var maxFit float64 = -1.0
	for startIdx := lo; startIdx >= lo-1; startIdx-- {
		if startIdx >= 0 && startIdx < len(opSeq) {
			thisOpAct := opSeq[startIdx]
			leftCy := math.Max(float64(startCycle), float64(thisOpAct.StartCycle()))
			rightCy := math.Min(float64(endCycle), float64(thisOpAct.EndCycle()))
			r := (rightCy - leftCy) / float64(thisOpAct.Duration())
			if r > maxFit {
				maxFit = r
				opId = thisOpAct.GetOp().OpId

				// check sub idx
				thisTaskId := thisOpAct.GetTaskID()
				elSubIndex := -1
				if subMap, ok := sot.taskToPidSubIdx[thisTaskId]; ok {
					if index, ok := subMap[thisOpAct.Start.PacketID]; ok {
						elSubIndex = index
					}
				}
				if elSubIndex >= 0 {
					subOpIndex = elSubIndex
				} else {
					fmt.Fprintf(os.Stderr, "could not determine sub index")
				}
			}
		}
	}
	// fmt.Printf("max fit %v\n", maxFit)
	return
}
