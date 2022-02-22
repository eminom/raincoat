package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/dbexport"
	"git.enflame.cn/hai.bai/dmaster/efintf"
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
	fMetaStartup = flag.String("meta", "",
		"meta startup folder, if need to do some post-processing meta must be specified")
	fPbMode = flag.Bool("pb", false, "protobuf mode, the latest state-of-art")

	fEnableExTime = flag.Bool("extimeline", false, "enable extended timeline")
	fDiableDma    = flag.Bool("disabledma", false, "disable dma event dispatching")
	fJob          = flag.Int("job", 7, "jobs to go concurrent")

	// if PbMode is enabled:
	fDumpmeta = flag.Bool("dumpmeta", false, "dump meta in pbmode")
	fExec     = flag.Bool("exec", false, "dump from executable")

	//decode go routine count
	fDecodeRoutineCount = flag.Int("subr", 7, "sub process count")
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

func DoProcess(jobCount int, sess *sess.SessBroadcaster,
	algo vgrule.ActMatchAlgo,
	seqId int,
	oneTask bool,
) PostProcessor {
	loader := sess.GetLoader()

	processer := NewPostProcesser(loader, algo, *fEnableExTime, seqId, oneTask)

	startTime := time.Now()
	log.Printf("Starting dispatch events at %v", startTime.Format(time.RFC3339))

	if jobCount <= 0 {
		sess.DispatchToSinkers(
			processer.GetSinkers(
				*fDiableDma,
			)...)
	} else {
		sess.DispatchToConcurSinkers(
			jobCount,
			processer.GetConcurSinkers()...,
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

	var loader efintf.InfoReceiver
	var contentLoader efintf.RingBufferLoader

	var oneTask = (func() bool {
		switch *fArch {
		case "pavo":
			return true
		}
		return false
	})()

	if len(flag.Args()) >= 1 {
		if *fExec {
			topsdev.DumpProfileSectionFromExecutable(flag.Args()[0])
			return
		}

		if *fPbMode {
			inputName := flag.Args()[0]
			pbLoader, err := topsdev.NewPbComplex(inputName, oneTask)
			if err != nil {
				log.Fatalf("error load in pbmode: %v", err)
			}

			if *fDumpmeta {
				pbLoader.DumpMeta()
				pbLoader.DumpRuntimeInformation(inputName)
				return
			}

			// Cast into content-loader
			loader = pbLoader
			contentLoader = &pbLoader
		} else {
			metaStartup := *fMetaStartup
			loader = infoloader.NewMetaFileLoader(metaStartup, oneTask)
			contentLoader = infoloader.NewContentBufferLoader(flag.Args()...)
		}
	}

	decoder := codec.NewDecodeMaster(*fArch)

	// The very ancient way
	if len(flag.Args()) == 0 {
		TextProcess(decoder)
		return
	}

	if *fDump {
		if *fPbMode {
			// only decode the very first one
			cidToDecode := 0
			chunk := contentLoader.LoadRingBufferContent(cidToDecode, 0)
			BinaryProcess(chunk, decoder, *fDecodeRoutineCount)
		} else {
			// single raw file
			filename := flag.Args()[0]
			if chunk, err := os.ReadFile(filename); err == nil {
				BinaryProcess(chunk, decoder, *fDecodeRoutineCount)
			} else {
				panic(fmt.Errorf("could not read %v: %v", filename, err))
			}
		}
		return
	}

	curAlgo := vgrule.NewDoradoRule(decoder)

	// Start concurrency
	rbCount := contentLoader.GetRingBufferCount()
	resChan := make(chan PostProcessor, rbCount)
	var wg sync.WaitGroup
	perCardProcess := func(fileIdx int, outputChan chan<- PostProcessor) {
		defer wg.Done()
		cidToDecode := 0
		chunk := contentLoader.LoadRingBufferContent(cidToDecode, fileIdx)
		sess := sess.NewSessBroadcaster(loader)
		sess.DecodeChunk(chunk, decoder, oneTask, *fDecodeRoutineCount)
		outputChan <- DoProcess(*fJob, sess, curAlgo, fileIdx, oneTask)
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
	for i := 0; i < rbCount; i++ {
		ps[i].DumpToDb(coord, dbObj)
		coord.DeviceID++
	}

	fmt.Printf("dumped to %v\n", outputVpd)
}
