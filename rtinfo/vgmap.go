package rtinfo

func matchIntLowest(v int, idx int) bool {
	for i := 0; i < 31; i++ {
		if (1<<i)&v != 0 {
			return i == idx
		}
	}
	return false
}

func (r RuntimeTask) MatchCqm(cqmIdx int) bool {
	return matchIntLowest(r.PgMask, cqmIdx)
}
