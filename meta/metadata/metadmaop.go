package metadata

type DmaOp struct {
	PktId       int
	DmaOpString string
	EngineTy    string
	EngineIndex int
	Input       string
	Output      string
	Attrs       map[string]string
}
