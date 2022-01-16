package rtdata

import (
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
