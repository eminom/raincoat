package infoloader

import "strings"

func XSplit(a string, count int) []string {
	var intv [][]int
	lz := len(a)
	start := 0
	extractOne := func() int {
		i := start
		for i < lz && a[i] != ' ' {
			i++
		}
		if start < i {
			intv = append(intv, []int{start, i})
		}
		return i
	}
	for start < lz && len(intv) < count-1 {
		for start < lz && a[start] == ' ' {
			start++
		}
		start = extractOne()
	}
	// Still, skipping all spaces
	for start < lz && a[start] == ' ' {
		start++
	}
	if start < lz {
		intv = append(intv, []int{start, lz})
	}
	vs := make([]string, len(intv))
	for i, r := range intv {
		vs[i] = a[r[0]:r[1]]
		if len(intv)-1 == i {
			vs[i] = strings.TrimRight(vs[i], " ")
		}
	}
	return vs
}
