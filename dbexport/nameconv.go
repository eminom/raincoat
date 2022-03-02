package dbexport

import "fmt"

type NameConverter struct {
	dc map[int]map[string]int
}

func NewNameConverter() NameConverter {
	return NameConverter{
		dc: make(map[int]map[string]int),
	}
}

func (nc NameConverter) GetIndexedName(mid, ctx int, name string) string {
	idx := combineId(mid, ctx)
	if _, ok := nc.dc[idx]; !ok {
		nc.dc[idx] = make(map[string]int)
	}

	seq := nc.dc[idx][name]
	nc.dc[idx][name] = seq + 1
	return fmt.Sprintf("%v.%v", name, seq)
}

func combineId(masterId int, ctx int) int {
	return masterId<<4 + ctx
}
