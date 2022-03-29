package rtinfo

import (
	"fmt"
	"math"
	"os"

	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type SubOpTracker struct {
	taskIdToOpSeq   map[int][]rtdata.OpActivity
	taskToPidSubIdx map[int]*metadata.PacketIdInfoMap
}

func NewSubOpTracker(
	taskIdToOpSeq map[int][]rtdata.OpActivity,
	taskToPidSubIdx map[int]*metadata.PacketIdInfoMap) SubOpTracker {
	return SubOpTracker{
		taskIdToOpSeq:   taskIdToOpSeq,
		taskToPidSubIdx: taskToPidSubIdx,
	}
}

// -1 for no op id
func (sot SubOpTracker) LocateOpId(taskId int,
	engineIndex int,
	startCycle, endCycle uint64) (opId int, subOpIndex int, subOpName string) {
	opId = -1
	subOpIndex = -1

	// Wrap engine index:
	engineIndex %= 4

	// Check for duration zero
	if endCycle <= startCycle {
		return
	}
	duration := float64(endCycle - startCycle)

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
	var opMatched rtdata.OpActivity
	for startIdx := lo; startIdx >= lo-1; startIdx-- {
		if startIdx >= 0 && startIdx < len(opSeq) {
			thisOpAct := opSeq[startIdx]
			leftCy := math.Max(float64(startCycle), float64(thisOpAct.StartCycle()))
			rightCy := math.Min(float64(endCycle), float64(thisOpAct.EndCycle()))
			r := (rightCy - leftCy) / duration
			if r > maxFit {
				maxFit = r
				opMatched = thisOpAct
			}
		}
	}

	if maxFit >= 0 {

		opId = opMatched.GetOp().OpId

		// check sub idx
		thisTaskId := opMatched.GetTaskID()
		elSubIndex := -1
		var subName string
		if subMap, ok := sot.taskToPidSubIdx[thisTaskId]; ok &&
			subMap != nil {
			if index, ok := subMap.PktIdToSubIdx[opMatched.Start.PacketID]; ok {

				subName = "" // update sub name along
				var subNameFound bool
				if subInfoDict, ok := subMap.PktIdToName[opMatched.Start.PacketID]; ok {

					var subIdxOk bool
					subIdx, subIdxOk := subMap.PktIdToSubIdx[opMatched.Start.PacketID]
					if !subIdxOk {
						subIdx = -1
					}

					subName, subNameFound = subInfoDict[engineIndex]
					if !subNameFound {
						fmt.Fprintf(os.Stderr, "# ERROR: no info for tid(%v), op_id(%v), sub(%v)\n",
							engineIndex,
							opId,
							subIdx,
						)
					}
				}

				if subNameFound {
					elSubIndex = index
				}
			}
		}
		if elSubIndex >= 0 {
			// Update return values
			subOpIndex = elSubIndex
			subOpName = subName
		} else {
			fmt.Fprintf(os.Stderr, "could not determine sub index\n")
		}
	}

	// fmt.Printf("max fit %v\n", maxFit)
	return
}
