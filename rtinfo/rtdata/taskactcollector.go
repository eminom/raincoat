package rtdata

import (
	"fmt"
	"io"
	"sort"
	"time"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

var (
	taskActLog = io.Discard
)

type TaskActCollector struct {
	acts TaskActivityVec
	DebugEventVec
	algo vgrule.ActMatchAlgo
}

func NewTaskActCollector(eAlgo vgrule.ActMatchAlgo) ActCollector {
	return &TaskActCollector{algo: eAlgo}
}

func (taskColl TaskActCollector) GetAlgo() vgrule.ActMatchAlgo {
	return taskColl.algo
}

func (taskColl *TaskActCollector) AddAct(start, end codec.DpfEvent) {
	taskColl.acts = append(taskColl.acts, TaskActivity{DpfAct{start, end}})
}

func (taskColl TaskActCollector) DumpInfo() {
	//TODO Dump exec information
}

func (taskColl TaskActCollector) GetActivity() interface{} {
	return taskColl.acts
}

func (taskColl TaskActCollector) ActCount() int {
	return len(taskColl.acts)
}

func (taskColl TaskActCollector) AxSelfClone() ActCollector {
	return &TaskActCollector{algo: taskColl.algo}
}

func (taskColl TaskActCollector) MergeInto(lhs ActCollector) {
	master := lhs.(*TaskActCollector)
	// taskColl.DoSort()
	fmt.Fprintf(taskActLog, "merge %v task acts into master(currently %v)\n",
		len(taskColl.acts), len(master.acts))
	master.acts = append(master.acts, taskColl.acts...)
	master.debugEventVec = append(master.debugEventVec, taskColl.debugEventVec...)
}

func (taskColl TaskActCollector) DoSort() {
	// In-place sort works
	startTs := time.Now()
	sort.Sort(taskColl.acts)
	fmt.Fprintf(taskActLog, "sort %v task acts in %v\n", len(taskColl.acts), time.Since(startTs))
}
