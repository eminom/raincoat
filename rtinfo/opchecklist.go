package rtinfo

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"sort"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

/*
 * Sort all Op activities consisting of CQM debug packet
 * by TaskID
 */

type TaskInfoState struct {
	PgMask           int
	TotalCycles      uint64
	ExecScope        metadata.ExecScope
	OpGathering      []rtdata.OpActivity
	BoundedByExecAct bool
	HasExecAct       bool
	OffsetState      string
}

func (t *TaskInfoState) CollectOp(op rtdata.OpActivity) {
	t.OpGathering = append(t.OpGathering, op)
}

func (t TaskInfoState) GetExecUuid() uint64 {
	return t.ExecScope.GetExecUuid()
}

type ExecInfoState struct {
	TaskIdVec           []int
	CyclesInAll         uint64
	CyclesInAllForBound uint64
	BoundCount          int // Number of exec act bound task
}

func (e *ExecInfoState) CollectTaskInfo(
	taskId int, cycleCost uint64, bound bool) {
	e.TaskIdVec = append(e.TaskIdVec, taskId)
	e.CyclesInAll += cycleCost
	if bound {
		e.CyclesInAllForBound += cycleCost
		e.BoundCount++
	}
}

func (e ExecInfoState) AverageCycles() float64 {
	if e.BoundCount > 0 {
		return float64(e.CyclesInAllForBound) / float64(e.BoundCount)
	}
	return 1
}

func GenerateBriefOpsStat(
	execLocator func(uint64) metadata.ExecScope,
	bundle []rtdata.OpActivity,
	executableMap map[int]rtdata.FwActivity,
	orderTaskVec []rtdata.OrderTask,
	runtimeTaskVec []rtdata.RuntimeTaskBase,
) {

	var taskVals []int
	taskVisited := make(map[int]bool)
	execVisited := make(map[uint64]bool)
	statByTask := make(map[int]*TaskInfoState)
	statByExec := make(map[uint64]*ExecInfoState)
	for _, act := range bundle {
		tid := act.GetTaskID()
		execUuid := act.GetTask().ExecutableUUID

		if !taskVisited[tid] {
			taskVisited[tid] = true
			taskVals = append(taskVals, tid)
			statByTask[tid] = &TaskInfoState{
				PgMask:    act.GetTask().PgMask,
				ExecScope: execLocator(execUuid),
			}
		}

		if !execVisited[execUuid] {
			execVisited[execUuid] = true
			statByExec[execUuid] = &ExecInfoState{}
		}

		// Always collect
		statByTask[tid].CollectOp(act)
	}
	sort.Ints(taskVals)

	// task id to brief report
	taskReportMap := make(map[int]string)
	for _, tid := range taskVals {
		reportStream := bytes.NewBuffer(nil)
		execScope := statByTask[tid].ExecScope
		thisExecuuid := execScope.GetExecUuid()
		opSeq := statByTask[tid].OpGathering
		opIdToStrMap := execScope.CopyOpIdMap()

		// Clone a bool map for visiting mark
		opIdMap := make(map[int]bool)
		for opId := range opIdToStrMap {
			opIdMap[opId] = false
		}

		var startCy, endCy uint64
		if len(opSeq) > 0 {
			startCy = opSeq[0].StartCycle()
		}
		checkedCount := 0
		for _, op := range opSeq {
			if startCy > op.StartCycle() {
				startCy = op.StartCycle()
			}
			if endCy < op.EndCycle() {
				endCy = op.EndCycle()
			}
			opId := op.GetOp().OpId
			if checked, ok := opIdMap[opId]; ok {
				if !checked {
					opIdMap[opId] = true
					checkedCount++
				}
			} else {
				assert.Assert(false, "please panic")
			}
		}

		var gapsCycle uint64
		var hasExecutableAct = false
		isBound := false
		if fwAct, ok := executableMap[tid]; ok && fwAct.End.Event >= 0 {
			hasExecutableAct = true
			leftBound, rightBound := false, false
			if startCy >= fwAct.StartCycle() {
				startCy = fwAct.StartCycle()
				leftBound = true
			}
			if endCy <= fwAct.EndCycle() {
				endCy = fwAct.EndCycle()
				rightBound = true
			}
			isBound = leftBound && rightBound
		}

		hintStr := "error  "
		if isBound {
			hintStr = "success"
		}

		if checkedCount == len(opIdMap) {
			fmt.Fprintf(reportStream,
				"  %v: all %v op(s) checked\n", hintStr, checkedCount)
		} else {
			fmt.Fprintf(reportStream,
				"  %v: %v out of %v op(s) checked\n",
				hintStr, checkedCount, len(opIdMap))
		}

		// Let's assume that op id is in order
		var opIdSeq []int
		for opId := range opIdToStrMap {
			opIdSeq = append(opIdSeq, opId)
		}
		sort.Ints(opIdSeq)

		// Op in details
		strStream := bytes.NewBuffer(nil)
		previousCy := startCy
		totalCy := endCy - startCy

		// indexing op sequence
		opSeqIdx := 0
		for _, op := range opSeq {
			curOpId := op.GetOp().OpId

			// Check missing ops
			for opSeqIdx < len(opIdSeq) && opIdSeq[opSeqIdx] != curOpId {
				fmt.Fprintf(strStream,
					"  Op id: %v, %v, missing\n",
					opIdSeq[opSeqIdx],
					opIdToStrMap[opIdSeq[opSeqIdx]],
				)
				opSeqIdx++
			}
			if opSeqIdx < len(opIdSeq) {
				opSeqIdx++
			}

			if previousCy < op.StartCycle() {
				gapsCycle += op.StartCycle() - previousCy
			}
			var rate float64
			if totalCy > 0 {
				durCy := op.EndCycle() - op.StartCycle()
				rate = float64(durCy) / float64(totalCy)
			}
			fmt.Fprintf(strStream, "  Op id: %v, %v, %.4f%%\n",
				curOpId,
				op.GetOp().OpName,
				rate*100)
			fmt.Fprintf(strStream, "      Start at %v\n", op.Start.ToString())
			fmt.Fprintf(strStream, "      End   at %v\n", op.End.ToString())

			// update previousCy
			previousCy = op.EndCycle()
		}
		if previousCy < endCy {
			gapsCycle += endCy - previousCy
		}
		gapRate := float64(gapsCycle) / float64(totalCy)
		if hasExecutableAct {
			fmt.Fprintf(reportStream,
				"  total cycles: %v, gap: %.2f%%\n",
				totalCy,
				gapRate*100)
		} else {
			fmt.Fprintf(reportStream,
				"  no executable activity\n")
		}
		io.WriteString(reportStream, strStream.String())
		fmt.Fprintf(reportStream, "\n")
		// Cache to map
		taskReportMap[tid] = reportStream.String()
		statByTask[tid].TotalCycles = totalCy
		statByTask[tid].BoundedByExecAct = isBound
		statByTask[tid].HasExecAct = hasExecutableAct

		statByExec[thisExecuuid].CollectTaskInfo(tid, totalCy, isBound)
	}

	// Calculate overall info
	for _, tid := range taskVals {
		thisExeScope := statByTask[tid].ExecScope
		thisExecuuid := thisExeScope.GetExecUuid()

		average := statByExec[thisExecuuid].AverageCycles()
		sign := "+"
		thisCycleCost := statByTask[tid].TotalCycles
		if float64(thisCycleCost) <= average {
			sign = "-"
		}
		offset :=
			math.Abs(float64(thisCycleCost)-average) / float64(average)
		offsetState := fmt.Sprintf("%v %.2f%%", sign, offset*100)
		if statByTask[tid].BoundedByExecAct {
			statByTask[tid].OffsetState = offsetState // cache offset info
		} else {
			statByTask[tid].OffsetState = "N/A"
		}
	}

	// Summary
	if fout, err := os.Create("tasklist.txt"); err == nil {
		defer fout.Close()
		dumpStatInfoToFile(fout,
			taskVals, statByTask, statByExec,
			taskReportMap,
			orderTaskVec,
			runtimeTaskVec,
		)
	}
}

func dumpStatInfoToFile(fout *os.File,
	taskVals []int,
	statByTask map[int]*TaskInfoState,
	statByExec map[uint64]*ExecInfoState,
	taskReport map[int]string,
	orderTaskVec []rtdata.OrderTask,
	runtimeTaskVec []rtdata.RuntimeTaskBase,
) {
	// Task by task
	for _, tid := range taskVals {
		thisExeScope := statByTask[tid].ExecScope
		thisExecuuid := thisExeScope.GetExecUuid()
		fmt.Fprintf(fout, "Task %v, Exec 0x%16x, PgMask: %v, Exec act bound: %v\n",
			tid,
			thisExecuuid,
			statByTask[tid].PgMask,
			statByTask[tid].BoundedByExecAct,
		)
		fmt.Fprintf(fout, "  cycle to average [%v] (%v/%v)\n",
			statByTask[tid].OffsetState,
			statByTask[tid].TotalCycles,
			statByExec[thisExecuuid].AverageCycles(),
		)
		io.WriteString(fout, taskReport[tid])
	}

	// Dump stat sorted by Executable
	fmt.Fprintf(fout, "\n")
	for exec := range statByExec {
		fmt.Fprintf(fout, "Exec 0x%16x, task list:\n", exec)
		for _, tid := range statByExec[exec].TaskIdVec {
			statInfo := statByTask[tid]
			fmt.Fprintf(fout, "  %v, pg %v, cycles: %v, exec bound: %v, with exec act: %v, execution time: %v\n",
				tid,
				statInfo.PgMask,
				statInfo.TotalCycles,
				statInfo.BoundedByExecAct,
				statInfo.HasExecAct,
				statInfo.OffsetState,
			)
		}
		fmt.Fprintf(fout, "\n")
	}

	// Dump task overview
	fmt.Fprintf(fout, "\n")
	taskVisited := make(map[int]bool)
	for _, task := range orderTaskVec {
		taskVisited[task.GetTaskID()] = true
		fmt.Fprintf(fout, "Task %-4d, %22d, 0x%16x, %v\n",
			task.GetTaskID(),
			task.StartCy,
			task.GetRefToTask().ExecutableUUID,
			task.GetRefToTask().PgMask)
	}

	fmt.Fprintf(fout, "\n")
	for _, task := range runtimeTaskVec {
		if !taskVisited[task.TaskID] {
			fmt.Fprintf(fout, "Task %-4d, 0x%16x, %v, missing\n",
				task.TaskID,
				task.ExecutableUUID,
				task.PgMask)
		}
	}

	//
	fmt.Fprintf(fout, "\n")
}
