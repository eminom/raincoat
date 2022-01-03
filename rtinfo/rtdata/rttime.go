package rtdata

import "fmt"

type DevCycleTime struct {
	DpfSyncIndex int // dpfSyncIndex is at most 31 bit-wide
	DevCycle     uint64
}

type HostTimeEntry struct {
	Cid          int
	Hosttime     uint64
	DpfSyncIndex int
}

type DevCycleAligned struct {
	DevCycleTime
	Cid      int
	Hosttime uint64
}

func (d DevCycleAligned) ToString() string {
	return fmt.Sprintf("%v %v %v", d.DpfSyncIndex, d.Hosttime, d.DevCycle)
}
