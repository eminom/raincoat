package rtdata

import (
	"fmt"
	"sort"
	"time"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type FwActCollector struct {
	acts FwActivityVec
	algo vgrule.ActMatchAlgo
}

func NewFwActCollector(eAlgo vgrule.ActMatchAlgo) ActCollector {
	return &FwActCollector{algo: eAlgo}
}

func (q FwActCollector) GetAlgo() vgrule.ActMatchAlgo {
	return q.algo
}

func (q *FwActCollector) AddAct(start, end codec.DpfEvent) {
	q.acts = append(q.acts, FwActivity{DpfAct{start, end}})
}

func (q FwActCollector) DumpInfo() {

}
func (q FwActCollector) GetActivity() interface{} {
	return q.acts
}

func (q FwActCollector) ActCount() int {
	return len(q.acts)
}

func (dmaVec FwActCollector) AxSelfClone() ActCollector {
	return &FwActCollector{algo: dmaVec.algo}
}

func (q FwActCollector) MergeInto(lhs ActCollector) {
	master := lhs.(*FwActCollector)
	q.DoSort()
	fmt.Printf("merge %v Fw Acts into master(currently %v)\n",
		len(q.acts), len(master.acts))
	master.acts = append(master.acts, q.acts...)
}

func (q FwActCollector) DoSort() {
	// In-place sort works
	startTs := time.Now()
	sort.Sort(q.acts)
	fmt.Printf("sort %v fw acts in %v\n", len(q.acts), time.Since(startTs))
}
