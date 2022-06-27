package affinity

type DoradoCdmaAffinityDefault struct{}

func (DoradoCdmaAffinityDefault) GetCdmaIdxToPg(cid int, eid int) int {
	// (0, 0), (0, 1), (0, 2), pg 012 if single
	// (1, 0), (1, 1), (1, 2), pg 345 if single
	if eid == 3 {
		// (0, 3), the fourth CDMA engine is mapped to (3pg pg000111)
		// (1, 3), the fourth CDMA engine is mapped to (3pg pg111000)
		return cid * 3
	}
	return cid*3 + eid
}
