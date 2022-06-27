package affinity

type DoradoCdmaAffinityDefault struct{}

func (DoradoCdmaAffinityDefault) GetCdmaIdxToPg(cid int, eid int) int {
	// (0, 0), (0, 1), (0, 2), pg 012 if single
	// (1, 0), (1, 1), (1, 2), pg 345 if single
	return cid*3 + eid
}
