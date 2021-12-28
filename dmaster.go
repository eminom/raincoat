package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"git.enflame.cn/hai.bai/dmaster/algo"
	"git.enflame.cn/hai.bai/dmaster/codec"
	decmconf "git.enflame.cn/hai.bai/dmaster/conf"
	"git.enflame.cn/hai.bai/dmaster/meta"
	"git.enflame.cn/hai.bai/dmaster/rtinfo"
	"git.enflame.cn/hai.bai/dmaster/sess"
)

var (
	fDebug      = flag.Bool("debug", false, "for debug output")
	fArch       = flag.String("arch", "dorado", "hardware arch")
	fDecodeFull = flag.Bool("decodefull", false, "decode all line")
	fSort       = flag.Bool("sort", false, "sort by order")
	fEng        = flag.String("eng", "", "engine to filter in")

	fDump = flag.Bool("dump", false, "decode file and dump to stdout")
	fRaw  = flag.Bool("raw", false,
		"dump raw value\n"+
			"if -dump is set, dmaster is going to dump the original value from ring buffer\n",
	)
	fProc        = flag.Bool("proc", false, "post-processing")
	fMetaStartup = flag.String("meta", "/home/hai.bai/data16/meta",
		"meta startup folder, if need to do some post-processing meta must be specified")
)

func init() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	switch *fArch {
	case "pavo":
	case "dorado":
	default:
		fmt.Fprintf(os.Stderr, "unknown arch %v\n", *fArch)
		os.Exit(1)
	}
}

func DoProcess(sess *sess.Session) {
	cqmOpDbgCount := 0
	allCount := 0

	pathConf := decmconf.NewDecmConf(*fMetaStartup)

	rtDict := rtinfo.NewRuntimeTaskManager()
	rtDict.LoadRuntimeTask(pathConf.GetRuntimeTaskPath())
	qm := rtinfo.NewCqmEventQueue(algo.NewAlgo1())
	tm := rtinfo.NewTimelineManager()
	tm.LoadTimepoints(pathConf.GetTimepointsPath())
	doFunc := func(evt codec.DpfEvent) {
		allCount++
		switch evt.EngineTypeCode {
		case codec.EngCat_PCIE:
			// For pavo/dorado there is only one kind of PCIE:
			// Sync info
			tm.PutEvent(evt)

		case codec.EngCat_CQM:
			if codec.IsCqmOpEvent(evt) {
				if err := qm.PutEvent(evt); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
				}
				cqmOpDbgCount++
			}
		case codec.EngCat_TS:
			if rtDict != nil {
				rtDict.CollectTsEvent(evt)
			}
		}
	}
	sess.EmitForEach(doFunc)
	fmt.Printf("op debug event count %v\n", cqmOpDbgCount)
	fmt.Printf("event %v(0x%x) in all\n", allCount, allCount)
	qm.DumpInfo()
	tm.AlignToHostTimeline()
	if tm.Verify() {
		log.Printf("timeline aligned verified successfully")
	} else {
		log.Printf("timeline aligned verified error")
		panic("shall stop")
	}
	tm.DumpInfo()

	if rtDict != nil {
		rtDict.LoadMeta(pathConf.GetMetaStartupPath())
		rtDict.BuildOrderInfo()
		rtDict.DumpInfo()
		meta.TestExecRaw(rtDict.GetExecRaw())

		var tr rtinfo.TraceEventSession
		unProcessed := rtDict.CookCqm(qm.CqmActBundle())
		rtDict.OvercookCqm(unProcessed)
		wildProcess := rtDict.WildCookCqm(unProcessed)
		tr.DumpToEventTrace(qm.CqmActBundle(), tm,
			func(act rtinfo.CqmActBundle) (bool, string, string) {
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
		tr.DumpToEventTrace(unProcessed, tm,
			func(act rtinfo.CqmActBundle) (bool, string, string) {
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
		tr.DumpToEventTrace(wildProcess, tm,
			func(act rtinfo.CqmActBundle) (bool, string, string) {
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
		fmt.Printf("# subSampleCc: %v\n", subSampleCc)
		tr.DumpToFile("dtuop_trace.json")
	}
}

func main() {
	sess := sess.NewSession(sess.SessionOpt{
		Debug:        *fDebug,
		Sort:         *fSort,
		DecodeFull:   *fDecodeFull,
		EngineFilter: *fEng,
	})
	decoder := codec.NewDecodeMaster(*fArch)
	if len(flag.Args()) > 0 {
		filename := flag.Args()[0]
		sess.DecodeFromFile(filename, decoder)
	} else {
		sess.DecodeFromTextStream(os.Stdin, decoder)
	}
	if *fDump {
		sess.PrintItems(*fRaw)
	}

	if *fProc {
		DoProcess(sess)
	}

	// fmt.Printf("done")
}
