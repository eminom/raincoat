package rtdata

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type OpActCollector struct {
	acts  OpActivityVector
	eAlgo vgrule.ActMatchAlgo
}

func NewOpActCollector(algo vgrule.ActMatchAlgo) *OpActCollector {
	return &OpActCollector{
		eAlgo: algo,
	}
}

func (q OpActCollector) GetActivity() interface{} {
	return q.acts
}

func (q OpActCollector) GetAlgo() vgrule.ActMatchAlgo {
	return q.eAlgo
}

func (q OpActCollector) DumpInfo() {
	fmt.Printf("%v acts found\n", len(q.acts))

	chDictInAll := make(map[int]int)
	for _, v := range q.acts {
		index := q.eAlgo.MapToChan(
			v.Start.MasterIdValue(),
			v.Start.Context,
		)
		chDictInAll[index]++
	}

	fmt.Printf("Cqm Op debug packet distribution:\n")
	for index, count := range chDictInAll {
		masterVal, ctx := q.eAlgo.DecodeChan(index)
		engTy, engIdx, clusterId := q.eAlgo.DecodeMasterValue(masterVal)
		fmt.Printf(" %v(%v) Cid(%v) ctx(%v) count: %v\n",
			engTy, engIdx, clusterId, ctx, count,
		)
	}
}

func (opVec *OpActCollector) AddAct(start, end codec.DpfEvent) {
	opVec.acts = append(opVec.acts, OpActivity{
		DpfAct: DpfAct{
			start, end,
		},
	})
}
