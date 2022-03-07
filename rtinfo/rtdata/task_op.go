package rtdata

type TaskActivity struct {
	DpfAct
}

type TaskActivityVec []TaskActivity

func (fwa TaskActivityVec) Len() int {
	return len(fwa)
}

func (fwa TaskActivityVec) Less(i, j int) bool {
	return fwa[i].StartCycle() < fwa[j].StartCycle()
}

func (fwa TaskActivityVec) Swap(i, j int) {
	fwa[i], fwa[j] = fwa[j], fwa[i]
}
