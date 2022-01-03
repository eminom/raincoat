package topsdev

import "testing"

func TestOne(t *testing.T) {
	t.Logf("size of header: %v", HeaderSize())
	ph, body, err := DecodeFile(
		"/home/hai.bai/data18/new6a/20211230141724.16847.topspti.data")
	if err != nil {
		t.Logf("error: %v", err)
		t.FailNow()
	}
	t.Logf("%v", ph.ToString())
	t.Logf("body size is %v", len(body))

	pb, err := ParseFromChunk(body)
	if err != nil {
		t.Logf("error parse from istream: %v", err)
		t.FailNow()
	}
	_ = pb

	dumpTask(pb)
	dumpTimepoints(pb)
	dumpDpfringbuffer(pb)
}
