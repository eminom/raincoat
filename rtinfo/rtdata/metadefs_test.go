package rtdata

import "testing"

func TestMetadefs(t *testing.T) {
	if decorateName("CQM_LAUNCH") != "Cqm Launch" {
		t.Fail()
	}

	if parseInt32("0x11") != 17 {
		t.Fail()
	}

	if parseInt32("") != 0 {
		t.Fail()
	}

	if parseInt32("1023") != 1023 {
		t.Fail()
	}

}
