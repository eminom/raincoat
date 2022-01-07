package main

import (
	"fmt"
	"log"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/dbexport"
	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/meta"
	"git.enflame.cn/hai.bai/dmaster/rtinfo"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/sess"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type PostProcessor struct {
	rtDict *rtinfo.RuntimeTaskManager
	qm     *rtdata.OpEventQueue
	fwVec  *rtdata.OpEventQueue
	tm     *rtinfo.TimelineManager

	curAlgo vgrule.ActMatchAlgo
	loader  efintf.InfoReceiver
}

func NewPostProcesser(loader efintf.InfoReceiver) PostProcessor {
	rtDict := rtinfo.NewRuntimeTaskManager()
	rtDict.LoadRuntimeTask(loader)
	curAlgo := vgrule.NewDoradoRule()
	qm := rtdata.NewOpEventQueue(curAlgo,
		codec.DbgPktDetector{},
	)
	fwVec := rtdata.NewOpEventQueue(curAlgo,
		codec.FwPktDetector{},
	)
	tm := rtinfo.NewTimelineManager()
	tm.LoadTimepoints(loader)
	return PostProcessor{
		loader:  loader,
		curAlgo: curAlgo,
		rtDict:  rtDict,
		qm:      qm,
		fwVec:   fwVec,
		tm:      tm,
	}
}

func (p PostProcessor) GetSinkers() []sess.EventSinker {
	return []sess.EventSinker{
		p.rtDict,
		p.qm,
		p.fwVec,
		p.tm,
	}
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
		bundle []rtdata.OpActivity,
		tm *rtinfo.TimelineManager,
		extractor dbexport.ExtractOpInfo,
	)
}

func (p PostProcessor) DoPostProcessing(coord rtdata.Coords, dbe DbDumper) {
	log.Printf("fwVec count: %v", len(p.fwVec.OpActivity()))

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
			func(act rtdata.OpActivity) (bool, string, string) {
				if act.IsOpRefValid() {
					return true,
						act.GetTask().ToShortString(),
						act.GetOp().OpName
				}
				return false, "Unknown Task", "Unk"
			},
			true,
		)
		notWildInCount := 0
		tr.DumpToEventTrace(unProcessed, p.tm,
			func(act rtdata.OpActivity) (bool, string, string) {
				//+ act.GetTask().ToShortString(),
				if act.IsOpRefValid() {
					return true, "Wild In",
						act.GetOp().OpName
				}
				notWildInCount++
				return false, "", ""
			},
			false,
		)
		subSampleCc := 0
		tr.DumpToEventTrace(wildProcess, p.tm,
			func(act rtdata.OpActivity) (bool, string, string) {
				//+ act.GetTask().ToShortString(),
				subSampleCc++
				if subSampleCc%17 == 0 {
					return true, "Wild Out", "some op"
				}
				return false, "", ""
			},
			false,
		)
		fmt.Printf("# notWildInCount: %v\n", notWildInCount)
		fmt.Printf("# uncertained (could not detmerined at all): %v\n",
			len(wildProcess))
		tr.DumpToFile("dtuop_trace.json")

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
			p.fwVec.OpActivity(), p.tm,
			func(act rtdata.OpActivity) (bool, string, string) {
				switch act.Start.EngineTypeCode {
				case codec.EngCat_TS:
					str, _ := rtdata.ToTSEventString(act.Start.Event)
					return true, "", str
				case codec.EngCat_CQM, codec.EngCat_GSYNC:
					str, _ := rtdata.ToCQMEventString(act.Start.Event)
					return true, "", str
				}
				return false, "", ""
			},
		)

	}
}
