package sess

import (
	"fmt"
	"io"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

type PgStatInfoSub struct {
	distribute []IEngineEvtStat
	count      int
	pgMask     int
}

func NewPgStatInfoSub(pgMask int, maxMid int, midToString func(int) (string, bool)) PgStatInfoSub {
	distribute := make([]IEngineEvtStat, maxMid)

	for i := 0; i < maxMid; i++ {
		masterName, _ := midToString(i)
		switch {
		case strings.HasPrefix(masterName, "CDMA") ||
			strings.HasPrefix(masterName, "SDMA"):
			distribute[i] = &DteStat{}
		default:
			distribute[i] = &EngineEvtStat{}
		}
	}

	return PgStatInfoSub{
		distribute: distribute,
		count:      0,
		pgMask:     pgMask,
	}
}

func (pgInfo *PgStatInfoSub) TickSub(mid int, format int, event int) {
	pgInfo.count++
	pgInfo.distribute[mid].TickEvent(format, event)
}

func (pgInfo PgStatInfoSub) IsEmpty() bool {
	return pgInfo.count <= 0
}

func (pgInfo PgStatInfoSub) DumpInfo(midToString func(int) string, out io.Writer) {
	fmt.Fprintf(out, "=== PgMask %06b ===\n", pgInfo.pgMask)
	// master id such as 9(nine)
	// is not assigned to a particular engine
	// so the statistics work does not ever see mid=9 on Ticking
	for mid, distStat := range pgInfo.distribute {
		if !distStat.Empty() {
			fmt.Fprintf(out, "%v: %v\n", midToString(mid), distStat.ToString())
		}
	}
	fmt.Fprintf(out, "\n")
}

type PgStatInfo struct {
	pgInfoArr   []PgStatInfoSub
	engineOrder vgrule.EngineOrder
}

func NewPgStatInfo(engineOrder vgrule.EngineOrder) PgStatInfo {
	pgMax := engineOrder.GetMaxPgOrderIndex()
	pgInfoArr := make([]PgStatInfoSub, pgMax)
	for i := 0; i < pgMax; i++ {
		pgInfoArr[i] = NewPgStatInfoSub(
			1<<i, engineOrder.GetMaxMasterId(),
			engineOrder.MayCheckoutEngineString)
	}
	return PgStatInfo{
		pgInfoArr:   pgInfoArr,
		engineOrder: engineOrder,
	}
}

func (pgS *PgStatInfo) Tick(dpf codec.DpfEvent) {
	pgIdx := pgS.engineOrder.GetEngineOrderIndex(dpf)
	if pgIdx >= 0 {
		pgS.pgInfoArr[pgIdx].TickSub(
			dpf.EngineUniqIdx, dpf.Flag, dpf.Event)
	}
}

func (pgS PgStatInfo) DumpInfo(out io.Writer) {
	checker := pgS.engineOrder.CheckoutEngineString
	for i := 0; i < len(pgS.pgInfoArr); i++ {
		if !pgS.pgInfoArr[i].IsEmpty() {
			pgS.pgInfoArr[i].DumpInfo(checker, out)
		}
	}
}
