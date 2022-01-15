package rtdata

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type DmaEventQueue struct {
	acts  DmaActivityVec
	eAlgo vgrule.ActMatchAlgo
}

func (q DmaEventQueue) DmaActivity() []DmaActivity {
	return q.acts
}

func (q DmaEventQueue) GetAlgo() vgrule.ActMatchAlgo {
	return q.eAlgo
}

func (q DmaEventQueue) DumpInfo() {
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

func (dmaVec *DmaEventQueue) AddAct(start, end codec.DpfEvent) {
	dmaVec.acts = append(dmaVec.acts, DmaActivity{
		DpfAct: DpfAct{
			start, end,
		},
	})
}
