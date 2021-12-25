package rtinfo

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	"git.enflame.cn/hai.bai/dmaster/algo"
	"git.enflame.cn/hai.bai/dmaster/algo/linklist"
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/meta"
)

type CqmActBundles []CqmActBundle

type CqmEventQueue struct {
	distr []linklist.Lnk
	acts  CqmActBundles
	algo  algo.ActMatchAlgo
}

func NewCqmEventQueue(algo algo.ActMatchAlgo) *CqmEventQueue {
	rv := CqmEventQueue{
		distr: linklist.NewLnkArray(algo.GetChannelNum()),
		algo:  algo,
	}
	return &rv
}

// In-place cook
func (q *CqmEventQueue) CqmActBundle() []CqmActBundle {
	return q.acts
}

func (q *CqmEventQueue) PutEvent(este codec.DpfEvent) error {
	index := q.algo.MapToChan(
		este.EngineIndex,
		este.Context,
	)
	if este.Event == codec.CqmEventOpStart {
		q.distr[index].AppendNode(este)
		return nil
	}
	if start := q.distr[index].Extract(func(one interface{}) bool {
		un := one.(codec.DpfEvent)
		return un.PacketID+1 == este.PacketID
	}); start != nil {
		startUn := start.(codec.DpfEvent)
		q.acts = append(q.acts, CqmActBundle{
			DpfAct: DpfAct{
				Start: startUn,
				End:   este,
			},
		})
		return nil
	}
	return fmt.Errorf("could not find start for %v", este.ToString())
}

func (q CqmEventQueue) DumpInfo() {
	fmt.Printf("%v acts found\n", len(q.acts))

	chDictInAll := make(map[int]int)
	for _, v := range q.acts {
		index := q.algo.MapToChan(
			v.Start.EngineIndex,
			v.Start.Context,
		)
		chDictInAll[index]++
	}

	fmt.Printf("Cqm Op debug packet distribution:\n")
	for index, count := range chDictInAll {
		engId, ctx := q.algo.DecodeChan(index)
		fmt.Printf(" Cqm(%v) ctx(%v) count: %v\n",
			engId, ctx, count,
		)
	}

	for ch, v := range q.distr {
		if v.ElementCount() > 0 {
			engIdx, ctx := q.algo.DecodeChan(ch)
			fmt.Printf("Engine(%d) Ctx(%d) has %v in dangle\n",
				engIdx, ctx, v.ElementCount(),
			)
			v.ConstForEach(func(evt interface{}) {
				dpfEvent := evt.(codec.DpfEvent)
				fmt.Printf("%v %v\n",
					dpfEvent.ToString(),
					dpfEvent.RawRepr(),
				)
			})
		}
	}
}

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

func (bundle CqmActBundles) DumpToEventTrace(out string) {

	var eventVec []TraceEvent
	var dtuOpCount = 0
	for _, act := range bundle {
		if act.opRef.dtuOp != nil {
			dtuOpCount++
			eventVec = append(eventVec, NewTraceEventBegin(
				act.opRef.pgMask,
				*act.opRef.dtuOp,
				act.StartCycle(),
			))
			eventVec = append(eventVec, NewTraceEventEnd(
				act.opRef.pgMask,
				*act.opRef.dtuOp,
				act.EndCycle(),
			))
		}
	}

	sort.Sort(TraceEvents(eventVec))
	chunk, err := json.MarshalIndent(eventVec, "", "  ")
	if err != nil {
		panic(err)
	}

	fout, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer fout.Close()
	fout.Write(chunk)
	log.Printf("%v dtuop(s) have been written to %v\n",
		dtuOpCount,
		out,
	)
}
