package sess

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/efintf/sessintf"
)

type WorkSlot struct {
	subscribers map[codec.EngineTypeCode][]sessintf.ConcurEventSinker
	prevChan    chan<- codec.DpfEvent
	thisChan    <-chan codec.DpfEvent
	nameI       int

	unprocVec []codec.DpfEvent
}

func NewWorkSlot(nameI int, sinkers map[codec.EngineTypeCode][]sessintf.ConcurEventSinker) *WorkSlot {
	rv := &WorkSlot{
		nameI:       nameI,
		subscribers: make(map[codec.EngineTypeCode][]sessintf.ConcurEventSinker),
	}

	codec.EngineTypeCodeFor(func(tyCode codec.EngineTypeCode) {
		for _, subscriber := range sinkers[tyCode] {
			// index to index
			// be sure to restore in this sequence
			rv.subscribers[tyCode] =
				append(rv.subscribers[tyCode],
					subscriber.SelfClone())
		}
	})

	return rv
}

func (ws *WorkSlot) DoReduce(sinkers map[codec.EngineTypeCode][]sessintf.ConcurEventSinker) {
	codec.EngineTypeCodeFor(func(tyCode codec.EngineTypeCode) {
		// restore in the same sequence it is cloned
		for idx, subscriber := range ws.subscribers[tyCode] {
			subscriber.MergeTo(sinkers[tyCode][idx])
		}
	})
}

func (ws *WorkSlot) processSync() {
	if ws.thisChan == nil {
		return
	}
	// No default: Wait for the final close
	// default:
	processedCount := 0
	for evt := range ws.thisChan {
		processedCount++
		needPropagate := false
		for _, prevSinker := range ws.subscribers[evt.EngineTypeCode] {
			if err := prevSinker.DispatchEvent(evt); err != nil {
				needPropagate = true
				// Do not break
			}
		}
		if needPropagate {
			ws.CacheToUnprocessed(evt)
		}
	}
	fmt.Printf("%v has processed %v events from chan\n", ws.ToString(),
		processedCount)
}

func (ws *WorkSlot) PropagateToNext(evt codec.DpfEvent) {
	if ws.prevChan != nil {
		ws.prevChan <- evt
	}
}

func (ws *WorkSlot) CacheToUnprocessed(evt codec.DpfEvent) {
	ws.unprocVec = append(ws.unprocVec, evt)
}

func (ws *WorkSlot) FinalizeSlot() {
	ws.processSync()
	// propagate to previous instance
	if ws.prevChan != nil {
		for _, evt := range ws.unprocVec {
			ws.prevChan <- evt
		}
		// fmt.Printf("worker{%v} is closing previous chan\n", ws.ToString())
		close(ws.prevChan)
	}
}

func (ws WorkSlot) ToString() string {
	return fmt.Sprintf("WorkSlot{%v}", ws.nameI)
}
