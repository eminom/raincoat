package rtinfo

import (
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type MatchPhysicalEngine interface {
	DoMatchTo(codec.DpfEvent, rtdata.OrderTask, vgrule.EngineOrder) bool
}

type MatchExtraConds func(uint64, int) bool

type MatchToSip struct{}

func (MatchToSip) DoMatchTo(
	evt codec.DpfEvent,
	task rtdata.OrderTask,
	rule vgrule.EngineOrder) bool {
	return task.AbleToMatchSip(evt, rule)
}

type MatchToCqm struct{}

func (MatchToCqm) DoMatchTo(
	evt codec.DpfEvent, task rtdata.OrderTask, rule vgrule.EngineOrder) bool {
	return task.AbleToMatchCqm(evt, rule)
}
