package rtinfo

import (
	"fmt"
	"os"
	"sort"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

func GenerateBriefOpsStat(execLocator func(uint64) metadata.ExecScope, bundle []rtdata.OpActivity) {
	opGathering := make(map[int][]rtdata.OpActivity)
	taskToExec := make(map[int]metadata.ExecScope)
	var taskVals []int
	for _, act := range bundle {
		assert.Assert(act.IsOpRefValid(), "must be valid for op generation")
		tid := act.GetTaskID()
		opGathering[tid] = append(opGathering[tid], act)
		execUuid := act.GetTask().ExecutableUUID
		taskToExec[tid] = execLocator(execUuid)
	}
	for tid := range opGathering {
		taskVals = append(taskVals, tid)
	}
	sort.Ints(taskVals)
	if fout, err := os.Create("tasklist.txt"); err == nil {
		defer fout.Close()
		for _, tid := range taskVals {
			fmt.Fprintf(fout, "Task %v\n", tid)

			opSeq := opGathering[tid]
			opIdMap := taskToExec[tid].CopyOpIdMap()
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
			if checkedCount == len(opIdMap) {
				fmt.Fprintf(fout, "  success: all %v op(s) checked\n", checkedCount)
			} else {
				fmt.Fprintf(fout, "  error: %v out of %v op(s) checked\n", checkedCount, len(opIdMap))
			}
			totalCy := endCy - startCy
			for _, op := range opSeq {
				fmt.Fprintf(fout, "  Op id: %v, %v\n",
					op.GetOp().OpId,
					op.GetOp().OpName)
				fmt.Fprintf(fout, "      Start at %v\n", op.Start.ToString())
				fmt.Fprintf(fout, "      End   at %v\n", op.End.ToString())
				var rate float64
				if totalCy > 0 {
					durCy := op.EndCycle() - op.StartCycle()
					rate = float64(durCy) / float64(totalCy)
					fmt.Fprintf(fout, "    %.4f%%\n", rate*100)
				}
			}
			fmt.Fprintf(fout, "\n")
		}
	}
}
