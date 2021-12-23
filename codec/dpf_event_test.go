package codec

import "testing"

func TestPcie(t *testing.T) {
	decoder := NewDecodeMaster("dorado")
	evt, err := decoder.NewDpfEvent([]uint32{4, 0x2c0, 0, 0}, 0)
	if err != nil || evt.EngineTypeCode != EngCat_PCIE {
		t.Log("not as expected")
		t.Fail()
	}
	t.Logf("EVENT: %v", evt.ToString())
	t.Logf("RAW: %v", evt.RawRepr())
}

func TestCqm(t *testing.T) {
	decoder := NewDecodeMaster("dorado")
	evt, err := decoder.NewDpfEvent([]uint32{
		0x01329c0e, 0x00005148, 0xcb9d9669, 0x00000013}, 0)
	if err != nil || evt.EngineTypeCode != EngCat_CQM {
		t.Log("not as expected")
		t.Fail()
	}
	t.Logf("EVENT: %v", evt.ToString())
	t.Logf("RAW: %v", evt.RawRepr())
}
