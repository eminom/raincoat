package infoloader

import (
	"testing"

	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
)

func checkT(t *testing.T, dc metadata.DmaInfoMap) {
	for _, dop := range dc.Info {
		switch dop.EngineTy {
		case "CDMA", "SDMA", "unk", "sip launch":
		default:
			t.Logf("error unexpected engine type:<%s> <%s>",
				dop.EngineTy, dop.ToString())
			t.Fail()
		}
	}
}

func TestLoadDma(t *testing.T) {
	var loader DmaOpFormatFetcher
	var dc metadata.DmaInfoMap

	loader = NewCompatibleDmaInfoLoader()
	dc = loader.FetchDmaOpDict("0x97a7adea_memcpy_meta.dumptxt")
	checkT(t, dc)
	loader = NewPbDmaInfoLoader()
	dc = loader.FetchDmaOpDict("0x5d88c90c_memcpy.pbdumptxt")
	checkT(t, dc)
}

func TestLoadMetaFromPavo(t *testing.T) {
	loader := NewCompatibleDmaInfoLoader()
	loader.FetchDmaOpDict("0xf7a303ff_memcpy_meta.dumptxt")
}

func TestLoadSingleLine(t *testing.T) {
	raw := "8388623 dtu.launch_cpu sip launch -1 !dtu.tensor<32x1x5x5xf32:L3> !dtu.tensor<32x1x5x5xf32:L3>"
	pktId, dmaMeta, err := parseSingleLineV0(raw)
	if err != nil {
		t.Logf("Error parsing single line: %v", raw)
		t.Fail()
	}
	if pktId != 8388623 {
		t.Log("error parsing packet id")
		t.Fail()
	}
	if dmaMeta.DmaOpString != "dtu.launch_cpu" {
		t.Fail()
	}
}
