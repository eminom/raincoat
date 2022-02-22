package topsdev

import (
	"os"
	"testing"
)

const (
	sampleInput = "/home/hai.bai/data_dorado/0222/112607_3487.topspti.rawdata.data"
)

func TestBasicTopsdev(t *testing.T) {
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
		ParseProfileSection(seri, os.Stdout)
	}
	t.Logf("done parse profile section")
}

func TestSize(t *testing.T) {
	doAssertOnProfileSection()
}
