package topsdev

import (
	"fmt"
	"os"

	"git.enflame.cn/hai.bai/dmaster/efintf/affinity"
	"git.enflame.cn/hai.bai/dmaster/meta/dtuarch"
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/infoloader"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/topsdev/mimic/mimicdefs"
)

type execFileLoader struct {
	files []string
}

func NewExecLoader(names []string) *execFileLoader {
	files := make([]string, len(names))
	copy(files, names)
	return &execFileLoader{files: files}
}

func (e execFileLoader) LoadTask(oneSolid bool) (
	dc map[int]*rtdata.RuntimeTask,
	taskSeq []int, ok bool) {
	return infoloader.OneSolidTaskLoader{}.LoadTask(oneSolid)
}

func (e execFileLoader) GetArchType() dtuarch.ArchType {
	return dtuarch.EnflameI20
}

func (e execFileLoader) LoadExecScope(execUuid uint64) *metadata.ExecScope {
	return nil
}

func (e execFileLoader) GetCdmaAffinity() affinity.CdmaAffinitySet {
	return affinity.NewDoradoCdmaAffinityDefault()
}

func (e execFileLoader) LoadTimepoints() ([]rtdata.HostTimeEntry, bool) {
	return nil, false
}

func (e execFileLoader) LoadWildcards(
	checkExist func(str string) bool,
	notifyNew func(uint64, *metadata.ExecScope)) {
	// The outter container does not hold anything
	fmt.Fprintf(os.Stderr, "load meta from %v\n", e.files)
	var execScopes []*metadata.ExecScope
	for _, file := range e.files {
		scopes := LoadExecScopeFromExec(file)
		execScopes = append(execScopes, scopes...)
	}
	for _, exec := range execScopes {
		notifyNew(exec.GetExecUuid(), exec)
	}
}

func (e execFileLoader) ExtractHostInfo() *mimicdefs.HostInfo {
	return nil
}
