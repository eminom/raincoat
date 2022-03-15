package rtdata

import (
	"fmt"
	"sort"
	"time"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type OpActCollector struct {
	acts OpActivityVector
	DebugEventVec
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
	newAct := OpActivity{
		DpfAct: DpfAct{
			start, end,
		},
	}
	// Append directly
	opVec.acts = append(opVec.acts, newAct)
}

func (opVec OpActCollector) ActCount() int {
	return len(opVec.acts)
}

func (opVec OpActCollector) AxSelfClone() ActCollector {
	return &OpActCollector{eAlgo: opVec.eAlgo}
}

func (opVec OpActCollector) MergeInto(lhs ActCollector) {
	master := lhs.(*OpActCollector)
	// opVec.DoSort()
	fmt.Printf("merge %v OpActs into master(currently %v)\n",
		len(opVec.acts), len(master.acts),
	)
	master.acts = append(master.acts, opVec.acts...)
	master.debugEventVec = append(master.debugEventVec, opVec.debugEventVec...)
}

func (opVec OpActCollector) DoSort() {
	startTs := time.Now()
	sort.Sort(opVec.acts)
	fmt.Printf("sort %v dtuops in %v\n", len(opVec.acts), time.Since(startTs))
}

func maxU64(a, b uint64) uint64 {
	if a < b {
		return b
	}
	return a
}

func minU64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
