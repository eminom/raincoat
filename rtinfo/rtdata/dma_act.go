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

type DmaActivityVec []DmaActivity
