package sess

import (
	"fmt"
	"io"

	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type PgStatInfo struct {
	distribute  []int
	count       int
	pgMask      int
	engineOrder vgrule.EngineOrder
}

func NewPgStatInfo(pgMask int, engineOrder vgrule.EngineOrder) PgStatInfo {
	return PgStatInfo{
		distribute:  make([]int, engineOrder.GetMaxMasterId()),
		count:       0,
		pgMask:      pgMask,
		engineOrder: engineOrder,
	}
}

func (pgInfo *PgStatInfo) Tick(mid int) {
	pgInfo.count++
	pgInfo.distribute[mid]++
}

func (pgInfo PgStatInfo) IsEmpty() bool {
	return pgInfo.count <= 0
}

func (pgInfo PgStatInfo) DumpInfo(out io.Writer) {
	fmt.Fprintf(out, "=== PgMask %06b ===\n", pgInfo.pgMask)
	// master id such as 9(nine)
	// is not assigned to a particular engine
	// so the statistics work does not ever see mid=9 on Ticking
	for mid, count := range pgInfo.distribute {
		if count > 0 {
			fmt.Fprintf(out, "%v: %v\n", pgInfo.engineOrder.CheckoutEngineString(mid), count)
		}
	}
	fmt.Fprintf(out, "\n")
}
