package rtdata

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/misc/linklist"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type ActCollector interface {
	AddAct(start, end codec.DpfEvent)
	GetAlgo() vgrule.ActMatchAlgo
	DumpInfo()
	GetActivity() interface{}
}

type EventQueue struct {
	ActCollector
	distr     []linklist.Lnk
	evtFilter EventFilter
}

type EventFilter interface {
	IsStarterMark(codec.DpfEvent) (bool, bool)
	TestIfMatch(codec.DpfEvent, codec.DpfEvent) bool
	GetEngineTypes() []codec.EngineTypeCode
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

func (q EventQueue) GetEngineTypeCodes() []codec.EngineTypeCode {
	return q.evtFilter.GetEngineTypes()
}

func (q *EventQueue) DispatchEvent(este codec.DpfEvent) error {
	index := q.GetAlgo().MapToChan(
		este.MasterIdValue(),
		este.Context,
	)
	isStart, isEnd := q.evtFilter.IsStarterMark(este)
	if isStart {
		q.distr[index].AppendAtFront(este)
		return nil
	}
	if !isEnd {
		// Filter-out
		return nil
	}
	if start := q.distr[index].Extract(func(one interface{}) bool {
		un := one.(codec.DpfEvent)
		return q.evtFilter.TestIfMatch(un, este)
	}); start != nil {
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

// Some casters
func (q EventQueue) OpActivity() []OpActivity {
	return ([]OpActivity)(q.GetActivity().(OpActivityVector))
}
