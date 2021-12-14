package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/codec"
)

var (
	fDebug      = flag.Bool("debug", false, "for debug output")
	fArch       = flag.String("arch", "dorado", "hardware arch")
	fDecodeFull = flag.Bool("decodefull", false, "decode all line")
)

var (
	inputErr  = errors.New("input format error")
	decodeErr = errors.New("decode error")
)

func init() {
	flag.Parse()

	switch *fArch {
	case "pavo":
	case "dorado":
	default:
		fmt.Fprintf(os.Stderr, "unknown arch %v\n", *fArch)
		os.Exit(0)
	}
}

func ProcessMasterText(text string, decoder *codec.DecodeMaster) bool {
	if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
		text = text[2:]
	}
	if len(text) <= 0 {
		fmt.Printf("\n")
		return true
	}
	val, err := strconv.ParseUint(text, 16, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parse hex: %v\n", err)
		return false
	}
	engineId, engineType, ctx, ok := decoder.GetEngineInfo(uint32(val))
	if !ok {
		fmt.Printf("decode error for 0x%08x\n", val)
		return false
	}
	fmt.Printf("%08x  %v %v %v", val, engineType, engineId, ctx)
	fmt.Printf("\n")
	return true
}

func toItems(vs []string) []uint32 {
	var arr []uint32
	for _, s := range vs {
		if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
			s = s[2:]
		}
		if len(s) <= 0 {
			fmt.Fprintf(os.Stderr, "not a valid number in hex format")
			return nil
		}
		val, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parse hex: %v\n", err)
			return nil
		}
		arr = append(arr, uint32(val))
	}
	return arr
}

func ProcessFullItem(text string, decoder *codec.DecodeMaster) (error, bool) {
	vs := toItems(strings.Split(text, " "))
	if len(vs) != 4 {
		return inputErr, true
	}

	if vs[0]&1 == 0 {
		// uint32_t flag_ : 1;  // should be always 0
		// uint32_t event_ : 8;
		// uint32_t packet_id_ : 23;
		event := (vs[0] >> 1) & 0xFF
		packet_id := (vs[0] >> 9)
		engIdx, ctx, engTy, ok := decoder.GetEngineInfo(vs[1])
		if !ok {
			return decodeErr, true
		}
		fmt.Printf("%v %v %v event=%v pkt_id=%v\n",
			engTy, engIdx, ctx, event, packet_id)
	} else {
		// uint32_t flag_ : 1;  // should be always 1
		// uint32_t event_ : 7;
		// uint32_t payload_ : 24;
		event := (vs[0] >> 1) & 0x7F
		payload := (vs[0] >> 8)
		engineIdx, engTy, ok := decoder.GetEngineInfoV2(vs[1])
		if !ok {
			return decodeErr, true
		}
		fmt.Printf("%v %v event=%v payload=%v\n",
			engTy, engineIdx, event, payload)
	}

	return nil, true
}

func DecodeDpfItem() {
	reader := bufio.NewReader(os.Stdin)
	decoder := codec.NewDecodeMaster(*fArch)
	errorCounter := 0
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
		if *fDecodeFull {
			err, shallCont := ProcessFullItem(text, decoder)
			if nil != err {
				errorCounter++
			}
			if !shallCont {
				break
			}
		} else {
			if !ProcessMasterText(text, decoder) {
				break
			}
		}
	}

	if errorCounter > 0 {
		fmt.Fprintf(os.Stderr, "error decode %v\n", errorCounter)
	}
}

func main() {

	DecodeDpfItem()
}
