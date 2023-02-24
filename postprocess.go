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
	"git.enflame.cn/hai.bai/dmaster/topsdev/mimic/mimicdefs"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type PostProcessOpt struct {
	OneTask       bool
	PgMaskEncoded bool
	DumpSipBusy   bool
	SeqIdx        int
	NoSubop       bool
	DumpOpDebug   bool
	CpuOps        []rtdata.CpuOpAct
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
	hostInfo   mimicdefs.HostInfo
}

type DumpOpt struct {
	CpuOp bool
}

func NewPostProcesser(loader efintf.InfoReceiver,
	curAlgo vgrule.ActMatchAlgo,
	enableExtendedTimeline bool,
	ppOpt PostProcessOpt,
) PostProcessor {
	rtDict := rtinfo.NewRuntimeTaskManager(ppOpt.OneTask, ppOpt.PgMaskEncoded)
	rtDict.LoadRuntimeTask(loader)

	qm := rtdata.NewOpEventQueue(rtdata.NewOpActCollector(curAlgo),
		codec.DbgPktDetector{},
	)
	fwVec := rtdata.NewOpEventQueue(rtdata.NewFwActCollector(curAlgo),
		codec.FwPktDetector{},
	)
	taskVec := rtdata.NewOpEventQueue(rtdata.NewTaskActCollector(curAlgo),
		codec.TaskDetector{})
	dmaDetector := codec.DmaDetector{}
	dmaVec := rtdata.NewOpEventQueue(rtdata.NewDmaCollector(curAlgo),
		&dmaDetector,
	)
	kernelVec := rtdata.NewOpEventQueue(rtdata.NewKernelActCollector(curAlgo),
		codec.SipDetector{},
	)
	tm := rtinfo.NewTimelineManager(
		rtinfo.TimeLineManagerOpt{
			EnableExtendedTimeline: enableExtendedTimeline,
		})
	tm.LoadTimepoints(loader)

	var hostInfo mimicdefs.HostInfo
	if hi := loader.ExtractHostInfo(); hi != nil {
		hostInfo = *hi //copy
	}

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
		hostInfo:  hostInfo,
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
	DumpHostInfo(
		hostInfo mimicdefs.HostInfo,
	)
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
	DumpCpuOpTrace(
		coords rtdata.Coords,
		cpuOps []rtdata.CpuOpAct,
		rowName string,
	)
}

func (p PostProcessor) DumpToDb(coord rtdata.Coords,
	dOpt DumpOpt, dbe DbDumper) {

	dbe.DumpHostInfo(
		p.hostInfo,
	)

	dbe.DumpDtuOps(
		coord,
		p.dtuOps, p.tm,
	)

	if dOpt.CpuOp {
		dbe.DumpCpuOpTrace(
			coord,
			p.procOpt.CpuOps,
			"CPU Op",
		)
	}

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

		var subOps []rtdata.KernelActivity
		if !p.procOpt.NoSubop {
			subOps = p.rtDict.GenerateKernelActs(
				p.kernelVec.KernelActivity(),
				p.qm.OpActivity(),
				p.curAlgo)
		}

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

		if !p.procOpt.OneTask {
			// Only valid for vg mode
			// There is no true One-Task session.
			// There is only session with a bundle of task without host trace
			rtinfo.GenerateBriefOpsStat(
				p.rtDict.FindExecFor,
				dtuOps,
				p.taskActMap,
				p.rtDict.GetOrderedTaskVec(),
				p.rtDict.CopyTaskVec(),
			)
		}

		if p.procOpt.DumpOpDebug {
			DumpOpsToPythonDebugCode(dtuOps)
		}

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

		if evtFilt := p.dmaVec.GetEventFilter(); evtFilt != nil {
			if dmaDetector, ok := evtFilt.(*codec.DmaDetector); ok {
				ignoreCount := dmaDetector.IdmaPrefetchCount
				fmt.Printf("DMA events up to %v are safely ignored\n", ignoreCount)
			}
		}
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
