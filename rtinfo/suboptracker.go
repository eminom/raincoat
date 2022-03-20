package rtinfo

import (
	"math"

	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type SubOpTracker struct {
	taskIdToOpSeq map[int][]rtdata.OpActivity
}

func NewSubOpTracker(
	taskIdToOpSeq map[int][]rtdata.OpActivity) SubOpTracker {
	return SubOpTracker{
		taskIdToOpSeq: taskIdToOpSeq,
	}
}

// -1 for no op id
func (sot SubOpTracker) LocateOpId(taskId int, startCycle, endCycle uint64) int {
	// Must be non-descending order to do binary search
	opSeq, taskValid := sot.taskIdToOpSeq[taskId]
	if !taskValid {
		return -1
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
	var opId = -1
	for startIdx := lo; startIdx >= lo-1; startIdx-- {
		if startIdx >= 0 && startIdx < len(opSeq) {
			thisOpAct := opSeq[startIdx]
			leftCy := math.Max(float64(startCycle), float64(thisOpAct.StartCycle()))
			rightCy := math.Min(float64(endCycle), float64(thisOpAct.EndCycle()))
			r := (rightCy - leftCy) / float64(thisOpAct.Duration())
			if r > maxFit {
				maxFit = r
				opId = thisOpAct.GetOp().OpId
			}
		}
	}
	// fmt.Printf("max fit %v\n", maxFit)
	return opId
}
