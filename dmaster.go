package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/dbexport"
	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/meta"
	"git.enflame.cn/hai.bai/dmaster/rtinfo"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/infoloader"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/sess"
	"git.enflame.cn/hai.bai/dmaster/topsdev"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
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
	fMetaStartup = flag.String("meta", "",
		"meta startup folder, if need to do some post-processing meta must be specified")
	fPbMode = flag.Bool("pb", false, "protobuf mode, the latest state-of-art")
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

	rtDict := rtinfo.NewRuntimeTaskManager()
	loader := sess.GetLoader()
	rtDict.LoadRuntimeTask(loader)
	curAlgo := vgrule.NewDoradoRule()
	qm := rtdata.NewOpEventQueue(curAlgo)
	tm := rtinfo.NewTimelineManager()
	tm.LoadTimepoints(loader)
	doFunc := func(evt codec.DpfEvent) {
		allCount++
		switch evt.EngineTypeCode {
		case codec.EngCat_PCIE:
			// For pavo/dorado there is only one kind of PCIE:
			// Sync info
			tm.PutEvent(evt)

		case codec.EngCat_CQM, codec.EngCat_GSYNC:
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
		rtDict.LoadMeta(loader)
		rtDict.BuildOrderInfo()
		rtDict.DumpInfo()
		meta.TestExecRaw(rtDict.GetExecRaw())

		var tr dbexport.TraceEventSession
		unProcessed := rtDict.CookCqm(qm.OpActivity(), curAlgo)
		var wildProcess []rtdata.OpActivity
		if false {
			rtDict.OvercookCqm(unProcessed, curAlgo)
			wildProcess = rtDict.WildCookCqm(unProcessed)
		} else {
			rtDict.CookCqmEverSince(unProcessed, curAlgo)
		}

		dumpFullCycles(qm.OpActivity())

		tr.DumpToEventTrace(qm.OpActivity(), tm,
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
		tr.DumpToEventTrace(unProcessed, tm,
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
		tr.DumpToEventTrace(wildProcess, tm,
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

		outputVpd := "fake.vpd"
		if dbe, err := dbexport.NewDbSession(outputVpd); nil == err {
			defer dbe.Close()
			dbe.DumpToEventTrace(
				qm.OpActivity(), tm,
				func(act rtdata.OpActivity) (bool, string, string) {
					if act.IsOpRefValid() {
						return true,
							act.GetTask().ToShortString(),
							act.GetOp().OpName
					}
					return false, "Unknown Task", "Unk"
				},
				false,
			)
			fmt.Printf("dumped to %v\n", outputVpd)
		}
	}
}

func main() {

	var loader efintf.InfoReceiver

	if len(flag.Args()) >= 1 {
		if *fPbMode {
			var err error
			loader, err = topsdev.NewPbLoader(flag.Args()[0])
			if err != nil {
				log.Fatal("error load in pbmode: %v", err)
			}
		} else {
			metaStartup := *fMetaStartup
			loader = infoloader.NewMetaFileLoader(metaStartup, flag.Args()[0])
		}
	}

	sess := sess.NewSession(sess.SessionOpt{
		Debug:        *fDebug,
		Sort:         *fSort,
		DecodeFull:   *fDecodeFull,
		EngineFilter: *fEng,
		InfoLoader:   loader,
	})
	decoder := codec.NewDecodeMaster(*fArch)
	if len(flag.Args()) > 0 {
		sess.DecodeFromFile(decoder)
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

func dumpFullCycles(bundle []rtdata.OpActivity) {
	var intvs rtdata.Interval
	// Check intervals
	for _, act := range bundle {
		intvs = append(intvs, []uint64{
			act.StartCycle(),
			act.EndCycle(),
		})
	}
	// sort.Sort(intvs)
	overlappedCount := 0
	for i := 0; i < len(intvs)-1; i++ {
		if intvs[i][1] >= intvs[i+1][0] {
			log.Printf("error: %v >= %v at [%d] out of [%v]",
				intvs[i][1], intvs[i+1][0],
				i, len(intvs))
			overlappedCount++
			break
		}
	}
	if overlappedCount != 0 {
		log.Printf("warning: must be no overlap in full dump: but there is %v",
			overlappedCount,
		)
	}
}
