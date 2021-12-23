package utils

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/algo"
	"git.enflame.cn/hai.bai/dmaster/codec"
)

type DpfAct struct {
	Start codec.DpfEvent
	End   codec.DpfEvent
}

type CqmEventQueue struct {
	distr []Lnk
	acts  []DpfAct
	algo  algo.ActMatchAlgo
}

func NewCqmEventQueue(algo algo.ActMatchAlgo) *CqmEventQueue {
	rv := CqmEventQueue{
		distr: NewLnkArray(algo.GetChannelNum()),
		algo:  algo,
	}
	return &rv
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
	if start := q.distr[index].Extract(este); start != nil {
		q.acts = append(q.acts, DpfAct{
			Start: *start,
			End:   este,
		})
		return nil
	}
	return fmt.Errorf("could not find start for %v", este.ToString())
}

func (q CqmEventQueue) DumpInfo() {
	fmt.Printf("%v acts found\n", len(q.acts))
	for ch, v := range q.distr {
		if v.elCount > 0 {
			engIdx, ctx := q.algo.DecodeChan(ch)
			fmt.Printf("Engine(%d) Ctx(%d) has %v in dangle\n",
				engIdx, ctx, v.elCount,
			)
			for ptr := v.head.Next; ptr != nil; ptr = ptr.Next {
				fmt.Printf("%v %v\n",
					ptr.dpfEvent.ToString(),
					ptr.dpfEvent.RawRepr(),
				)
			}
		}
	}
}
