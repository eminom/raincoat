package archdetect

import (
	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/meta/dtuarch"
)

type ArchDetector struct {
	arch         string
	forceOneTask bool
	getter       efintf.ArchTypeGet
}

type deaultGetter struct{}

func (deaultGetter) GetArchType() dtuarch.ArchType {
	return dtuarch.EnflameI20
}

func getOrDefault(getter efintf.ArchTypeGet) efintf.ArchTypeGet {
	if getter != nil {
		return getter
	}
	return deaultGetter{}
}

func NewArchDetector(arch string,
	forceOneTask bool,
	getter efintf.ArchTypeGet) ArchDetector {
	return ArchDetector{
		arch,
		forceOneTask,
		getOrDefault(getter),
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
	return ad.getter.GetArchType() == dtuarch.EnflameT20 && ad.forceOneTask
}
