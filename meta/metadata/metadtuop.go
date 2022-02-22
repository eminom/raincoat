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
