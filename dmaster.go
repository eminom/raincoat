package main

import (
	"flag"
	"fmt"
	"os"

	"git.enflame.cn/hai.bai/dmaster/algo"
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/rtinfo"
	"git.enflame.cn/hai.bai/dmaster/sess"
)

var (
	fDebug      = flag.Bool("debug", false, "for debug output")
	fArch       = flag.String("arch", "dorado", "hardware arch")
	fDecodeFull = flag.Bool("decodefull", false, "decode all line")
	fSort       = flag.Bool("sort", false, "sort by order")
	fEng        = flag.String("eng", "", "engine to filter in")
	fDump       = flag.Bool("dump", false, "decode file and dump to stdout")
	fRaw        = flag.Bool("raw", false, "dump raw value")
	fProc       = flag.Bool("proc", false, "post-processing")
)

func init() {
	flag.Parse()

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

	rtDict := rtinfo.LoadRuntimeTask("runtime_task.txt")
	qm := rtinfo.NewCqmEventQueue(algo.NewAlgo1())
	doFunc := func(evt codec.DpfEvent) {
		allCount++
		switch evt.EngineTypeCode {
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

	if rtDict != nil {
		rtDict.LoadMeta("/home/hai.bai/data16/meta")
		rtDict.DumpInfo()
		// rtDict.CookCqm()

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
