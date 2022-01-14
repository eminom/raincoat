package vgrule

import "git.enflame.cn/hai.bai/dmaster/codec"

type EngineOrder interface {
	GetEngineOrder(dpf codec.DpfEvent) int
}

type MasterValueDecoder interface {
	DecodeMasterValue(int) (string, int, int)
}

type ActMatchAlgo interface {
	EngineOrder
	MasterValueDecoder
	GetChannelNum() int
	MapToChan(masterValue, ctx int) int
	DecodeChan(chNum int) (int, int)
}

type doradoRule struct {
	mDecoder MasterValueDecoder
}

func NewDoradoRule(decoder MasterValueDecoder) *doradoRule {
	return &doradoRule{mDecoder: decoder}
}

func (a doradoRule) DecodeMasterValue(val int) (string, int, int) {
	return a.mDecoder.DecodeMasterValue(val)
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
