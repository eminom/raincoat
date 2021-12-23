package algo

type ActMatchAlgo interface {
	GetChannelNum() int
	MapToChan(engineIdx, ctx int) int
	DecodeChan(chNum int) (int, int)
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
