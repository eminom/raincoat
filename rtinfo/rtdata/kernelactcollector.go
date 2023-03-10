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
	kernelActLog = io.Discard
)

type KernelActCollector struct {
	acts KernelActivityVec
	DebugEventVec
	eAlgo vgrule.ActMatchAlgo
}

func NewKernelActCollector(algo vgrule.ActMatchAlgo) ActCollector {
	return &KernelActCollector{eAlgo: algo}
}

func (q KernelActCollector) GetActivity() interface{} {
	return q.acts
}

func (q KernelActCollector) GetAlgo() vgrule.ActMatchAlgo {
	return q.eAlgo
}

func (q KernelActCollector) DumpInfo() {
	fmt.Fprintf(kernelActLog, "%v Dma Acts found\n", len(q.acts))

	chDictInAll := make(map[int]int)
	for _, v := range q.acts {
		index := q.eAlgo.MapToChan(
			v.Start.MasterIdValue(),
			v.Start.Context,
		)
		chDictInAll[index]++
	}
}

func (q *KernelActCollector) AddAct(start, end codec.DpfEvent) {
	q.acts = append(q.acts, KernelActivity{
		DpfAct: DpfAct{
			start, end,
		},
	})
}

func (q KernelActCollector) ActCount() int {
	return len(q.acts)
}

func (q KernelActCollector) AxSelfClone() ActCollector {
	return &KernelActCollector{eAlgo: q.eAlgo}
}

func (q KernelActCollector) MergeInto(lhs ActCollector) {
	master := lhs.(*KernelActCollector)
	// dmaVec.DoSort()
	fmt.Fprintf(kernelActLog, "merge %v Dma Acts into master(currently %v)\n",
		len(q.acts), len(master.acts))
	master.acts = append(master.acts, q.acts...)
	master.debugEventVec = append(master.debugEventVec, q.debugEventVec...)
}

func (q KernelActCollector) DoSort() {
	startTs := time.Now()
	sort.Sort(q.acts)
	fmt.Fprintf(kernelActLog, "sort %v dma ops in %v\n", len(q.acts), time.Since(startTs))
}
