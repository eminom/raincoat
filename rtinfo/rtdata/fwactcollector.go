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
	fwActLog = io.Discard
)

type FwActCollector struct {
	acts FwActivityVec
	DebugEventVec
	algo vgrule.ActMatchAlgo
}

func NewFwActCollector(eAlgo vgrule.ActMatchAlgo) ActCollector {
	return &FwActCollector{algo: eAlgo}
}

func (fwa FwActCollector) GetAlgo() vgrule.ActMatchAlgo {
	return fwa.algo
}

func (fwa *FwActCollector) AddAct(start, end codec.DpfEvent) {
	fwa.acts = append(fwa.acts, FwActivity{DpfAct{start, end}})
}

func (fwa FwActCollector) DumpInfo() {

}
func (fwa FwActCollector) GetActivity() interface{} {
	return fwa.acts
}

func (fwa FwActCollector) ActCount() int {
	return len(fwa.acts)
}

func (fwa FwActCollector) AxSelfClone() ActCollector {
	return &FwActCollector{algo: fwa.algo}
}

func (fwa FwActCollector) MergeInto(lhs ActCollector) {
	master := lhs.(*FwActCollector)
	// fwa.DoSort()
	fmt.Fprintf(fwActLog, "merge %v Fw Acts into master(currently %v)\n",
		len(fwa.acts), len(master.acts))
	master.acts = append(master.acts, fwa.acts...)
	master.debugEventVec = append(master.debugEventVec, fwa.debugEventVec...)
}

func (fwa FwActCollector) DoSort() {
	// In-place sort works
	startTs := time.Now()
	sort.Sort(fwa.acts)
	fmt.Fprintf(fwActLog, "sort %v fw acts in %v\n", len(fwa.acts), time.Since(startTs))
}
