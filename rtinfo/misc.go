package rtinfo

type Interval [][]uint64

func (intv Interval) Len() int {
	return len(intv)
}

func (intv Interval) Swap(i, j int) {
	intv[i], intv[j] = intv[j], intv[i]
}
func (intv Interval) Less(i, j int) bool {
	return intv[i][0] < intv[j][0]
}
