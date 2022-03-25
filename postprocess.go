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

type PostProcessOpt struct {
	OneTask     bool
	DumpSipBusy bool
	SeqIdx      int
}

type PostProcessor struct {
	rtDict    *rtinfo.RuntimeTaskManager
	qm        *rtdata.EventQueue
	fwVec     *rtdata.EventQueue
	dmaVec    *rtdata.EventQueue
	taskVec   *rtdata.EventQueue
	kernelVec *rtdata.EventQueue
	tm        *rtinfo.TimelineManager

	curAlgo vgrule.ActMatchAlgo
	loader  efintf.InfoReceiver
	procOpt PostProcessOpt

	// Results
	dtuOps     []rtdata.OpActivity
	subOps     []rtdata.KernelActivity
	taskActMap map[int]rtdata.FwActivity
}

func NewPostProcesser(loader efintf.InfoReceiver,
	curAlgo vgrule.ActMatchAlgo,
	enableExtendedTimeline bool,
	ppOpt PostProcessOpt,
) PostProcessor {
	rtDict := rtinfo.NewRuntimeTaskManager(ppOpt.OneTask)
	rtDict.LoadRuntimeTask(loader)

	qm := rtdata.NewOpEventQueue(rtdata.NewOpActCollector(curAlgo),
		codec.DbgPktDetector{},
	)
	fwVec := rtdata.NewOpEventQueue(rtdata.NewFwActCollector(curAlgo),
		codec.FwPktDetector{},
	)
	taskVec := rtdata.NewOpEventQueue(rtdata.NewTaskActCollector(curAlgo),
		codec.TaskDetector{})
	dmaVec := rtdata.NewOpEventQueue(rtdata.NewDmaCollector(curAlgo),
		codec.DmaDetector{},
	)
	kernelVec := rtdata.NewOpEventQueue(rtdata.NewKernelActCollector(curAlgo),
		codec.SipDetector{},
	)
	tm := rtinfo.NewTimelineManager(
		rtinfo.TimeLineManagerOpt{
			EnableExtendedTimeline: enableExtendedTimeline,
		})
	tm.LoadTimepoints(loader)
	return PostProcessor{
		loader:    loader,
		curAlgo:   curAlgo,
		rtDict:    rtDict,
		qm:        qm,
		fwVec:     fwVec,
		dmaVec:    dmaVec,
		taskVec:   taskVec,
		kernelVec: kernelVec,
		tm:        tm,
		procOpt:   ppOpt,
	}
}

type DpfActionOptions struct {
	NoDma bool
	NoSip bool
}

func (p PostProcessor) GetSinkers(
	dopts DpfActionOptions,
) []sessintf.EventSinker {
	rv := []sessintf.EventSinker{
		p.taskVec,
		p.qm,
		p.fwVec,
		p.tm,
	}
	if !dopts.NoDma {
		rv = append(rv, p.dmaVec)
	}
	if !dopts.NoSip {
		rv = append(rv, p.kernelVec)
	}
	return rv
}

func (p PostProcessor) GetConcurSinkers(
	dopts DpfActionOptions,
) []sessintf.ConcurEventSinker {
	rv := []sessintf.ConcurEventSinker{
		p.taskVec,
		p.qm,
		p.fwVec,
		p.tm,
	}
	if !dopts.NoDma {
		rv = append(rv, p.dmaVec)
	}
	if !dopts.NoSip {
		rv = append(rv, p.kernelVec)
	}
	return rv
}

type DbDumper interface {
	DumpDtuOps(
		coords rtdata.Coords,
		bundle []rtdata.OpActivity,
		tm *rtinfo.TimelineManager,
	)
	DumpTaskVec(
		coords rtdata.Coords,
		taskVec []rtdata.OrderTask,
		taskActMap map[int]rtdata.FwActivity,
		tm *rtinfo.TimelineManager,
	)
	DumpFwActs(
		coords rtdata.Coords,
		bundle []rtdata.FwActivity,
		taskActMap map[int]rtdata.FwActivity,
		tm *rtinfo.TimelineManager,
	)
	DumpDmaActs(
		coords rtdata.Coords,
		bundle []rtdata.DmaActivity,
		tm *rtinfo.TimelineManager,
	)
	DumpKernelActs(
		coords rtdata.Coords,
		bundle []rtdata.KernelActivity,
		tm *rtinfo.TimelineManager,
		rowName string,
	)
}

func (p PostProcessor) DumpToDb(coord rtdata.Coords, dbe DbDumper) {
	dbe.DumpDtuOps(
		coord,
		p.dtuOps, p.tm,
	)

	dbe.DumpTaskVec(
		coord,
		p.rtDict.GetOrderedTaskVec(),
		p.taskActMap, // Task id to Cqm Exec map
		p.tm,
	)
	dbe.DumpFwActs(
		coord,
		p.fwVec.FwActivity(),
		p.taskActMap, p.tm,
	)
	dbe.DumpDmaActs(
		coord,
		p.dmaVec.DmaActivity(), p.tm,
	)
	dbe.DumpKernelActs(
		coord,
		p.subOps, p.tm,
		"Sub Ops",
	)
	if p.procOpt.DumpSipBusy {
		dbe.DumpKernelActs(
			coord,
			p.kernelVec.KernelActivity(), p.tm,
			"SIP BUSY",
		)
	}
}

func (p *PostProcessor) DoPostProcessing() {
	log.Printf("# fwVec count: %v", p.fwVec.ActCount())
	log.Printf("# dmaVec count: %v", p.dmaVec.ActCount())
	log.Printf("# task count: %v", p.taskVec.ActCount())

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
		p.rtDict.ProcessTaskActVector(p.taskVec.TaskActivity())
		p.rtDict.LoadMeta(p.loader)
		p.rtDict.BuildOrderInfo()
		p.rtDict.DumpInfo()
		meta.TestExecRaw(p.rtDict.GetExecRaw())

		var tr dbexport.TraceEventSession

		// Generate op runtime info
		dtuOps, unProcessed := p.rtDict.GenerateDtuOps(p.qm.OpActivity(),
			p.curAlgo)
		// Process kernel activities(SIPs)
		subOps := p.rtDict.GenerateKernelActs(
			p.kernelVec.KernelActivity(),
			p.qm.OpActivity(),
			p.curAlgo)

		// Generate task timeline(depends on OPs information)
		p.taskActMap = p.rtDict.GenerateTaskOps(
			p.fwVec.FwActivity(),
			p.fwVec.GetDebugEventVec(),
			dtuOps,
			p.curAlgo)

		p.dtuOps = dtuOps
		p.subOps = subOps
		var wildProcess []rtdata.OpActivity
		if false {
			p.rtDict.OvercookCqm(unProcessed, p.curAlgo)
			wildProcess = p.rtDict.WildCookCqm(unProcessed)
		} else {
			p.rtDict.CookCqmEverSince(unProcessed, p.curAlgo)
		}

		dumpFullCycles(dtuOps)
		rtinfo.GenerateBriefOpsStat(
			p.rtDict.FindExecFor,
			dtuOps,
			p.taskActMap,
			p.rtDict.GetOrderedTaskVec(),
			p.rtDict.CopyTaskVec(),
		)

		tr.DumpTaskVec(p.taskActMap,
			p.rtDict.GetOrderedTaskVec(),
			p.tm,
		)

		tr.DumpToEventTrace(dtuOps, p.tm,
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
		pp.kernelVec,
		pp.taskVec,
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
	return ps[i].procOpt.SeqIdx < ps[j].procOpt.SeqIdx
}

func (ps PostProcessors) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}
