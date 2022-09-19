package sess

import (
	"bytes"
	"fmt"
)

// 4 events
type DteEventStat struct {
	Channel [64]int
	Total   int
}

func (de *DteEventStat) TickChannel(ch int) {
	de.Channel[ch]++
	de.Total++
}

func (de DteEventStat) ToString() string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "%v", de.Total)
	return buf.String()
}

// 1. in event classification
// 2. in vc classification
type DteStat struct {
	EvtStat  [4]DteEventStat
	OtherCnt int
	TotalCnt int
}

func (dvs *DteStat) TickEvent(format int, event int) {
	dvs.TotalCnt++
	if format == 0 {
		vc := (event >> 2) & 0x3F
		dvs.EvtStat[event&3].TickChannel(vc)
	} else {
		dvs.OtherCnt++
	}
}

// 0, 1 for engine end/start
// 2, 3 for vc execution end/start
func (dvs DteStat) ToString() string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "engine(%v), ", dvs.EvtStat[0].Total+dvs.EvtStat[1].Total)
	fmt.Fprintf(buf, "vc-end:%v, vc-start:%v",
		dvs.EvtStat[2].ToString(), dvs.EvtStat[3].ToString())
	return buf.String()
}

func (dvs DteStat) Empty() bool {
	return dvs.TotalCnt == 0
}
