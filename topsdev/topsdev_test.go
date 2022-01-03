package topsdev

import (
	"testing"
)

const (
	sampleInput = "/home/hai.bai/data18/new6a/20211230141724.16847.topspti.data"
)

func TestBasic(t *testing.T) {
	ph, body, err := DecodeFile(sampleInput)
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

	dumpTask(pb)
	dumpTimepoints(pb)
	dumpDpfringbuffer(pb)
	dumpExecRaw(pb)
	for _, seri := range pb.Dtu.Meta.ExecutableProfileSerialize {
		ParseProfileSection(seri)
	}
	t.Logf("done parse profile section")
}

func TestSize(t *testing.T) {
	doAssertOnProfileSection()
}
