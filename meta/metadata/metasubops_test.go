package metadata

import (
	"testing"
)

func TestSubopsLoad(t *testing.T) {
	info := JsonLoader{}.LoadInfo("subops.json")
	if info == nil {
		t.Log("could not load target")
		t.FailNow()
	}
}
