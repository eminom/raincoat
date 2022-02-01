package metadata

import (
	"fmt"
	"os"
	"sort"
)

type DmaOp struct {
	PktId       int
	DmaOpString string
	EngineTy    string
	EngineIndex int
	Input       string
	Output      string
	Attrs       map[string]string
}

func (d DmaOp) ToString() string {
	return fmt.Sprintf("%v; %v; %v", d.PktId, d.DmaOpString, d.EngineIndex)
}

type DmaInfoMap struct {
	Info map[int]DmaOp
}

func (d DmaInfoMap) DumpToOstream(fout *os.File) {
	var pktIdVec []int
	for pktId := range d.Info {
		pktIdVec = append(pktIdVec, pktId)
	}
	sort.Ints(pktIdVec)
	for _, pktId := range pktIdVec {
		dmaOp := d.Info[pktId]
		// packet id, module id, engine id. emm...
		fmt.Fprintf(fout, "%v %v %v %v\n", pktId, 0, 0, dmaOp.DmaOpString)
		for k, v := range dmaOp.Attrs {
			fmt.Fprintf(fout, "  %v %v\n", k, v)
		}
	}
}
