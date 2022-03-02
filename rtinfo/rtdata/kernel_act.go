package rtdata

type KernelActivity struct {
	DpfAct
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
