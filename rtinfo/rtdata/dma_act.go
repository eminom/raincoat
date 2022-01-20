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

func (q DmaActivity) GetVcId() int {
	return (q.Start.Event >> 2) & (1<<6 - 1)
}

func (q DmaActivity) GetEngineIndex() int {
	return q.Start.EngineIndex
}

type DmaActivityVec []DmaActivity

func (d DmaActivityVec) Len() int {
	return len(d)
}

func (d DmaActivityVec) Less(i, j int) bool {
	return d[i].StartCycle() < d[j].StartCycle()
}

func (d DmaActivityVec) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
