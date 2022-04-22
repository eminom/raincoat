package archdetect

import (
	"git.enflame.cn/hai.bai/dmaster/efintf"
)

type ArchDetector struct {
	arch         string
	forceOneTask bool
	getter       efintf.ArchTypeGet
}

func NewArchDetector(arch string,
	forceOneTask bool,
	getter efintf.ArchTypeGet) ArchDetector {
	return ArchDetector{
		arch,
		forceOneTask,
		getter,
	}
}

// One of the valid arch
func (ad ArchDetector) GetArch() string {
	switch ad.arch {
	case "auto", "":
		return ad.getter.GetArchType().ToString()
	case "pavo", "dorado":
		return ad.arch
	}
	return "dorado"
}

// OneTask: strictly
// Only pavo can be one-tasked
func (ad ArchDetector) GetOneTaskFlag() bool {
	return ad.GetArch() == "pavo" && ad.forceOneTask
}
