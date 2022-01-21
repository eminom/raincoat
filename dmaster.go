package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

func DoProcess(jobCount int, sess *sess.SessBroadcaster, coord rtdata.Coords,
	algo vgrule.ActMatchAlgo, dbe DbDumper) {
	loader := sess.GetLoader()

	processer := NewPostProcesser(loader, algo, *fEnableExTime)

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
	processer.DoPostProcessing(coord, dbe)
}

func getOutputName(a string) string {
	return filepath.Base(a) + ".vpd"
}

func main() {

	var loader efintf.InfoReceiver
	var contentLoader efintf.RingBufferLoader

	if len(flag.Args()) >= 1 {
		if *fPbMode {
			pbLoader, err := topsdev.NewPbComplex(flag.Args()[0])
			if err != nil {
				log.Fatalf("error load in pbmode: %v", err)
			}
			// Cast into content-loader
			loader = pbLoader
			contentLoader = &pbLoader
		} else {
			metaStartup := *fMetaStartup
			loader = infoloader.NewMetaFileLoader(metaStartup)
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
		BinaryProcess(flag.Args()[0], decoder)
		return
	}

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
	curAlgo := vgrule.NewDoradoRule(decoder)
	for contentLoader.HasMore() {
		cidToDecode := 0
		chunk := contentLoader.LoadRingBufferContent(cidToDecode)
		sess := sess.NewSessBroadcaster(loader)
		sess.DecodeChunk(chunk, decoder)
		DoProcess(*fJob, sess, coord, curAlgo, dbObj)
		coord.DeviceID++
	}

	fmt.Printf("dumped to %v\n", outputVpd)
}
