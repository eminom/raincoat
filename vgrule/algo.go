package vgrule

import "git.enflame.cn/hai.bai/dmaster/codec"

type EngineOrder interface {
	GetEngineOrder(dpf codec.DpfEvent) int
}

type ActMatchAlgo interface {
	EngineOrder
	GetChannelNum() int
	MapToChan(masterValue, ctx int) int
	DecodeChan(chNum int) (int, int)
}

type doradoRule struct{}

func NewDoradoRule() *doradoRule {
	return new(doradoRule)
}

func (a doradoRule) GetChannelNum() int {
	return codec.MASTERVALUE_COUNT * codec.RTCONTEXT_COUNT
}

func (a doradoRule) MapToChan(masterValue, ctx int) int {
	return masterValue<<codec.RTCONTEXT_BITCOUNT + ctx
}

func (a doradoRule) DecodeChan(chNum int) (int, int) {
	return chNum >> codec.RTCONTEXT_BITCOUNT,
		chNum & (1<<codec.RTCONTEXT_BITCOUNT - 1)
}

func (a doradoRule) GetEngineOrder(dpf codec.DpfEvent) int {
	return dpf.EngineIndex + 3*dpf.ClusterID
}
