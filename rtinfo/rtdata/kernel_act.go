package rtdata

import (
	"strings"
)

type RuntimeTaskInfo struct {
	TaskId int
}

type RuntimeInfo struct {
	RuntimeTaskInfo
	SubIdx   int
	Name     string
	OpId     int
	SubValid bool
}

func (rti *RuntimeInfo) Update(subIdx int, name string, opId int) {
	rti.SubIdx = subIdx
	rti.Name = name
	rti.OpId = opId
	rti.SubValid = true
}

type KernelActivity struct {
	DpfAct
	RtInfo RuntimeInfo
}

func (kAct KernelActivity) GetSipOpName() (string, bool) {

	if kAct.RtInfo.SubValid {
		return stripDoubleColm(kAct.RtInfo.Name), true
	}
	return "", false
}

func (kAct KernelActivity) GetEngineIndex() int {
	return kAct.Start.EngineIndex
}

type KernelActivityVec []KernelActivity

func (kActVec KernelActivityVec) Len() int {
	return len(kActVec)
}

func (kActVec KernelActivityVec) Less(i, j int) bool {
	return kActVec[i].StartCycle() < kActVec[j].StartCycle()
}

func (kActVec KernelActivityVec) Swap(i, j int) {
	kActVec[i], kActVec[j] = kActVec[j], kActVec[i]
}

func stripDoubleColm(a string) string {
	if pos := strings.LastIndex(a, "::"); pos >= 0 && len(a)-pos-2 > 0 {
		return a[pos+2:]
	}
	return a
}
