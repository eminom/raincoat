package rtdata

import (
	"crypto/md5"
	"encoding/binary"

	"git.enflame.cn/hai.bai/dmaster/codec"
)

type DpfAct struct {
	Start codec.DpfEvent
	End   codec.DpfEvent
}

func (q DpfAct) StartCycle() uint64 {
	return q.Start.Cycle
}

func (q DpfAct) EndCycle() uint64 {
	return q.End.Cycle
}

func (q DpfAct) ContextId() int {
	return q.Start.Context
}

func (q DpfAct) Duration() int64 {
	return int64(q.EndCycle()) - int64(q.StartCycle())
}

func (q DpfAct) GetHashCode() uint64 {
	buf := make([]byte, 32)
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint32(buf[i*4:], q.Start.RawValue[i])
		binary.LittleEndian.PutUint32(buf[i*4+8:], q.End.RawValue[i])
	}
	hash := md5.New()
	hash.Write(buf)
	res := hash.Sum(nil)
	return binary.LittleEndian.Uint64(res)
}

// Combine,
// the earlier start
// the latter end
func (dpfAct *DpfAct) CombineCycle(rhs DpfAct) {
	if dpfAct.Start.Cycle > rhs.Start.Cycle {
		dpfAct.Start.Cycle = rhs.Start.Cycle
	}
	if dpfAct.End.Cycle < rhs.End.Cycle {
		dpfAct.End.Cycle = rhs.End.Cycle
	}
}
