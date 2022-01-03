package rtdata

import (
	"fmt"
	"log"
	"sort"
)

func CheckTimespanOverlapping(bundle []OpActivity) {
	log.Printf("start checking overlap over %v item(s)", len(bundle))
	var intvs Interval
	// Check intervals
	for _, act := range bundle {
		if act.opRef.dtuOp != nil {
			intvs = append(intvs, []uint64{
				act.StartCycle(),
				act.EndCycle(),
			})
		}
	}
	sort.Sort(intvs)
	overlappedCount := 0
	for i := 0; i < len(intvs)-1; i++ {
		if intvs[i][1] >= intvs[i+1][0] {
			log.Printf("error: %v >= %v at [%d] out of [%v]",
				intvs[i][1], intvs[i+1][0],
				i, len(intvs))
			overlappedCount++
			break
		}
	}
	// dumpIntvs(intvs)
	if overlappedCount > 0 {
		fmt.Printf("warning: there is %v overlapping\n",
			overlappedCount)
		// assert.Assert(overlappedCount == 0,
		// 	"overlapped count must be zero: but %v",
		// 	overlappedCount)
	} else {
		log.Printf("no overlapped confirmed")
	}
}
