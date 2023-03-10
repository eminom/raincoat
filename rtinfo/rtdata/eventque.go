package rtdata

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/efintf/sessintf"
	"git.enflame.cn/hai.bai/dmaster/misc/linklist"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type ActCollector interface {
	AddAct(start, end codec.DpfEvent)
	AddDebugEvent(codec.DpfEvent)
	GetDebugEventVec() []codec.DpfEvent
	GetAlgo() vgrule.ActMatchAlgo
	DumpInfo()
	GetActivity() interface{}
	ActCount() int

	MergeInto(ActCollector)
	AxSelfClone() ActCollector

	DoSort()
}

type EventQueue struct {
	ActCollector
	distr     []linklist.Lnk
	evtFilter EventFilter
	finCall   int
}

type EventFilter interface {
	// Starter, Closer, and the Terminator
	IsStarterMark(codec.DpfEvent) (bool, bool, bool)
	IsRecyclable(codec.DpfEvent) bool
	TestIfMatch(codec.DpfEvent, codec.DpfEvent) bool
	GetEngineTypes() []codec.EngineTypeCode
	PurgePreviousEvents() bool
}

func NewOpEventQueue(act ActCollector,
	evtFilter EventFilter,
) *EventQueue {
	rv := EventQueue{
		ActCollector: act,
		distr:        linklist.NewLnkArray(act.GetAlgo().GetChannelNum()),
		evtFilter:    evtFilter,
	}
	return &rv
}

func (q EventQueue) GetEventFilter() EventFilter {
	return q.evtFilter
}

func (q EventQueue) GetEngineTypeCodes() []codec.EngineTypeCode {
	return q.evtFilter.GetEngineTypes()
}

func (q *EventQueue) DispatchEvent(este codec.DpfEvent) error {
	index := q.GetAlgo().MapToChan(
		este.MasterIdValue(),
		este.Context,
	)

	// Since este carries a context id
	// May be the purgation shall only affect to the Context scope only?
	// Some facts:
	// Pavo for now:
	//    1. only holds context id value of ZERO
	//    2. there are NO multiple Step acts simutaneously
	//         So purging over all Debug packts events works

	// Terminator use to be an interface method
	// But to make this problem general we shall not take special steps
	// if q.evtFilter.IsTerminator(este) {
	// 	q.distr[index].PurgeContent()
	// 	q.DoMergeOpAct()
	// 	return nil
	// }

	isStart, isEnd, isTerminator := q.evtFilter.IsStarterMark(este)
	if isStart {
		q.distr[index].AppendAtFront(este)
		return nil
	}
	// a termiator shall insert into the sequence
	if isTerminator {
		q.ActCollector.AddAct(este, este)
		return nil
	}

	if !isEnd {
		// Filter-out
		return nil
	}

	tester := func(one interface{}) bool {
		un := one.(codec.DpfEvent)
		return q.evtFilter.TestIfMatch(un, este)
	}

	if start := q.distr[index].Extract(tester); start != nil {
		if q.evtFilter.PurgePreviousEvents() {
			for q.distr[index].Extract(tester) != nil {
			}
		}

		startUn := start.(codec.DpfEvent)
		q.ActCollector.AddAct(startUn, este)
		return nil
	}
	return fmt.Errorf("could not find start for %v", este.ToString())
}

func (q EventQueue) DumpInfo() {
	q.ActCollector.DumpInfo()
	for ch, v := range q.distr {
		if v.ElementCount() > 0 {
			masterVal, ctx := q.GetAlgo().DecodeChan(ch)
			engTy, engIdx, clusterId := q.GetAlgo().DecodeMasterValue(masterVal)
			fmt.Printf("Engine %v(%v) Cid(%v) Ctx(%d) has %v in dangle\n",
				engTy, engIdx, clusterId, ctx, v.ElementCount(),
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

func (q *EventQueue) collectUnmatchedStart() {
	for _, v := range q.distr {
		v.ConstForEach(func(evt interface{}) {
			dpfEvent := evt.(codec.DpfEvent)
			if q.evtFilter.IsRecyclable(dpfEvent) {
				q.ActCollector.AddDebugEvent(dpfEvent)
			}
		})
	}
}

func (q *EventQueue) Finalizes() {
	q.finCall++
	assert.Assert(q.finCall == 1, "must be eq to 1(%v)", q.finCall)
	q.collectUnmatchedStart()
}

// Some casters
func (q EventQueue) OpActivity() []OpActivity {
	return ([]OpActivity)(q.GetActivity().(OpActivityVector))
}

func (q EventQueue) DmaActivity() []DmaActivity {
	return ([]DmaActivity)(q.GetActivity().(DmaActivityVec))
}

func (q EventQueue) FwActivity() []FwActivity {
	return ([]FwActivity)(q.GetActivity().(FwActivityVec))
}

func (q EventQueue) KernelActivity() []KernelActivity {
	return ([]KernelActivity)(q.GetActivity().(KernelActivityVec))
}

func (q EventQueue) TaskActivity() []TaskActivity {
	return ([]TaskActivity)(q.GetActivity().(TaskActivityVec))
}

func (q EventQueue) AllZero() bool {
	for _, el := range q.distr {
		if el.ElementCount() > 0 {
			return false
		}
	}
	return true
}

func (q EventQueue) SelfClone() sessintf.ConcurEventSinker {
	assert.Assert(q.AllZero(), "Must be empty")

	cloned := &EventQueue{
		ActCollector: q.ActCollector.AxSelfClone(),
		distr:        linklist.NewLnkArray(q.GetAlgo().GetChannelNum()),
		evtFilter:    q.evtFilter,
	}
	return cloned
}

func (cloned EventQueue) MergeTo(lhs interface{}) bool {
	master := lhs.(*EventQueue)
	cloned.MergeInto(master.ActCollector)
	return true
}
