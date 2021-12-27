package rtinfo

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/algo"
	"git.enflame.cn/hai.bai/dmaster/algo/linklist"
	"git.enflame.cn/hai.bai/dmaster/codec"
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
