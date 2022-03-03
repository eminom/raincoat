package rtdata

import (
	"git.enflame.cn/hai.bai/dmaster/codec"
)

type DpfAct struct {
	Start codec.DpfEvent
	End   codec.DpfEvent
}

func (q DpfAct) StartCycle() uint64 {
	return q.Start.Cycle
}

func (q DpfAct) EndCycle() uint64 {
	return q.End.Cycle
}

func (q DpfAct) Duration() int64 {
	return int64(q.EndCycle()) - int64(q.StartCycle())
}

func (q DpfAct) IsOfEngine(engineTypeCode codec.EngineTypeCode) bool {
	return q.Start.EngineTypeCode == engineTypeCode
}

// Combine,
// the earlier start
// the latter end
func (dpfAct *DpfAct) CombineCycle(rhs DpfAct) {
	if dpfAct.Start.Cycle > rhs.Start.Cycle {
		dpfAct.Start.Cycle = rhs.Start.Cycle
	}
	if dpfAct.End.Cycle < rhs.End.Cycle {
		dpfAct.End.Cycle = rhs.End.Cycle
	}
}
