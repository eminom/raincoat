package infoloader

import "testing"

func TestXsplit(t *testing.T) {
	vs := XSplit("engine sip launch", 2)
	if vs[1] != "sip launch" {
		t.Fail()
	}
}
