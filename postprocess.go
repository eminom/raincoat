package main

import (
	"fmt"
	"log"
	"time"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/dbexport"
	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/efintf/sessintf"
	"git.enflame.cn/hai.bai/dmaster/meta"
	"git.enflame.cn/hai.bai/dmaster/rtinfo"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type PostProcessor struct {
	rtDict *rtinfo.RuntimeTaskManager
	qm     *rtdata.EventQueue
	fwVec  *rtdata.EventQueue
	dmaVec *rtdata.EventQueue
	tm     *rtinfo.TimelineManager

	curAlgo vgrule.ActMatchAlgo
	loader  efintf.InfoReceiver
	seqIdx  int
}

func NewPostProcesser(loader efintf.InfoReceiver,
	curAlgo vgrule.ActMatchAlgo,
	enableExtendedTimeline bool,
	seqIdx int,
	oneTask bool,
) PostProcessor {
	rtDict := rtinfo.NewRuntimeTaskManager(oneTask)
	rtDict.LoadRuntimeTask(loader)
	if oneTask {
		// Preload
		rtDict.LoadMeta(loader)
	}

	qm := rtdata.NewOpEventQueue(rtdata.NewOpActCollector(curAlgo,
		rtdata.OpActCollectorOpt{
			OpIdMapper:    rtDict.GetExecRaw(),
			CacheAndMerge: oneTask,
		}),
		codec.DbgPktDetector{},
	)
	fwVec := rtdata.NewOpEventQueue(rtdata.NewFwActCollector(curAlgo),
		codec.FwPktDetector{},
	)
	dmaVec := rtdata.NewOpEventQueue(rtdata.NewDmaCollector(curAlgo),
		codec.DmaDetector{},
	)
	tm := rtinfo.NewTimelineManager(
		rtinfo.TimeLineManagerOpt{
			EnableExtendedTimeline: enableExtendedTimeline,
		})
	tm.LoadTimepoints(loader)
	return PostProcessor{
		loader:  loader,
		curAlgo: curAlgo,
		rtDict:  rtDict,
		qm:      qm,
		fwVec:   fwVec,
		dmaVec:  dmaVec,
		tm:      tm,
		seqIdx:  seqIdx,
	}
}

func (p PostProcessor) GetSinkers(disableDma bool) []sessintf.EventSinker {
	rv := []sessintf.EventSinker{
		p.rtDict,
		p.qm,
		p.fwVec,
		p.tm,
	}
	if !disableDma {
		rv = append(rv, p.dmaVec)
	}
	return rv
}

func (p PostProcessor) GetConcurSinkers() []sessintf.ConcurEventSinker {
	rv := []sessintf.ConcurEventSinker{
		p.rtDict,
		p.qm,
		p.fwVec,
		p.tm,
		p.dmaVec,
	}
	return rv
}

type DbDumper interface {
	DumpDtuOps(
		coords rtdata.Coords,
		bundle []rtdata.OpActivity,
		tm *rtinfo.TimelineManager,
		extractor dbexport.ExtractOpInfo,
	)
	DumpFwActs(
		coords rtdata.Coords,
		bundle []rtdata.FwActivity,
		tm *rtinfo.TimelineManager,
	)
	DumpDmaActs(
		coords rtdata.Coords,
		bundle []rtdata.DmaActivity,
		tm *rtinfo.TimelineManager,
	)
}

func (p PostProcessor) DumpToDb(coord rtdata.Coords, dbe DbDumper) {
	dbe.DumpDtuOps(
		coord,
		p.qm.OpActivity(), p.tm,
		func(act rtdata.OpActivity) (bool, string, string) {
			if act.IsOpRefValid() {
				return true,
					act.GetTask().ToShortString(),
					act.GetOp().OpName
			}
			return false, "Unknown Task", "Unk"
		},
	)

	dbe.DumpFwActs(
		coord,
		p.fwVec.FwActivity(), p.tm,
	)
	dbe.DumpDmaActs(
		coord,
		p.dmaVec.DmaActivity(), p.tm,
	)

}

func (p PostProcessor) DoPostProcessing() {
	log.Printf("# fwVec count: %v", p.fwVec.ActCount())
	log.Printf("# dmaVec count: %v", p.dmaVec.ActCount())

	p.qm.DumpInfo()
	p.tm.AlignToHostTimeline()
	if p.tm.Verify() {
		log.Printf("timeline aligned verified successfully")
	} else {
		log.Printf("timeline aligned verified error")
		panic("shall stop")
	}
	p.tm.DumpInfo()

	if p.rtDict != nil {
		p.rtDict.LoadMeta(p.loader)
		p.rtDict.BuildOrderInfo()
		p.rtDict.DumpInfo()
		meta.TestExecRaw(p.rtDict.GetExecRaw())

		var tr dbexport.TraceEventSession

		unProcessed := p.rtDict.CookCqm(p.qm.OpActivity(),
			p.curAlgo)
		var wildProcess []rtdata.OpActivity
		if false {
			p.rtDict.OvercookCqm(unProcessed, p.curAlgo)
			wildProcess = p.rtDict.WildCookCqm(unProcessed)
		} else {
			p.rtDict.CookCqmEverSince(unProcessed, p.curAlgo)
		}

		dumpFullCycles(p.qm.OpActivity())

		tr.DumpToEventTrace(p.qm.OpActivity(), p.tm,
			NewNormalDumper(),
			true,
		)
		wildInDumper0 := wildInDumper{}
		tr.DumpToEventTrace(unProcessed, p.tm,
			&wildInDumper0,
			false,
		)
		wildOutDumper0 := wildOutDumper{}
		tr.DumpToEventTrace(wildProcess, p.tm,
			&wildOutDumper0,
			false,
		)
		fmt.Printf("# notWildInCount: %v\n", wildInDumper0.notWildInCount)
		fmt.Printf("# uncertained (could not detmerined at all): %v\n",
			len(wildProcess))
		tr.DumpToFile("dtuop_trace.json")

		startDmaTs := time.Now()
		p.rtDict.CookDma(p.dmaVec.DmaActivity(), p.curAlgo)

		fmt.Printf("dma cook and save to db cost %v\n", time.Since(startDmaTs))

	}
}

func (pp *PostProcessor) Sorts() {
	for _, v := range []rtdata.ActCollector{
		pp.dmaVec,
		pp.qm,
		pp.fwVec,
	} {
		v.DoSort()
	}
}

// Result must be sorted to keep aligned with result in sequential processing
type PostProcessors []PostProcessor

func (ps PostProcessors) Len() int {
	return len(ps)
}

func (ps PostProcessors) Less(i, j int) bool {
	return ps[i].seqIdx < ps[j].seqIdx
}

func (ps PostProcessors) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}
