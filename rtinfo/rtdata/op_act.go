package rtdata

import (
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
)

type OpRef struct {
	dtuOp     *metadata.DtuOp
	refToTask *RuntimeTask
}

func NewOpRef(dtuOp *metadata.DtuOp, refToTask *RuntimeTask) OpRef {
	return OpRef{
		dtuOp,
		refToTask,
	}
}

type OpActivity struct {
	DpfAct
	opRef OpRef
}

func (q OpActivity) IsOpRefValid() bool {
	return q.opRef.dtuOp != nil && q.opRef.refToTask != nil
}

func (q OpActivity) GetOp() metadata.DtuOp {
	return *q.opRef.dtuOp
}

func (q OpActivity) GetTask() RuntimeTask {
	return *q.opRef.refToTask
}

func (q *OpActivity) SetOpRef(opRef OpRef) {
	q.opRef = opRef
}

type OpActivityVector []OpActivity
