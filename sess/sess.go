package sess

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/efintf/sessintf"
)

var (
	errInputValue = errors.New("input error")
)

type SessionOpt struct {
	EngineFilter string
	Debug        bool
	DecodeFull   bool
	Sort         bool
}

type Session struct {
	items   []codec.DpfEvent
	sessOpt SessionOpt
}

type DpfEventArray struct {
	array      []codec.DpfEvent
	errWatcher ErrorWatcher
}

func (d *DpfEventArray) AppendItem(item codec.DpfEvent) {
	d.array = append(d.array, item)
}

func NewSession(sessOpt SessionOpt) Session {
	return Session{sessOpt: sessOpt}
}

func (sess *Session) appendItem(newItem codec.DpfEvent) {
	sess.items = append(sess.items, newItem)
}

func (sess *Session) appendItemVector(newItemVec []codec.DpfEvent) {
	sess.items = append(sess.items, newItemVec...)
}

func getLow32(a uint64) uint32 {
	return uint32(a & ((1 << 32) - 1))
}

func getHigh32(a uint64) uint32 {
	return uint32(a >> 32)
}

/*
func (sess *Session) FakeStepEnd(decoder *codec.DecodeMaster) {
	lz := len(sess.items)
	if lz > 0 {
		last := sess.items[lz-1]
		lastCy := last.Cycle

		// [005d3f90: 00094616 00000012 96e0fe20 00000000]
		// [005d3fa0: 00094814 00000012 96e0ff3e 00000000]
		lastCy++
		raw0 := []uint32{0x00094616, 0x00000012, getLow32(lastCy), getHigh32(lastCy)}
		stepDoneStart, err := decoder.NewDpfEvent(raw0, 0)
		assert.Assert(err == nil, "must be nil")
		sess.appendItem(stepDoneStart)

		lastCy++
		raw1 := []uint32{0x00094814, 0x00000012, getLow32(lastCy), getHigh32(lastCy)}
		stepDoneEnd, err := decoder.NewDpfEvent(raw1, 0)
		assert.Assert(err == nil, "must be nil")
		sess.appendItem(stepDoneEnd)
	}
}
*/

// Process master text, no cache
func (sess *Session) ProcessMasterText(text string, decoder *codec.DecodeMaster) bool {
	if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
		text = text[2:]
	}
	if len(text) <= 0 {
		return true
	}
	text = strings.Trim(text, " ")
	val, err := strconv.ParseUint(text, 16, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parse hex: %v\n", err)
		return false
	}
	engineId, engineUniqIdx, ctx, clusterId, ok := decoder.GetEngineInfo(uint32(val))
	if !ok {
		fmt.Fprintf(os.Stderr, "decode error for 0x%08x\n", val)
		return false
	}
	if ctx >= 16 {
		panic(fmt.Errorf("assertion error: ctx = %v, val = 0x%x", ctx, val))
	}
	engineTypeStr := decoder.EngUniqueIndexToTypeName(engineUniqIdx)
	fmt.Printf("%08x  %v %v %v %v\n", val,
		engineTypeStr, clusterId, engineId, ctx)
	return true
}

func toItems(vs []string) []uint32 {
	var arr []uint32
	for _, s := range vs {
		s = strings.Trim(s, " ")
		if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
			s = s[2:]
		}
		if len(s) <= 0 {
			// fmt.Fprintf(os.Stderr, "not a valid number in hex format: '%v'\n", s)
			continue
		}
		val, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parse hex: %v\n", err)
			return nil
		}
		arr = append(arr, uint32(val))
	}
	return arr
}

func (sess *Session) ProcessFullItem(
	text string, offsetIdx int,
	decoder *codec.DecodeMaster,
	eventArray *DpfEventArray,
) (bool, error) {
	vs := toItems(strings.Split(text, " "))
	if len(vs) != 4 {
		return true, errInputValue
	}
	return sess.ProcessItems(vs, offsetIdx, decoder, eventArray)
}

// Process one item, always append
func (sess Session) ProcessItems(vs []uint32,
	offsetIdx int,
	decoder *codec.DecodeMaster,
	eventArray *DpfEventArray,
) (bool, error) {
	item, err := decoder.NewDpfEvent(vs, offsetIdx)
	if err != nil {
		eventArray.errWatcher.ReceiveError(vs, offsetIdx)
		return true, err
	}
	var toAdd = true
	if len(sess.sessOpt.EngineFilter) > 0 &&
		!strings.HasPrefix(item.EngineTy, sess.sessOpt.EngineFilter) {
		toAdd = false
	}
	if toAdd {
		eventArray.errWatcher.TickSuccess()
		eventArray.AppendItem(item)
	} else {
		eventArray.errWatcher.TickIgnore()
	}
	return true, nil
}

func (sess *Session) DecodeFromTextStream(
	inHandle *os.File,
	decoder *codec.DecodeMaster,
) {
	reader := bufio.NewReader(inHandle)
	var eventArr DpfEventArray
	eventArr.errWatcher = ErrorWatcher{printQuota: 10}
	for lineno := 0; ; lineno++ {
		// fmt.Print("-> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			if sess.sessOpt.Debug {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			break
		}
		text = strings.TrimSuffix(text, "\n")
		if sess.sessOpt.DecodeFull {
			shallCont, _ := sess.ProcessFullItem(text, lineno, decoder, &eventArr)
			if !shallCont {
				break
			}
		} else {
			if !sess.ProcessMasterText(text, decoder) {
				break
			}
		}
	}

	sess.appendItemVector(eventArr.array)
	if sess.sessOpt.Sort {
		sort.Sort(codec.DpfItems(sess.items))
	}
	eventArr.errWatcher.SumUp()
}

func (sess *Session) DecodeChunk(
	chunk []byte,
	decoder *codec.DecodeMaster,
	sugguestJobCount int,
) {
	decodeChunkStartTs := time.Now()
	log.Printf("starting decoding dpf raw data")

	if len(chunk)%16 != 0 {
		log.Printf("warning: dpf buffer length not xx divide by 16")
	}
	itemCount := len(chunk) / 16
	var jobCount = sugguestJobCount
	if jobCount <= 0 {
		jobCount = 1
	}
	segItemCount := (itemCount + (jobCount - 1)) / jobCount
	if segItemCount < 1000 {
		segItemCount = 1000
	}
	jobCount = (itemCount + (segItemCount - 1)) / segItemCount
	log.Printf("itemCount is [%v], jobCount [%v], segmentItemCount [%v]",
		itemCount,
		jobCount,
		segItemCount)
	eventResult := make([]DpfEventArray, jobCount)
	for p := 0; p < jobCount; p++ {
		eventResult[p].errWatcher = ErrorWatcher{printQuota: 10}
	}
	var waitGroup sync.WaitGroup
	subDecodeProcess := func(subChunk []byte,
		eventItemArray *DpfEventArray,
		baseIdx int) {
		defer waitGroup.Done()
		var errWatcher = &eventItemArray.errWatcher
		subItemsCount := len(subChunk) / 16
		for i := 0; i < subItemsCount; i++ {
			offsetIdx := i << 4
			var u32vals = [4]uint32{
				binary.LittleEndian.Uint32(subChunk[offsetIdx:]),
				binary.LittleEndian.Uint32(subChunk[offsetIdx+4:]),
				binary.LittleEndian.Uint32(subChunk[offsetIdx+8:]),
				binary.LittleEndian.Uint32(subChunk[offsetIdx+12:]),
			}
			sess.ProcessItems(u32vals[:],
				(offsetIdx+baseIdx)>>4,
				decoder,
				eventItemArray)
		}
		errWatcher.SumUp()
	}
	for p := 0; p < jobCount; p++ {
		waitGroup.Add(1)
		startItemIdx, endItemIdx := p*segItemCount, (p+1)*segItemCount
		if endItemIdx > itemCount {
			endItemIdx = itemCount
		}
		log.Printf("event processing [%v] - [%v]", startItemIdx, endItemIdx)
		go subDecodeProcess(
			chunk[16*startItemIdx:16*endItemIdx],
			&eventResult[p],
			startItemIdx*16)
	}
	waitGroup.Wait()
	log.Printf("done forking %v", time.Since(decodeChunkStartTs))
	// Reduce
	errCountInAll, ignoreInAll, okInAll := 0, 0, 0
	for _, result := range eventResult {
		sess.appendItemVector(result.array)
		errCountInAll += result.errWatcher.errCount
		ignoreInAll += result.errWatcher.ignoreCount
		okInAll += result.errWatcher.okCount
	}
	log.Printf("error in all: %v", errCountInAll)
	log.Printf("ignore in all: %v", ignoreInAll)
	log.Printf("success in all: %v", okInAll)
	assert.Assert(len(sess.items)+errCountInAll+ignoreInAll == itemCount, "must be the same for %v:%v",
		len(sess.items), itemCount)
	log.Printf("done decoding in %v", time.Since(decodeChunkStartTs))
	// after all items are in place.

	if sess.sessOpt.Sort {
		sort.Sort(codec.DpfItems(sess.items))
	}
}

func (sess Session) PrintItems(out io.Writer, printRaw bool) {
	if printRaw {
		for _, v := range sess.items {
			fmt.Fprintf(out, "%-50v : %v\n", v.ToString(), v.RawRepr())
		}
	} else {
		for _, v := range sess.items {
			fmt.Fprintf(out, "%v\n", v.ToString())
		}
	}
}

type SessBroadcaster struct {
	Session
	loader efintf.InfoReceiver
}

func NewSessBroadcaster(loader efintf.InfoReceiver) *SessBroadcaster {
	return &SessBroadcaster{
		Session: NewSession(SessionOpt{}),
		loader:  loader,
	}
}

func (sess SessBroadcaster) GetLoader() efintf.InfoReceiver {
	return sess.loader
}

func (sess *SessBroadcaster) DispatchToSinkers(
	sinkers ...sessintf.EventSinker,
) {
	// subscribers dict
	subscribers := make(map[codec.EngineTypeCode][]sessintf.EventSinker)

	// register for all
	for _, sinker := range sinkers {
		for _, typeCode := range sinker.GetEngineTypeCodes() {
			subscribers[typeCode] = append(subscribers[typeCode],
				sinker)
		}
	}

	// Finalize in sequential mode
	sess.emitEventsToSubscribersSequentials(subscribers)
	for _, sinker := range sinkers {
		sinker.Finalizes()
	}
}

func (sess *SessBroadcaster) DispatchToConcurSinkers(
	jobCount int,
	sinkers ...sessintf.ConcurEventSinker,
) {
	subs := make(map[codec.EngineTypeCode][]sessintf.ConcurEventSinker)
	for _, sub := range sinkers {
		for _, typeCode := range sub.GetEngineTypeCodes() {
			subs[typeCode] = append(subs[typeCode], sub)
		}
	}

	sinkerCount := len(sinkers)
	disSinkCount := 0
	for _, subVec := range subs {
		disSinkCount += len(subVec)
	}
	assert.Assert(sinkerCount <= disSinkCount,
		"must be the less-equal to ?? (%v, vs %v)",
		sinkerCount,
		disSinkCount)

	startTs := time.Now()
	sess.emitEventsToSubscribersEx(jobCount, subs, ioutil.Discard)
	fmt.Printf("# event dispatching cost %v\n", time.Since(startTs))
}

func (sess SessBroadcaster) emitEventsToSubscribersEx(
	jobCount int,
	sinkers map[codec.EngineTypeCode][]sessintf.ConcurEventSinker,
	dbgStream io.Writer,
) {
	// Divide the cake
	totCount := len(sess.items)
	if totCount <= 0 {
		fmt.Fprintf(os.Stderr, "#Error : No dpf buffer\n")
		os.Exit(1)
	}
	workerItemCount, segmentSize := DefaultJobDivider().
		DetermineWorkThread(jobCount,
			totCount)

	// Create work slot array
	workers := make([]*WorkSlot, workerItemCount)

	// Clone work slot
	for i := 0; i < workerItemCount; i++ {
		workers[i] = NewWorkSlot(i, sinkers)
	}

	// Create working channels
	channels := make([]chan []codec.DpfEvent, workerItemCount)
	const BUFSIZ = 1
	for i := 0; i < workerItemCount; i++ {
		channels[i] = make(chan []codec.DpfEvent, BUFSIZ)
	}
	for i := 0; i < workerItemCount; i++ {
		if i > 0 {
			workers[i].prevChan = channels[i-1]
		}
		if i < workerItemCount-1 {
			workers[i].thisChan = channels[i]
		}
	}

	// Launch go-routines carrying the real work
	var wg sync.WaitGroup
	workerFunc := func(eventSlice []codec.DpfEvent, wSlot *WorkSlot, nameI int) {
		defer wg.Done()
		startTs := time.Now()
		fmt.Fprintf(dbgStream, "%v working on %v item(s), evntTypesCount(%v),\n",
			wSlot.ToString(),
			len(eventSlice),
			len(wSlot.subscribers))

		for _, evt := range eventSlice {
			needPropagate := false // per event. Do OR logic
			for _, subscriber := range wSlot.subscribers[evt.EngineTypeCode] {
				err := subscriber.DispatchEvent(evt)
				if err != nil {
					needPropagate = true
					// and no break
				}
			}
			if needPropagate {
				wSlot.CacheToUnprocessed(evt)
			}
		}
		wSlot.FinalizeSlot()
		fmt.Fprintf(dbgStream, "%v is quitting. %v consumed\n",
			wSlot.ToString(),
			time.Since(startTs),
		)
	}

	// Now start it
	for i := 0; i < workerItemCount; i++ {
		start, endi := i*segmentSize, (i+1)*segmentSize
		if endi > totCount {
			endi = totCount
		}
		wg.Add(1)
		go workerFunc(sess.items[start:endi], workers[i], i)
	}

	// Finalize
	wg.Wait()

	// Merge results
	fmt.Fprintf(dbgStream, "starting merging results\n")
	for i := 0; i < workerItemCount; i++ {
		fmt.Fprintf(dbgStream, "merging with [%v]...\n", i)
		startTs := time.Now()
		workers[i].DoReduce(sinkers)
		fmt.Fprintf(dbgStream, "done in %v\n", time.Since(startTs))
	}
	fmt.Fprintf(dbgStream, "done merging\n")
}

func (sess SessBroadcaster) emitEventsToSubscribersSequentials(
	subscribers map[codec.EngineTypeCode][]sessintf.EventSinker,
) {
	errCount := 0
	const ErrDisplayCountLimit = 30
	// The original way

	for _, evt := range sess.items {
		for _, subscriber := range subscribers[evt.EngineTypeCode] {
			err := subscriber.DispatchEvent(evt)
			if err != nil {
				errCount++
				if errCount < ErrDisplayCountLimit {
					fmt.Printf("error dispatch event: %v\n", err)
				} else if errCount == ErrDisplayCountLimit {
					fmt.Printf("too many errors for event dispatching\n")
				}
			}
		}
	}
	fmt.Printf("# error count: %v\n", errCount)
}
