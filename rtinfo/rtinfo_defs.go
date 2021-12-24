package rtinfo

import (
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/meta"
)

type DpfAct struct {
	Start codec.DpfEvent
	End   codec.DpfEvent
}

type DpfActBundle struct {
	act   DpfAct
	opRef *meta.DtuOp
}
