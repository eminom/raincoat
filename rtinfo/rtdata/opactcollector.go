package rtdata

import (
	"fmt"
	"sort"
	"time"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type OpIdMapper interface {
	GetOpIdForPacketId(int) (int, bool)
}

type OpCacheQueue struct {
	// Context to packet ID
	cacheDict map[int]map[int]*OpActivity
	oMapper   OpIdMapper
}

func NewOpCacheQueue(mapper OpIdMapper) *OpCacheQueue {
	return &OpCacheQueue{
		cacheDict: make(map[int]map[int]*OpActivity),
		oMapper:   mapper,
	}
}

func (a *OpCacheQueue) PurgeOps() OpActivityVector {
	var arr OpActivityVector
	for _, dc := range a.cacheDict {
		for _, act := range dc {
			arr = append(arr, *act)
		}
	}
	sort.Sort(arr)
	a.cacheDict = make(map[int]map[int]*OpActivity)
	return arr
}

func (a *OpCacheQueue) AddAct(act OpActivity) {
	ctxId := act.Start.Context
	if _, ok := a.cacheDict[ctxId]; !ok {
		a.cacheDict[ctxId] = make(map[int]*OpActivity)
	}

	pktId := act.Start.PacketID
	opId, ok := a.oMapper.GetOpIdForPacketId(pktId)
	if !ok {
		fmt.Printf("# error no op id for packet id %v\n", pktId)
		// Nothing has been added
		return
	}
	if that, ok := a.cacheDict[ctxId][opId]; ok {
		that.Start.Cycle = minU64(that.StartCycle(), act.StartCycle())
		that.End.Cycle = maxU64(that.EndCycle(), act.EndCycle())
	} else {
		newAct := act
		a.cacheDict[ctxId][opId] = &newAct
	}
}

type OpActCollector struct {
	acts          OpActivityVector
	eAlgo         vgrule.ActMatchAlgo
	cacheAndMerge bool
	opCache       *OpCacheQueue
}

type OpActCollectorOpt struct {
	OpIdMapper
	CacheAndMerge bool
}

func NewOpActCollector(algo vgrule.ActMatchAlgo, opt OpActCollectorOpt) *OpActCollector {
	return &OpActCollector{
		eAlgo:         algo,
		cacheAndMerge: opt.CacheAndMerge,
		opCache:       NewOpCacheQueue(opt.OpIdMapper),
	}
}

func (q OpActCollector) GetActivity() interface{} {
	return q.acts
}

func (q OpActCollector) GetAlgo() vgrule.ActMatchAlgo {
	return q.eAlgo
}

func (q OpActCollector) DumpInfo() {
	fmt.Printf("%v acts found\n", len(q.acts))

	chDictInAll := make(map[int]int)
	for _, v := range q.acts {
		index := q.eAlgo.MapToChan(
			v.Start.MasterIdValue(),
			v.Start.Context,
		)
		chDictInAll[index]++
	}

	fmt.Printf("Cqm Op debug packet distribution:\n")
	for index, count := range chDictInAll {
		masterVal, ctx := q.eAlgo.DecodeChan(index)
		engTy, engIdx, clusterId := q.eAlgo.DecodeMasterValue(masterVal)
		fmt.Printf(" %v(%v) Cid(%v) ctx(%v) count: %v\n",
			engTy, engIdx, clusterId, ctx, count,
		)
	}
}

func (opVec *OpActCollector) AddAct(start, end codec.DpfEvent) {
	newAct := OpActivity{
		DpfAct: DpfAct{
			start, end,
		},
	}
	if opVec.cacheAndMerge {
		opVec.opCache.AddAct(newAct)
	} else {
		// Append directly
		opVec.acts = append(opVec.acts, newAct)
	}
}

func (opVec *OpActCollector) DoMergeOpAct() {
	if opVec.cacheAndMerge {
		// fmt.Printf("#####################\n")
		// fmt.Printf("#####################\n")
		// fmt.Printf("#####################\n")
		// fmt.Printf("#####################\n")
		// fmt.Printf("#####################\n")
		opVec.acts = append(opVec.acts, opVec.opCache.PurgeOps()...)
	}
}

func (opVec OpActCollector) ActCount() int {
	return len(opVec.acts)
}

func (opVec OpActCollector) AxSelfClone() ActCollector {
	return &OpActCollector{eAlgo: opVec.eAlgo}
}

func (opVec OpActCollector) MergeInto(lhs ActCollector) {
	master := lhs.(*OpActCollector)
	// opVec.DoSort()
	fmt.Printf("merge %v OpActs into master(currently %v)\n",
		len(opVec.acts), len(master.acts),
	)
	master.acts = append(master.acts, opVec.acts...)
}

func (opVec OpActCollector) DoSort() {
	startTs := time.Now()
	sort.Sort(opVec.acts)
	fmt.Printf("sort %v dtuops in %v\n", len(opVec.acts), time.Since(startTs))
}

func maxU64(a, b uint64) uint64 {
	if a < b {
		return b
	}
	return a
}

func minU64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
