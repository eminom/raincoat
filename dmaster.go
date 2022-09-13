package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/dbexport"
	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/inspector"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/archdetect"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/infoloader"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/sess"
	"git.enflame.cn/hai.bai/dmaster/topsdev"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

var (
	fDebug      = flag.Bool("debug", false, "for debug output")
	fArch       = flag.String("arch", "auto", "hardware arch")
	fDecodeFull = flag.Bool("decodefull", false, "decode all line")
	fSort       = flag.Bool("sort", false, "sort by order")
	fEng        = flag.String("eng", "", "engine to filter in")

	fDump = flag.Bool("dump", false, "decode file and dump to stdout")
	fRaw  = flag.Bool("raw", true,
		"dump raw value\n"+
			"if -dump is set, dmaster is going to dump the original value from ring buffer\n",
	)
	fMetaStartup = flag.String("meta", "",
		"meta startup folder, if need to do some post-processing meta must be specified")
	fRawDpf = flag.Bool("rawdpf", false, "raw dpf buffer content from file")

	fEnableExTime = flag.Bool("extimeline", false, "enable extended timeline")
	fWithoutDma   = flag.Bool("nodma", false, "disable dma event parsing")
	fWithoutSip   = flag.Bool("nosip", false, "disable sip event parsing")
	fJob          = flag.Int("job", 7, "jobs to go concurrent")

	// if PbMode is enabled:
	fDumpmeta = flag.Bool("dumpmeta", false, "dump meta in pbmode")
	fExec     = flag.Bool("exec", false, "dump from executable")
	fCheckFlg = flag.Bool("checkf", false, "check format only")

	//decode go routine count
	fDecodeRoutineCount = flag.Int("subr", 7, "sub process count")

	// DB rendering options
	fSipBusy = flag.Bool("sipbusy", false, "dump sip busy events")
	fNoSubop = flag.Bool("nosubop", false, "no sub op processing")

	// Force onetask
	fForceOneTask = flag.Bool("force1task", false, "force 1task for pavo")
	fDumpToStd    = flag.Bool("stdout", false, "dump to stdout")

	// Dump option control(which ones are imported into vpd)
	fDumpCpuOp = flag.Bool("dcpuop", false, "dump cpu op trace")

	// Dump op time info(cycles) into python format
	fDumpOpDebug = flag.Bool("dumpopdebug", false, "dump op cycle info for debug")
)

// package
var (
	fDoradoRun = flag.Bool("i20", false, "dorado go")
	fPavoRun   = flag.Bool("t20", false, "pavo go")
)

func init() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	// Override
	if *fDoradoRun && *fPavoRun {
		log.Fatal("error flag t20 i20 shall not meet")
	}
	if *fDoradoRun {
		*fArch = "dorado"
		*fDump = true
	}
	if *fPavoRun {
		*fArch = "pavo"
		*fDump = true
	}

	switch *fArch {
	case "auto":
	case "pavo":
	case "dorado":
	default:
		fmt.Fprintf(os.Stderr, "unknown arch %v\n", *fArch)
		os.Exit(1)
	}
}

func DoProcess(jobCount int, sess *sess.SessBroadcaster,
	algo vgrule.ActMatchAlgo,
	ppOpt PostProcessOpt,
) PostProcessor {
	loader := sess.GetLoader()

	processer := NewPostProcesser(loader, algo,
		*fEnableExTime,
		ppOpt,
	)

	startTime := time.Now()
	log.Printf("Starting dispatch events at %v", startTime.Format(time.RFC3339))

	dopts := DpfActionOptions{
		NoDma: *fWithoutDma,
		NoSip: *fWithoutSip,
	}

	if jobCount <= 0 {
		sess.DispatchToSinkers(
			processer.GetSinkers(dopts)...)
	} else {
		sess.DispatchToConcurSinkers(
			jobCount,
			processer.GetConcurSinkers(dopts)...,
		)
	}

	// All sorts
	processer.Sorts()

	endTime := time.Now()
	log.Printf("Done dispatch events at %v", endTime.Format(time.RFC3339))

	durationTime := endTime.Sub(startTime)
	log.Printf("dispatching cost %v", durationTime)
	processer.DoPostProcessing()
	return processer
}

func getOutputName(a string) string {
	return filepath.Base(a) + ".vpd"
}

func main() {

	if len(flag.Args()) > 0 && strings.HasSuffix(flag.Args()[0], ".vpd") {
		inspector.InspectMain(flag.Args(), os.Stdout)
		return
	}

	var loader efintf.InfoReceiver
	var contentLoader efintf.RingBufferLoader

	if len(*fMetaStartup) == 0 {
		if cwd, err := os.Getwd(); err == nil {
			if abspath, err := filepath.Abs(cwd); err == nil {
				*fMetaStartup = abspath
			}
		}
	}

	if len(flag.Args()) >= 1 {
		if *fExec && *fDumpmeta {
			for _, inputFile := range flag.Args() {
				topsdev.DumpSectionsFromExecutable(inputFile, *fCheckFlg)
			}
			return
		}
	}

	if len(flag.Args()) >= 1 {

		if !*fRawDpf {
			inputName := flag.Args()[0]
			pbLoader, err := topsdev.NewPbComplex(inputName)
			if err != nil {
				log.Fatalf("error load in pbmode: %v", err)
			}

			if *fDumpmeta {
				pbLoader.DumpMeta()
				pbLoader.DumpRuntimeInformation(inputName)
				pbLoader.DumpCpuOpTrace(inputName)
				pbLoader.DumpMisc(inputName)
				fmt.Fprintf(os.Stderr, "dumpmeta done\n")
				return
			}

			// Cast into content-loader
			loader = pbLoader
			contentLoader = &pbLoader
		} else if *fExec {
			// Positional arguments:
			// [0] rawdpf file as input
			// [1:] executable files
			files := flag.Args()
			if len(files) < 2 {
				fmt.Fprintf(os.Stderr, "expecting at least one executable-file, and one raw-dpf-file\n")
				return
			}
			loader = topsdev.NewExecLoader(flag.Args()[1:])
			contentLoader = infoloader.NewContentBufferLoader(flag.Args()[0])
		} else {
			metaStartup := *fMetaStartup
			loader = infoloader.NewMetaFileLoader(metaStartup, *fForceOneTask)
			contentLoader = infoloader.NewContentBufferLoader(flag.Args()...)
		}
	}

	archDetector := archdetect.NewArchDetector(*fArch, *fForceOneTask, loader)
	decoder := codec.NewDecodeMaster(archDetector.GetArch())

	// Dumping are now equiped with statistics work
	curAlgo := vgrule.NewDoradoRule(decoder, loader.GetCdmaAffinity())

	// The very ancient way
	if len(flag.Args()) == 0 {
		TextProcess(decoder)
		return
	}

	if *fDump {

		var fout *os.File
		if *fDumpToStd {
			fout = os.Stdout
		} else {
			var err error
			outName := contentLoader.GetInputName() + ".evtseq"
			fout, err = os.Create(outName)
			if err != nil {
				panic(err)
			}
			defer (func() {
				fout.Close()
				log.Printf("event sequence dump to %v\n", outName)
				fmt.Printf("%v\n", outName)
			})()
		}

		if !*fRawDpf {
			// only decode the very first one
			cidToDecode := 0
			chunk := contentLoader.LoadRingBufferContent(cidToDecode, 0)
			BinaryProcess(chunk, fout, decoder, *fDecodeRoutineCount, curAlgo)
		} else {
			// single raw file
			filename := flag.Args()[0]
			if chunk, err := os.ReadFile(filename); err == nil {
				BinaryProcess(chunk, fout, decoder, *fDecodeRoutineCount, curAlgo)
			} else {
				panic(fmt.Errorf("could not read %v: %v", filename, err))
			}
		}
		return
	}

	// Start concurrency
	rbCount := contentLoader.GetRingBufferCount()
	resChan := make(chan PostProcessor, rbCount)
	var wg sync.WaitGroup
	perCardProcess := func(fileIdx int, outputChan chan<- PostProcessor) {
		defer wg.Done()
		cidToDecode := 0
		chunk := contentLoader.LoadRingBufferContent(cidToDecode, fileIdx)
		sess := sess.NewSessBroadcaster(loader)

		var cpuOps []rtdata.CpuOpAct
		if cpuOpLoader, ok := contentLoader.(efintf.CpuOpTraceLoader); ok {
			cpuOps = cpuOpLoader.GetCpuOpTraceSeq()
		}

		sess.DecodeChunk(chunk, decoder, *fDecodeRoutineCount)
		outputChan <- DoProcess(*fJob, sess, curAlgo, PostProcessOpt{
			OneTask:     archDetector.GetOneTaskFlag(),
			DumpSipBusy: *fSipBusy,
			SeqIdx:      fileIdx,
			NoSubop:     *fNoSubop,
			CpuOps:      cpuOps,
			DumpOpDebug: *fDumpOpDebug,
		})
	}
	for i := 0; i < rbCount; i++ {
		wg.Add(1)
		go perCardProcess(i, resChan)
	}
	wg.Wait()

	// Collect result
	var ps []PostProcessor
	for i := 0; i < rbCount; i++ {
		ps = append(ps, <-resChan)
	}
	sort.Sort(PostProcessors(ps))

	// Dump to DB
	// Use the first input file as the output filename
	outputVpd := getOutputName(flag.Args()[0])
	dbObj, err := dbexport.NewDbSession(outputVpd)
	if err != nil {
		panic(err)
	}
	defer dbObj.Close()

	var coord = rtdata.Coords{
		NodeID:   0,
		DeviceID: 0,
	}
	var dOpt = DumpOpt{
		CpuOp: *fDumpCpuOp,
	}
	for i := 0; i < rbCount; i++ {
		ps[i].DumpToDb(coord, dOpt, dbObj)
		coord.DeviceID++
	}

	fmt.Printf("dumped to %v\n", outputVpd)
}
