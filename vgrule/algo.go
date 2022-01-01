package vgrule

import "git.enflame.cn/hai.bai/dmaster/codec"

type ActMatchAlgo interface {
	GetChannelNum() int
	MapToChan(engineIdx, ctx int) int
	DecodeChan(chNum int) (int, int)
	GetEngineOrder(dpf codec.DpfEvent) int
}

type Algo1 struct{}

func NewAlgo1() *Algo1 {
	return new(Algo1)
}

func (a Algo1) GetChannelNum() int {
	return 1 << 8
}

func (a Algo1) MapToChan(engineIdx, ctx int) int {
	return engineIdx<<4 + ctx
}

func (a Algo1) DecodeChan(chNum int) (int, int) {
	return chNum >> 4, chNum & 0xF
}

func (a Algo1) GetEngineOrder(dpf codec.DpfEvent) int {
	return dpf.EngineIndex + 3*dpf.ClusterID
}
