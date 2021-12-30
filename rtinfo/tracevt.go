package rtinfo

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
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
	pid string,
	name string,
	ts uint64,
) TraceEvent {
	return TraceEvent{
		Ph:   "B",
		Ts:   toMs(ts),
		Pid:  pid,
		Name: name,
	}
}

func NewTraceEventEnd(
	pid string,
	name string,
	ts uint64,
) TraceEvent {
	return TraceEvent{
		Ph:   "E",
		Ts:   toMs(ts),
		Pid:  pid,
		Name: name,
	}
}

func NewTraceEventStartUnk(ts uint64,
	sub string,
	name string,
) TraceEvent {
	return TraceEvent{
		Ph:   "B",
		Ts:   toMs(ts),
		Pid:  name,
		Name: sub,
	}
}

func NewTraceEventEndUnk(ts uint64,
	sub string,
	name string,
) TraceEvent {
	return TraceEvent{
		Ph:   "E",
		Ts:   toMs(ts),
		Pid:  name,
		Name: sub,
	}
}

type TraceEventSession struct {
	eventVec []TraceEvent
}

func (tr *TraceEventSession) AppendEvt(evt TraceEvent) {
	tr.eventVec = append(tr.eventVec, evt)
}

func checkTimespanOverlapping(bundle []CqmActBundle) {
	log.Printf("start checking overlap over %v item(s)", len(bundle))
	var intvs Interval
	// Check intervals
	for _, act := range bundle {
		if act.opRef.dtuOp != nil {
			intvs = append(intvs, []uint64{
				act.StartCycle(),
				act.EndCycle(),
			})
		}
	}
	sort.Sort(intvs)
	overlappedCount := 0
	for i := 0; i < len(intvs)-1; i++ {
		if intvs[i][1] >= intvs[i+1][0] {
			log.Printf("error: %v >= %v at [%d] out of [%v]",
				intvs[i][1], intvs[i+1][0],
				i, len(intvs))
			overlappedCount++
			break
		}
	}
	// dumpIntvs(intvs)
	if overlappedCount > 0 {
		fmt.Printf("warning: there is %v overlapping\n",
			overlappedCount)
		// assert.Assert(overlappedCount == 0,
		// 	"overlapped count must be zero: but %v",
		// 	overlappedCount)
	} else {
		log.Printf("no overlapped confirmed")
	}
}

func (tr *TraceEventSession) DumpToEventTrace(
	bundle []CqmActBundle,
	tm *TimelineManager,
	getPidAndName func(CqmActBundle) (bool, string, string),
	dumpWild bool,
) {
	checkTimespanOverlapping(bundle)

	var dtuOpCount = 0
	var convertToHostError = 0

	subSampleCount := 0
	for _, act := range bundle {
		///act.IsOpRefValid()
		if okToShow, pid, name := getPidAndName(act); okToShow {
			dtuOpCount++
			startHostTime, startOK := tm.MapToHosttime(act.StartCycle())
			endHostTime, endOK := tm.MapToHosttime(act.EndCycle())
			if startOK && endOK {
				tr.AppendEvt(NewTraceEventBegin(
					pid,
					name,
					startHostTime,
				))
				tr.AppendEvt(NewTraceEventEnd(
					pid,
					name,
					endHostTime,
				))
			} else {
				convertToHostError++
			}
		} else {
			subSampleCount++
			if dumpWild {
				startHostTime, startOK := tm.MapToHosttime(act.StartCycle())
				endHostTime, endOK := tm.MapToHosttime(act.EndCycle())
				if startOK && endOK && subSampleCount%30 == 0 {
					tr.AppendEvt(NewTraceEventStartUnk(startHostTime, pid, name))
					tr.AppendEvt(NewTraceEventEndUnk(endHostTime, pid, name))
				}
			}
		}
	}
	if convertToHostError > 0 {
		log.Printf("convert-to-hosttime-error count: %v", convertToHostError)
	}
}

func (tr TraceEventSession) DumpToFile(out string) {

	sort.Sort(TraceEvents(tr.eventVec))

	// const HourMes uint64 = 60 * 1000 * 1000
	// const MinMes uint64 = 1000 * 1000
	// for i := 0; i < len(eventVec); i += 2 {
	// 	idx := i / 2
	// 	eventVec[i].Ts = uint64(idx) * HourMes
	// 	eventVec[i+1].Ts = uint64(idx+1)*HourMes - MinMes
	// }

	chunk, err := json.MarshalIndent(tr.eventVec, "", "  ")
	if err != nil {
		panic(err)
	}

	fout, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer fout.Close()
	fout.Write(chunk)
	log.Printf("%v dtuop(s) have been written to %v successfully\n",
		len(tr.eventVec),
		out,
	)
}