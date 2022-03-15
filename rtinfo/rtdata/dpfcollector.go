package rtdata

import "git.enflame.cn/hai.bai/dmaster/codec"

type DebugEventVec struct {
	debugEventVec []codec.DpfEvent
}

func (dev DebugEventVec) GetDebugEventVec() []codec.DpfEvent {
	return dev.debugEventVec
}

func (dev *DebugEventVec) AddDebugEvent(evt codec.DpfEvent) {
	dev.debugEventVec = append(dev.debugEventVec, evt)
}
