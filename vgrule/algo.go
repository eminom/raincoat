package vgrule

import "git.enflame.cn/hai.bai/dmaster/codec"

type EngineOrder interface {
	GetEngineOrder(dpf codec.DpfEvent) int
}

type ActMatchAlgo interface {
	EngineOrder
	GetChannelNum() int
	MapToChan(engineIdx, ctx int) int
	DecodeChan(chNum int) (int, int)
}

type doradoRule struct{}

func NewDoradoRule() *doradoRule {
	return new(doradoRule)
}

func (a doradoRule) GetChannelNum() int {
	return 1 << 8
}

func (a doradoRule) MapToChan(engineIdx, ctx int) int {
	return engineIdx<<4 + ctx
}

func (a doradoRule) DecodeChan(chNum int) (int, int) {
	return chNum >> 4, chNum & 0xF
}

func (a doradoRule) GetEngineOrder(dpf codec.DpfEvent) int {
	return dpf.EngineIndex + 3*dpf.ClusterID
}
