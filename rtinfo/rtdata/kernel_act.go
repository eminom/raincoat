package rtdata

import "fmt"

type RuntimeInfo struct {
	TaskId int
	SubIdx int
	Name   string
	OpId   int
}

type KernelActivity struct {
	DpfAct
	RtInfo RuntimeInfo
}

func (kAct KernelActivity) GetName() (string, bool) {
	if len(kAct.RtInfo.Name) > 0 {
		return fmt.Sprintf("%v.%v.%v", kAct.RtInfo.Name,
			kAct.RtInfo.OpId,
			kAct.RtInfo.SubIdx), true
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
