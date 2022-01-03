package infoloader

type FormatFetcher interface {
	FetchOpIdOpName(string) (string, string)
}

type compatibleFetcher struct{}

func (c *compatibleFetcher) FetchOpIdOpName(text string) (string, string) {
	vs := XSplit(text, 2)
	return vs[0], vs[1]
}

type nuevoModeFetcher struct{}

func (c *nuevoModeFetcher) FetchOpIdOpName(text string) (string, string) {
	vs := XSplit(text, 3)
	return vs[0], vs[2]
}

func newCompatibleOpFetcher() FormatFetcher {
	return new(compatibleFetcher)
}

func newNuevoOpFetcher() FormatFetcher {
	return new(nuevoModeFetcher)
}

func XSplit(a string, count int) []string {
	var intv [][]int
	lz := len(a)
	start, i := 0, 0
	for i < lz && len(intv) < count {
		for i < lz && a[i] != ' ' {
			i++
		}
		intv = append(intv, []int{start, i})
		for i < lz && a[i] == ' ' {
			i++
		}
		start = i
	}
	if start < lz {
		intv = append(intv, []int{start, i})
	}
	vs := make([]string, len(intv))
	for i, r := range intv {
		vs[i] = a[r[0]:r[1]]
	}
	return vs
}
