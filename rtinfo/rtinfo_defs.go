package rtinfo

import (
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/meta"
)

type DpfAct struct {
	Start codec.DpfEvent
	End   codec.DpfEvent
}

type CqmActBundle struct {
	DpfAct
	opRef *meta.DtuOp
}

func (q CqmActBundle) StartCycle() uint64 {
	return q.Start.Cycle
}

func (q CqmActBundle) EndCycle() uint64 {
	return q.End.Cycle
}

func (q CqmActBundle) Duration() int64 {
	return int64(q.EndCycle()) - int64(q.StartCycle())
}
