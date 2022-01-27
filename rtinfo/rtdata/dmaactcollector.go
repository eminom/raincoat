package rtdata

import (
	"fmt"
	"sort"
	"time"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type DmaActCollector struct {
	acts  DmaActivityVec
	eAlgo vgrule.ActMatchAlgo
}

func NewDmaCollector(algo vgrule.ActMatchAlgo) ActCollector {
	return &DmaActCollector{eAlgo: algo}
}

func (q DmaActCollector) GetActivity() interface{} {
	return q.acts
}

func (q DmaActCollector) GetAlgo() vgrule.ActMatchAlgo {
	return q.eAlgo
}

func (q DmaActCollector) DumpInfo() {
	fmt.Printf("%v Dma Acts found\n", len(q.acts))

	chDictInAll := make(map[int]int)
	for _, v := range q.acts {
		index := q.eAlgo.MapToChan(
			v.Start.MasterIdValue(),
			v.Start.Context,
		)
		chDictInAll[index]++
	}

	fmt.Printf("# Dma Op event distribution:\n")
	for index, count := range chDictInAll {
		masterVal, ctx := q.eAlgo.DecodeChan(index)
		engTy, engIdx, clusterId := q.eAlgo.DecodeMasterValue(masterVal)
		fmt.Printf("  %v(%v) Cid(%v) ctx(%v) count: %v\n",
			engTy, engIdx, clusterId, ctx, count,
		)
	}
}

func (dmaVec *DmaActCollector) AddAct(start, end codec.DpfEvent) {
	dmaVec.acts = append(dmaVec.acts, DmaActivity{
		DpfAct: DpfAct{
			start, end,
		},
	})
}

func (dmaVec DmaActCollector) ActCount() int {
	return len(dmaVec.acts)
}

func (dmaVec DmaActCollector) AxSelfClone() ActCollector {
	return &DmaActCollector{eAlgo: dmaVec.eAlgo}
}

func (dmaVec DmaActCollector) MergeInto(lhs ActCollector) {
	master := lhs.(*DmaActCollector)
	// dmaVec.DoSort()
	fmt.Printf("merge %v Dma Acts into master(currently %v)\n",
		len(dmaVec.acts), len(master.acts))
	master.acts = append(master.acts, dmaVec.acts...)
}

func (dmaVec DmaActCollector) DoSort() {
	startTs := time.Now()
	sort.Sort(dmaVec.acts)
	fmt.Printf("sort %v dma ops in %v\n", len(dmaVec.acts), time.Since(startTs))
}

func (dmaVec DmaActCollector) DoMergeOpAct() {}
