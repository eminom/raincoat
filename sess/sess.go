package sess

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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

func NewSession(sessOpt SessionOpt) Session {
	return Session{sessOpt: sessOpt}
}

func (sess *Session) AppendItem(newItem codec.DpfEvent) {
	sess.items = append(sess.items, newItem)
}

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
) (bool, error) {
	vs := toItems(strings.Split(text, " "))
	if len(vs) != 4 {
		return true, errInputValue
	}
	return sess.ProcessItems(vs, offsetIdx, decoder)
}

// Process one item, always append
func (sess *Session) ProcessItems(vs []uint32,
	offsetIdx int,
	decoder *codec.DecodeMaster,
) (bool, error) {
	item, err := decoder.NewDpfEvent(vs, offsetIdx)
	if err != nil {
		return true, err
	}
	var toAdd = true
	if len(sess.sessOpt.EngineFilter) > 0 &&
		!strings.HasPrefix(item.EngineTy, sess.sessOpt.EngineFilter) {
		toAdd = false
	}
	if toAdd {
		sess.AppendItem(item)
	}
	return true, nil
}

func (sess *Session) DecodeFromTextStream(
	inHandle *os.File,
	decoder *codec.DecodeMaster,
) {
	reader := bufio.NewReader(inHandle)
	var errWatcher = ErrorWatcher{printQuota: 10}
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
			shallCont, err := sess.ProcessFullItem(text, lineno, decoder)
			if nil != err {
				errWatcher.ReceiveErr(err)
			} else {
				errWatcher.TickSuccess()
			}
			if !shallCont {
				break
			}
		} else {
			if !sess.ProcessMasterText(text, decoder) {
				break
			}
		}
	}

	if sess.sessOpt.Sort {
		sort.Sort(codec.DpfItems(sess.items))
	}

	errWatcher.SumUp()
}

func (sess *Session) DecodeChunk(
	chunk []byte,
	decoder *codec.DecodeMaster,
) {
	// realpath, e2 := os.Readlink(filename)
	// if nil == e2 {
	// 	filename = realpath
	// }
	itemSize := len(chunk) / 16 * 16
	var errWatcher = ErrorWatcher{printQuota: 10}
	for i := 0; i < itemSize; i += 16 {
		offsetIdx := i >> 4
		var u32vals = [4]uint32{
			binary.LittleEndian.Uint32(chunk[i:]),
			binary.LittleEndian.Uint32(chunk[i+4:]),
			binary.LittleEndian.Uint32(chunk[i+8:]),
			binary.LittleEndian.Uint32(chunk[i+12:]),
		}
		_, err := sess.ProcessItems(u32vals[:], offsetIdx, decoder)
		if err != nil {
			errWatcher.ReceiveError(u32vals[:], offsetIdx)
		} else {
			errWatcher.TickSuccess()
		}
	}
	if sess.sessOpt.Sort {
		sort.Sort(codec.DpfItems(sess.items))
	}
	errWatcher.SumUp()
}

func (sess Session) PrintItems(printRaw bool) {
	if printRaw {
		for _, v := range sess.items {
			fmt.Printf("%-50v : %v\n", v.ToString(), v.RawRepr())
		}
	} else {
		for _, v := range sess.items {
			fmt.Println(v.ToString())
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

	sess.emitEventsToSubscribersSequentials(subscribers)
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
	sess.emitEventsToSubscribersEx(jobCount, subs)
}

func (sess SessBroadcaster) emitEventsToSubscribersEx(
	jobCount int,
	sinkers map[codec.EngineTypeCode][]sessintf.ConcurEventSinker,
) {
	// Divide the cake
	totCount := len(sess.items)
	workerItemCount, segmentSize := DetermineWorkThread(jobCount,
		len(sess.items))

	// Create work slot array
	workers := make([]*WorkSlot, workerItemCount)

	// Clone work slot
	for i := 0; i < workerItemCount; i++ {
		workers[i] = NewWorkSlot(i, sinkers)
	}

	// Create working channels
	channels := make([]chan codec.DpfEvent, workerItemCount)
	const BUFSIZ = 16
	for i := 0; i < workerItemCount; i++ {
		channels[i] = make(chan codec.DpfEvent, BUFSIZ)
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
		fmt.Printf("%v working on %v item(s), evntTypesCount(%v),\n",
			wSlot.ToString(),
			len(eventSlice),
			len(wSlot.subscribers))

		for _, evt := range eventSlice {
			for _, subscriber := range wSlot.subscribers[evt.EngineTypeCode] {
				err := subscriber.DispatchEvent(evt)
				if err != nil {
					wSlot.CacheToUnprocessed(evt)
				}
			}
		}
		wSlot.FinalizeSlot()
		fmt.Printf("%v is quitting\n", wSlot.ToString())
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
	fmt.Printf("starting merging results\n")
	for i := 0; i < workerItemCount; i++ {
		fmt.Printf("merging with [%v]...\n", i)
		startTs := time.Now()
		workers[i].DoReduce(sinkers)
		fmt.Printf("done in %v\n", time.Since(startTs))
	}
	fmt.Printf("done merging\n")
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
