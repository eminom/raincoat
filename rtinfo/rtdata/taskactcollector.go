package rtdata

import (
	"fmt"
	"sort"
	"time"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type TaskActCollector struct {
	acts TaskActivityVec
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
	fmt.Printf("merge %v task acts into master(currently %v)\n",
		len(taskColl.acts), len(master.acts))
	master.acts = append(master.acts, taskColl.acts...)
}

func (taskColl TaskActCollector) DoSort() {
	// In-place sort works
	startTs := time.Now()
	sort.Sort(taskColl.acts)
	fmt.Printf("sort %v task acts in %v\n", len(taskColl.acts), time.Since(startTs))
}
