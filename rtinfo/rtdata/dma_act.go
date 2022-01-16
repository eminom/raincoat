package rtdata

import "git.enflame.cn/hai.bai/dmaster/meta/metadata"

type DmaMetaRef struct {
	dmaOp     *metadata.DmaOp
	refToTask *RuntimeTask
}

func NewDmaRef(dmaOp *metadata.DmaOp, refToTask *RuntimeTask) DmaMetaRef {
	return DmaMetaRef{
		dmaOp,
		refToTask,
	}
}

type DmaActivity struct {
	DpfAct
	dmaRef DmaMetaRef
}

func (q DmaActivity) IsDmaMetaRefValid() bool {
	return q.dmaRef.dmaOp != nil && q.dmaRef.refToTask != nil
}

func (q DmaActivity) GetDmaMeta() metadata.DmaOp {
	return *q.dmaRef.dmaOp
}

func (q DmaActivity) GetTask() RuntimeTask {
	return *q.dmaRef.refToTask
}

func (q *DmaActivity) SetDmaRef(dmaRef DmaMetaRef) {
	q.dmaRef = dmaRef
}

type DmaActivityVec []DmaActivity
