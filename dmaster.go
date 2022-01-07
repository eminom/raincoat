package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/infoloader"
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
	fProc        = flag.Bool("proc", false, "post-processing")
	fMetaStartup = flag.String("meta", "",
		"meta startup folder, if need to do some post-processing meta must be specified")
	fPbMode = flag.Bool("pb", false, "protobuf mode, the latest state-of-art")
)

func init() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)
	if *fPbMode {
		//Derives
		*fProc = true
	}

	if len(*fMetaStartup) > 0 {
		*fProc = true
	}

	switch *fArch {
	case "pavo":
	case "dorado":
	default:
		fmt.Fprintf(os.Stderr, "unknown arch %v\n", *fArch)
		os.Exit(1)
	}
}

func DoProcess(sess *sess.SessBroadcaster) {

	loader := sess.GetLoader()
	processer := NewPostProcesser(loader)
	sess.DispatchToSinkers(processer.GetSinkers()...)
	processer.DoPostProcessing()
}

func main() {

	var loader efintf.InfoReceiver

	if len(flag.Args()) >= 1 {
		if *fPbMode {
			var err error
			loader, err = topsdev.NewPbLoader(flag.Args()[0])
			if err != nil {
				log.Fatalf("error load in pbmode: %v", err)
			}
		} else {
			metaStartup := *fMetaStartup
			loader = infoloader.NewMetaFileLoader(metaStartup, flag.Args()[0])
		}
	}

	sess := sess.NewSessBroadcaster(sess.SessionOpt{
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
