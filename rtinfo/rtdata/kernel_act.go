package rtdata

import (
	"fmt"
	"strings"
)

type RuntimeInfo struct {
	TaskId   int
	SubIdx   int
	Name     string
	OpId     int
	SubValid bool
}

type KernelActivity struct {
	DpfAct
	RtInfo RuntimeInfo
}

func (kAct KernelActivity) GetSipOpName() (string, bool) {

	if kAct.RtInfo.SubValid {
		return stripDoubleColm(kAct.RtInfo.Name), true
	}
	return fmt.Sprintf("%v.[%v.%v]", kAct.RtInfo.Name,
		kAct.RtInfo.OpId,
		kAct.RtInfo.SubIdx), true
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
