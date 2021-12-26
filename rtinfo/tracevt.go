package rtinfo

import (
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
	Name string  `json:"name"`
	Ph   string  `json:"ph"`
	Pid  string  `json:"pid"`
	Tid  string  `json:"tid"`
	Ts   float64 `json:"ts"`
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

func toMs(hosttime uint64) float64 {
	return float64(hosttime) / (1000 * 1000)
}

func NewTraceEventBegin(
	rtTask RuntimeTask,
	op meta.DtuOp,
	ts uint64,
) TraceEvent {
	return TraceEvent{
		Ph:   "B",
		Ts:   toMs(ts),
		Pid:  rtTask.ToShortString(),
		Name: op.OpName,
	}
}

func NewTraceEventEnd(
	rtTask RuntimeTask,
	op meta.DtuOp,
	ts uint64,
) TraceEvent {
	return TraceEvent{
		Ph:   "E",
		Ts:   toMs(ts),
		Pid:  rtTask.ToShortString(),
		Name: op.OpName,
	}
}
