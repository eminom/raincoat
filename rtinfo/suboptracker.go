package rtinfo

import (
	"errors"
	"fmt"
	"math"
	"os"

	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

var (
	errNoPktEntryInfo = errors.New("no entry for pkt-id")
)

type SubOpTracker struct {
	taskIdToOpSeq   map[int][]rtdata.OpActivity
	taskToPidSubIdx map[int]*metadata.PacketIdInfoMap
	oneTaskFlag     bool
	errPidMet       map[int]bool
}

func NewSubOpTracker(
	taskIdToOpSeq map[int][]rtdata.OpActivity,
	taskToPidSubIdx map[int]*metadata.PacketIdInfoMap,
	oneTaskFlag bool) SubOpTracker {
	return SubOpTracker{
		taskIdToOpSeq:   taskIdToOpSeq,
		taskToPidSubIdx: taskToPidSubIdx,
		oneTaskFlag:     oneTaskFlag,
		errPidMet:       make(map[int]bool),
	}
}

// -1 for no op id
func (sot SubOpTracker) LocateOpId(taskId int,
	engineIndex int,
	startCycle, endCycle uint64) (opId int, subOpIndex int, subOpName string) {
	opId = -1
	subOpIndex = -1

	// Wrap engine index:
	if !sot.oneTaskFlag {
		engineIndex %= 4
	}

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

	var possibleErr error

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
			} else {
				possibleErr = errNoPktEntryInfo
			}
		}
		if elSubIndex >= 0 {
			// Update return values
			subOpIndex = elSubIndex
			subOpName = subName
		} else {
			sot.showDiagnosis(opMatched, possibleErr, maxFit)
		}
	}
	// fmt.Printf("max fit %v\n", maxFit)
	return
}

func (sot SubOpTracker) showDiagnosis(
	op rtdata.OpActivity,
	errPossible error,
	maxFit float64) {
	pid := op.Start.PacketID
	if sot.errPidMet[pid] {
		return
	}
	fmt.Fprintf(os.Stderr,
		"could not determine sub index: %v, pid(%v), engine_id(%v), max fit(%v)\n",
		errPossible,
		op.Start.PacketID,
		op.Start.EngineIndex,
		maxFit)

	sot.errPidMet[pid] = true
	taskId := op.GetTaskID()
	subMap, ok := sot.taskToPidSubIdx[taskId]
	if !ok {
		fmt.Fprintf(os.Stderr, "# no sub map for task id %v\n", taskId)
		return
	}

	_, ok = subMap.PktIdToSubIdx[pid]
	if !ok {
		fmt.Fprintf(os.Stderr, "# no sub index for packet id %v\n", pid)
		return
	}

	nameDc, ok := subMap.PktIdToName[pid]
	if !ok {
		fmt.Fprintf(os.Stderr, "# no sub name set for packet id %v\n", pid)
		return
	}

	_, ok = nameDc[op.Start.EngineIndex]
	if !ok {
		fmt.Fprintf(os.Stderr,
			"# no name info for engine %v, pkt id %v\n",
			op.Start.EngineIndex,
			pid,
		)
	}
}
