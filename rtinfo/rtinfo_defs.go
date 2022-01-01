package rtinfo

import (
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/meta"
)

type DpfAct struct {
	Start codec.DpfEvent
	End   codec.DpfEvent
}

type OpRef struct {
	dtuOp     *meta.DtuOp
	refToTask *RuntimeTask
}

type OpActivity struct {
	DpfAct
	opRef OpRef
}

func (q OpActivity) StartCycle() uint64 {
	return q.Start.Cycle
}

func (q OpActivity) EndCycle() uint64 {
	return q.End.Cycle
}

func (q OpActivity) Duration() int64 {
	return int64(q.EndCycle()) - int64(q.StartCycle())
}

func (q OpActivity) IsOpRefValid() bool {
	return q.opRef.dtuOp != nil && q.opRef.refToTask != nil
}

func (q OpActivity) GetOp() meta.DtuOp {
	return *q.opRef.dtuOp
}

func (q OpActivity) GetTask() RuntimeTask {
	return *q.opRef.refToTask
}
