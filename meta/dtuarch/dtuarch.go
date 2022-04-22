package dtuarch

type ArchType int

const (
	EnflameT20 ArchType = iota + 20
	EnflameI20
	EnflameUnknownArch ArchType = 404
)

const (
	DoradoNameTrait = "dorado"
	PavoNameTrait   = "pavo"
)

func (a ArchType) ToString() string {
	switch a {
	case EnflameT20:
		return PavoNameTrait
	case EnflameI20:
		return DoradoNameTrait
	}
	// default to dorado
	return DoradoNameTrait
}
