package sess

import (
	"bytes"
	"fmt"
)

type IEngineEvtStat interface {
	TickEvent(format int, event int)
	ToString() string
	Empty() bool
}

type Format0Stat struct {
	Event    [256]int
	OtherCnt int
	TotalCnt int
}
type EngineEvtStat struct {
	Format0Stat
}

func (ee EngineEvtStat) ToString() string {
	buf := bytes.NewBuffer(nil)
	for evnt, cnt := range ee.Event {
		if cnt > 0 {
			fmt.Fprintf(buf, "%v:%v, ", evnt, cnt)
		}
	}
	if buf.Len() == 0 {
		fmt.Fprintf(buf, "%v", ee.TotalCnt)
	} else {
		if ee.OtherCnt > 0 {
			fmt.Fprintf(buf, "others: %v", ee.OtherCnt)
		}
	}
	return buf.String()
}

func (ee *EngineEvtStat) TickEvent(format int, event int) {
	ee.TotalCnt++
	if format == 0 {
		ee.Event[event]++
	} else {
		ee.OtherCnt++
	}
}

func (ee EngineEvtStat) Empty() bool {
	return ee.TotalCnt == 0
}
