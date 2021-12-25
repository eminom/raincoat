package rtinfo

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/meta"
)

/*
[ {"name": "出方案", "ph": "B", "pid": "Main", "tid": "工作", "ts": 0},
  {"name": "出方案", "ph": "E", "pid": "Main", "tid": "工作", "ts": 28800000000},
  {"name": "看电影", "ph": "B", "pid": "Main", "tid": "休闲", "ts": 28800000000},
  {"name": "看电影", "ph": "E", "pid": "Main", "tid": "休闲", "ts": 32400000000},
  {"name": "写代码", "ph": "B", "pid": "Main", "tid": "工作", "ts": 32400000000},
  {"name": "写代码", "ph": "E", "pid": "Main", "tid": "工作", "ts": 36000000000},
  {"name": "遛狗", "ph": "B", "pid": "Main", "tid": "休闲", "ts": 36000000000},
  {"name": "遛狗", "ph": "E", "pid": "Main", "tid": "休闲", "ts": 37800000000}

*/
type TraceEvent struct {
	Name string `json:"name"`
	Ph   string `json:"ph"`
	Pid  string `json:"pid"`
	Tid  string `json:"tid"`
	Ts   uint64 `json:"ts"`
}

type TraceEvents []TraceEvent

func (t TraceEvents) Len() int {
	return len(t)
}

func (t TraceEvents) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TraceEvents) Less(i, j int) bool {
	return t[i].Ts < t[j].Ts
}

func pgMaskToString(pgMask int) string {
	return fmt.Sprintf("PG %v", pgMask)
}

func NewTraceEventBegin(
	pgMask int,
	op meta.DtuOp,
	ts uint64,
) TraceEvent {
	return TraceEvent{
		Ph:   "B",
		Ts:   ts,
		Pid:  pgMaskToString(pgMask),
		Name: op.OpName,
	}
}

func NewTraceEventEnd(
	pgMask int,
	op meta.DtuOp,
	ts uint64,
) TraceEvent {
	return TraceEvent{
		Ph:   "E",
		Ts:   ts,
		Pid:  pgMaskToString(pgMask),
		Name: op.OpName,
	}
}
