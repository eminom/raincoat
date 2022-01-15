package metadata

import "fmt"

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
