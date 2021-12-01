//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	fDebug = flag.Bool("debug", false, "for debug output")
	fArch  = flag.String("arch", "dorado", "hardware arch")
)

func init() {
	flag.Parse()
}

/*

typedef struct {
  int32_t cluster_id;
  int32_t master_id_hi;
  int32_t master_id_lo;
  int32_t engine_id;
  DPF_ENGINE_T engine_type;
} dpf_id_t;
*/

type EngineType int

const (
	ENGINE_SIP       = "SIP"
	ENGINE_SDMA      = "SDMA"
	ENGINE_CDMA      = "CDMA"
	ENGINE_CQM       = "CQM"
	ENGINE_GSYNC     = "GSYNC"
	ENGINE_SIP_LITE  = "SIP_LITE"
	ENGINE_SDMA_LITE = "SDMA_LITE"
	ENGINE_CDMA_LITE = "ENGINE_CDMA_LITE"
	ENGINE_PCIE      = "PCIE"
	ENGINE_TS        = "TS"
	ENGINE_ODMA      = "ODMA"
	ENGINE_HCVG      = "HCVG"
	ENGINE_VDEC      = "VDEC"
	ENGINE_UNKNOWN   = "UNKNOWN"
)

type DpfEngineT struct {
	ClusterID int
	MasterHi  int
	MasterLo  int
	EngineId  int
	EnTy      string
}

var (
	doradoDpfTy []DpfEngineT
	pavoDpfTy   []DpfEngineT
)

func init() {
	doradoDpfTy = []DpfEngineT{
		{0, 0, 0, 0, ENGINE_SIP},
		{0, 0, 1, 0, ENGINE_SDMA},
		{0, 0, 2, 1, ENGINE_SIP},
		{0, 0, 3, 1, ENGINE_SDMA},
		{0, 0, 4, 2, ENGINE_SIP},
		{0, 0, 5, 2, ENGINE_SDMA},
		{0, 0, 6, 3, ENGINE_SIP},
		{0, 0, 7, 3, ENGINE_SDMA},
		{0, 0, 8, 0, ENGINE_CQM},
		{0, 0, 13, 0, ENGINE_GSYNC},
		{0, 1, 0, 4, ENGINE_SIP},
		{0, 1, 1, 4, ENGINE_SDMA},
		{0, 1, 2, 5, ENGINE_SIP},
		{0, 1, 3, 5, ENGINE_SDMA},
		{0, 1, 4, 6, ENGINE_SIP},
		{0, 1, 5, 6, ENGINE_SDMA},
		{0, 1, 6, 7, ENGINE_SIP},
		{0, 1, 7, 7, ENGINE_SDMA},
		{0, 1, 8, 1, ENGINE_CQM},
		{0, 1, 13, 1, ENGINE_GSYNC},
		{0, 2, 0, 8, ENGINE_SIP},
		{0, 2, 1, 8, ENGINE_SDMA},
		{0, 2, 2, 9, ENGINE_SIP},
		{0, 2, 3, 9, ENGINE_SDMA},
		{0, 2, 4, 10, ENGINE_SIP},
		{0, 2, 5, 10, ENGINE_SDMA},
		{0, 2, 6, 11, ENGINE_SIP},
		{0, 2, 7, 11, ENGINE_SDMA},
		{0, 2, 8, 2, ENGINE_CQM},
		{0, 2, 13, 2, ENGINE_GSYNC},

		{0, 3, 0, 0, ENGINE_CDMA},
		{0, 4, 0, 1, ENGINE_CDMA},
		{0, 5, 0, 2, ENGINE_CDMA},
		{0, 6, 0, 3, ENGINE_CDMA},
		{0, 7, 0, 0, ENGINE_SIP_LITE},
		{0, 7, 1, 0, ENGINE_SDMA_LITE},

		{1, 8, 0, 0, ENGINE_SIP},
		{1, 8, 1, 0, ENGINE_SDMA},
		{1, 8, 2, 1, ENGINE_SIP},
		{1, 8, 3, 1, ENGINE_SDMA},
		{1, 8, 4, 2, ENGINE_SIP},
		{1, 8, 5, 2, ENGINE_SDMA},
		{1, 8, 6, 3, ENGINE_SIP},
		{1, 8, 7, 3, ENGINE_SDMA},
		{1, 8, 8, 0, ENGINE_CQM},
		{1, 8, 13, 0, ENGINE_GSYNC},
		{1, 9, 0, 4, ENGINE_SIP},
		{1, 9, 1, 4, ENGINE_SDMA},
		{1, 9, 2, 5, ENGINE_SIP},
		{1, 9, 3, 5, ENGINE_SDMA},
		{1, 9, 4, 6, ENGINE_SIP},
		{1, 9, 5, 6, ENGINE_SDMA},
		{1, 9, 6, 7, ENGINE_SIP},
		{1, 9, 7, 7, ENGINE_SDMA},
		{1, 9, 8, 1, ENGINE_CQM},
		{1, 9, 13, 1, ENGINE_GSYNC},
		{1, 10, 0, 8, ENGINE_SIP},
		{1, 10, 1, 8, ENGINE_SDMA},
		{1, 10, 2, 9, ENGINE_SIP},
		{1, 10, 3, 9, ENGINE_SDMA},
		{1, 10, 4, 10, ENGINE_SIP},
		{1, 10, 5, 10, ENGINE_SDMA},
		{1, 10, 6, 11, ENGINE_SIP},
		{1, 10, 7, 11, ENGINE_SDMA},
		{1, 10, 8, 2, ENGINE_CQM},
		{1, 10, 13, 2, ENGINE_GSYNC},

		{1, 11, 0, 0, ENGINE_CDMA},
		{1, 12, 0, 1, ENGINE_CDMA},
		{1, 13, 0, 2, ENGINE_CDMA},
		{1, 14, 0, 3, ENGINE_CDMA},
		{1, 15, 0, 0, ENGINE_SIP_LITE},
		{1, 15, 1, 0, ENGINE_SDMA_LITE},

		{2, 22, 0, 0, ENGINE_PCIE},
		{2, 24, 0, 0, ENGINE_TS},
		{2, 25, 0, 0, ENGINE_ODMA},
		{2, 27, 0, 0, ENGINE_HCVG},
		{2, 28, 0, 1, ENGINE_HCVG},
		{2, 29, 0, 0, ENGINE_VDEC},

		// make sure this is the last one
		{-1, -1, -1, -1, ENGINE_UNKNOWN},
	}

	pavoDpfTy = []DpfEngineT{

		/* cid,	mid_hi,	mid_lo,	eid,	engine_type */
		{0, 0, 0, 0, ENGINE_SIP},
		{0, 0, 1, 0, ENGINE_SDMA},
		{0, 0, 2, 1, ENGINE_SIP},
		{0, 0, 3, 1, ENGINE_SDMA},
		{0, 0, 4, 2, ENGINE_SIP},
		{0, 0, 5, 2, ENGINE_SDMA},
		{0, 0, 6, 3, ENGINE_SIP},
		{0, 0, 7, 3, ENGINE_SDMA},
		{0, 0, 8, 4, ENGINE_SIP},
		{0, 0, 9, 4, ENGINE_SDMA},
		{0, 0, 10, 5, ENGINE_SIP},
		{0, 0, 11, 5, ENGINE_SDMA},
		{0, 0, 12, 6, ENGINE_SIP},
		{0, 0, 13, 6, ENGINE_SDMA},
		{0, 0, 14, 0, ENGINE_CDMA},
		{0, 0, 15, 1, ENGINE_CDMA},
		{0, 0, 16, 2, ENGINE_CDMA},
		{0, 0, 17, 3, ENGINE_CDMA},
		{0, 0, 18, 0, ENGINE_CQM},
		{0, 0, 22, 0, ENGINE_GSYNC},

		{1, 1, 0, 0, ENGINE_SIP},
		{1, 1, 1, 0, ENGINE_SDMA},
		{1, 1, 2, 1, ENGINE_SIP},
		{1, 1, 3, 1, ENGINE_SDMA},
		{1, 1, 4, 2, ENGINE_SIP},
		{1, 1, 5, 2, ENGINE_SDMA},
		{1, 1, 6, 3, ENGINE_SIP},
		{1, 1, 7, 3, ENGINE_SDMA},
		{1, 1, 8, 4, ENGINE_SIP},
		{1, 1, 9, 4, ENGINE_SDMA},
		{1, 1, 10, 5, ENGINE_SIP},
		{1, 1, 11, 5, ENGINE_SDMA},
		{1, 1, 12, 6, ENGINE_SIP},
		{1, 1, 13, 6, ENGINE_SDMA},
		{1, 1, 14, 0, ENGINE_CDMA},
		{1, 1, 15, 1, ENGINE_CDMA},
		{1, 1, 16, 2, ENGINE_CDMA},
		{1, 1, 17, 3, ENGINE_CDMA},
		{1, 1, 18, 0, ENGINE_CQM},
		{1, 1, 22, 0, ENGINE_GSYNC},

		{2, 2, 0, 0, ENGINE_SIP},
		{2, 2, 1, 0, ENGINE_SDMA},
		{2, 2, 2, 1, ENGINE_SIP},
		{2, 2, 3, 1, ENGINE_SDMA},
		{2, 2, 4, 2, ENGINE_SIP},
		{2, 2, 5, 2, ENGINE_SDMA},
		{2, 2, 6, 3, ENGINE_SIP},
		{2, 2, 7, 3, ENGINE_SDMA},
		{2, 2, 8, 4, ENGINE_SIP},
		{2, 2, 9, 4, ENGINE_SDMA},
		{2, 2, 10, 5, ENGINE_SIP},
		{2, 2, 11, 5, ENGINE_SDMA},
		{2, 2, 12, 6, ENGINE_SIP},
		{2, 2, 13, 6, ENGINE_SDMA},
		{2, 2, 14, 0, ENGINE_CDMA},
		{2, 2, 15, 1, ENGINE_CDMA},
		{2, 2, 16, 2, ENGINE_CDMA},
		{2, 2, 17, 3, ENGINE_CDMA},
		{2, 2, 18, 0, ENGINE_CQM},
		{2, 2, 22, 0, ENGINE_GSYNC},

		{3, 3, 0, 0, ENGINE_SIP},
		{3, 3, 1, 0, ENGINE_SDMA},
		{3, 3, 2, 1, ENGINE_SIP},
		{3, 3, 3, 1, ENGINE_SDMA},
		{3, 3, 4, 2, ENGINE_SIP},
		{3, 3, 5, 2, ENGINE_SDMA},
		{3, 3, 6, 3, ENGINE_SIP},
		{3, 3, 7, 3, ENGINE_SDMA},
		{3, 3, 8, 4, ENGINE_SIP},
		{3, 3, 9, 4, ENGINE_SDMA},
		{3, 3, 10, 5, ENGINE_SIP},
		{3, 3, 11, 5, ENGINE_SDMA},
		{3, 3, 12, 6, ENGINE_SIP},
		{3, 3, 13, 6, ENGINE_SDMA},
		{3, 3, 14, 0, ENGINE_CDMA},
		{3, 3, 15, 1, ENGINE_CDMA},
		{3, 3, 16, 2, ENGINE_CDMA},
		{3, 3, 17, 3, ENGINE_CDMA},
		{3, 3, 18, 0, ENGINE_CQM},
		{3, 3, 22, 0, ENGINE_GSYNC},

		{6, 4, 0, 0, ENGINE_PCIE},
		{6, 6, 0, 0, ENGINE_TS},

		{4, 15, 0, 0, ENGINE_SIP_LITE},
		{4, 15, 1, 0, ENGINE_SDMA_LITE},
		{4, 15, 2, 0, ENGINE_CDMA_LITE},
		{4, 15, 3, 1, ENGINE_CDMA_LITE},

		{5, 16, 0, 0, ENGINE_SIP_LITE},
		{5, 16, 1, 0, ENGINE_SDMA_LITE},
		{5, 16, 2, 0, ENGINE_CDMA_LITE},
		{5, 16, 3, 1, ENGINE_CDMA_LITE},

		// make sure this is the last one
		{-1, -1, -1, -1, ENGINE_UNKNOWN},
	}
}

func getEngInfo(lo, hi int, dpfInfo []DpfEngineT) (int, string) {
	for _, item := range dpfInfo {
		if item.MasterLo == lo && item.MasterHi == hi {
			return item.EngineId, item.EnTy
		}
	}
	return -1, ENGINE_UNKNOWN
}

func getEngInfoDorado(lo, hi int64) (int, string) {
	return getEngInfo(int(lo), int(hi), doradoDpfTy)
}

func getEngInfoPavo(lo, hi int64) (int, string) {
	return getEngInfo(int(lo), int(hi), pavoDpfTy)
}

func main() {
	switch *fArch {
	case "pavo":
	case "dorado":
	default:
		fmt.Fprintf(os.Stderr, "unknown arch %v\n", *fArch)
		os.Exit(0)
	}
	reader := bufio.NewReader(os.Stdin)
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
		if len(text) <= 0 {
			continue
		}
		val, err := strconv.ParseInt(text, 16, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parse hex: %v\n", err)
			break
		}
		lo, hi := val&0x1f, ((val >> 5) & 0x1f)
		if *fDebug {
			fmt.Printf("decode %v\n", text)
			fmt.Printf("lo => %v, hi => %v\n", lo, hi)
		}
		var engineId int
		var engineType string
		switch *fArch {
		case "dorado":
			engineId, engineType = getEngInfoDorado(lo, hi)
		case "pavo":
			engineId, engineType = getEngInfoPavo(lo, hi)
		default:
			break
		}
		fmt.Printf("%08x  %v %v", val, engineType, engineId)
		fmt.Printf("\n")
	}
}
