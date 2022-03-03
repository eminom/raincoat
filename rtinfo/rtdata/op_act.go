package rtdata

import (
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
)

type OpRef struct {
	dtuOp     *metadata.DtuOp
	refToTask *RuntimeTask
}

// Same task, same op
func (opRef OpRef) Eq(rhs OpRef) bool {
	return opRef.dtuOp != nil && rhs.dtuOp != nil &&
		opRef.dtuOp.OpId == rhs.dtuOp.OpId &&
		opRef.refToTask != nil && rhs.refToTask != nil &&
		opRef.refToTask.TaskID == rhs.refToTask.TaskID
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

func (opAct OpActivity) Eq(rhs OpActivity) bool {
	return opAct.opRef.Eq(rhs.opRef)
}

func (q *OpActivity) SetOpRef(opRef OpRef) {
	q.opRef = opRef
}

type OpActivityVector []OpActivity

func (opa OpActivityVector) Len() int {
	return len(opa)
}

func (opa OpActivityVector) Less(i, j int) bool {
	return opa[i].StartCycle() < opa[j].StartCycle()
}

func (opa OpActivityVector) Swap(i, j int) {
	opa[i], opa[j] = opa[j], opa[i]
}
