package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/dbexport"
	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/infoloader"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/sess"
	"git.enflame.cn/hai.bai/dmaster/topsdev"
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

func DoProcess(sess *sess.SessBroadcaster, coord rtdata.Coords, dbe DbDumper) {
	loader := sess.GetLoader()
	processer := NewPostProcesser(loader)
	sess.DispatchToSinkers(processer.GetSinkers()...)
	processer.DoPostProcessing(coord, dbe)
}

func getInputName(a string) string {
	return a + ".vpd"
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
	outputVpd := getInputName(flag.Args()[0])
	dbObj, err := dbexport.NewDbSession(outputVpd)
	if err != nil {
		panic(err)
	}
	defer dbObj.Close()

	var coord = rtdata.Coords{
		NodeID:   0,
		DeviceID: 0,
	}
	for contentLoader.HasMore() {
		cidToDecode := 0
		chunk := contentLoader.LoadRingBufferContent(cidToDecode)
		sess := sess.NewSessBroadcaster(loader)
		sess.DecodeChunk(chunk, decoder)
		DoProcess(sess, coord, dbObj)
		coord.DeviceID++
	}

	fmt.Printf("dumped to %v\n", outputVpd)
}
