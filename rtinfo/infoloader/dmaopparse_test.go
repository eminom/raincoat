package infoloader

import "testing"

func checkT(t *testing.T, dc DmaInfoMap) {
	for _, dop := range dc.info {
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
	var dc DmaInfoMap

	loader = NewCompatibleDmaInfoLoader()
	dc = loader.FetchDmaOpDict("0x97a7adea_memcpy_meta.dumptxt")
	checkT(t, dc)
	loader = NewPbDmaInfoLoader()
	dc = loader.FetchDmaOpDict("0x5d88c90c_memcpy.pbdumptxt")
	checkT(t, dc)
}
