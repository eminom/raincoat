package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/codec"
)

var (
	fDebug = flag.Bool("debug", false, "for debug output")
	fArch  = flag.String("arch", "dorado", "hardware arch")
)

func init() {
	flag.Parse()
}

func DecodeMaster() {
	reader := bufio.NewReader(os.Stdin)
	decoder := codec.NewDecodeMaster(*fArch)
	for {
		// fmt.Print("-> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			if *fDebug {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			break
		}
		text = strings.TrimSuffix(text, "\n")
		if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
			text = text[2:]
		}
		if len(text) <= 0 {
			fmt.Printf("\n")
			continue
		}
		val, err := strconv.ParseUint(text, 16, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parse hex: %v\n", err)
			break
		}

		engineId, engineType, ctx, ok := decoder.GetEngineInfo(uint32(val))
		if !ok {
			fmt.Printf("decode error for 0x%08x\n", val)
			continue
		}
		fmt.Printf("%08x  %v %v %v", val, engineType, engineId, ctx)
		fmt.Printf("\n")
	}
}

func main() {
	switch *fArch {
	case "pavo":
	case "dorado":
	default:
		fmt.Fprintf(os.Stderr, "unknown arch %v\n", *fArch)
		os.Exit(0)
	}

	DecodeMaster()
}
