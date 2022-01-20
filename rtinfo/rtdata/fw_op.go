package rtdata

type FwActivity struct {
	DpfAct
}

type FwActivityVec []FwActivity

func (fwa FwActivityVec) Len() int {
	return len(fwa)
}

func (fwa FwActivityVec) Less(i, j int) bool {
	return fwa[i].StartCycle() < fwa[j].StartCycle()
}

func (fwa FwActivityVec) Swap(i, j int) {
	fwa[i], fwa[j] = fwa[j], fwa[i]
}
