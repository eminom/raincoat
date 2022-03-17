package metadata

type DtuOp struct {
	OpName string
	OpId   int

	// Extras
	ModuleId   int
	Kind       string
	FusionKind string
	ModuleName string
	MetaString string
}

type SubOpMeta struct {
	MasterOpId int
	SlaveOpId  int
	Tid        int //Gsync
	SubOpName  string
}

type SubOpMetaVec []SubOpMeta

func (som SubOpMetaVec) Len() int {
	return len(som)
}

func (som SubOpMetaVec) Swap(i, j int) {
	som[i], som[j] = som[j], som[i]
}

func (som SubOpMetaVec) Less(i, j int) bool {
	return som[i].SlaveOpId < som[j].SlaveOpId
}
