package rtdata

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/misc/linklist"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type OpActivityVector []OpActivity

type OpEventQueue struct {
	distr []linklist.Lnk
	acts  OpActivityVector
	eAlgo vgrule.ActMatchAlgo
}

func NewOpEventQueue(algo vgrule.ActMatchAlgo) *OpEventQueue {
	rv := OpEventQueue{
		distr: linklist.NewLnkArray(algo.GetChannelNum()),
		eAlgo: algo,
	}
	return &rv
}

// In-place cook
func (q *OpEventQueue) OpActivity() []OpActivity {
	return q.acts
}

func (q *OpEventQueue) PutEvent(este codec.DpfEvent) error {
	index := q.eAlgo.MapToChan(
		este.EngineIndex,
		este.Context,
	)
	if este.Event == codec.CqmEventOpStart {
		q.distr[index].AppendNode(este)
		return nil
	}
	if start := q.distr[index].Extract(func(one interface{}) bool {
		un := one.(codec.DpfEvent)
		return un.PacketID+1 == este.PacketID && un.ClusterID == este.ClusterID
	}); start != nil {
		startUn := start.(codec.DpfEvent)
		q.acts = append(q.acts, OpActivity{
			DpfAct: DpfAct{
				Start: startUn,
				End:   este,
			},
		})
		return nil
	}
	return fmt.Errorf("could not find start for %v", este.ToString())
}

func (q OpEventQueue) DumpInfo() {
	fmt.Printf("%v acts found\n", len(q.acts))

	chDictInAll := make(map[int]int)
	for _, v := range q.acts {
		index := q.eAlgo.MapToChan(
			v.Start.EngineIndex,
			v.Start.Context,
		)
		chDictInAll[index]++
	}

	fmt.Printf("Cqm Op debug packet distribution:\n")
	for index, count := range chDictInAll {
		engId, ctx := q.eAlgo.DecodeChan(index)
		fmt.Printf(" Cqm(%v) ctx(%v) count: %v\n",
			engId, ctx, count,
		)
	}

	for ch, v := range q.distr {
		if v.ElementCount() > 0 {
			engIdx, ctx := q.eAlgo.DecodeChan(ch)
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
