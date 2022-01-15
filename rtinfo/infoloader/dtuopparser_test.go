package infoloader

import (
	"testing"
)

func TestDtuopInfoNuevoParser(t *testing.T) {
	loader := newNuevoOpLoader()
	opMap := loader.LoadOpMap("0x5d88c90c_opmeta.pbdumptxt")
	var checkState = map[int]string{
		0: "common_rng_uniform",
		1: "elementwise_fusion_dtu",
		2: "broadcast_in_dim_cpu",
	}
	for opId, name := range checkState {
		dtuOp, ok := opMap[opId]
		if !ok || dtuOp.OpName != name {
			t.Fail()
		}
	}

}

func TestDtuopCompatibleParser(t *testing.T) {
	loader := newCompatibleOpLoader()
	opMap := loader.LoadOpMap("0x07960f33_dtuop.dumptxt")
	if len(opMap) != 1 {
		t.FailNow()
	}
	dtuOp, ok := opMap[10]
	if !ok {
		t.FailNow()
	}
	if dtuOp.OpName != "copy_cpu" {
		t.FailNow()
	}
}
