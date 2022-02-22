package main

import (
	"fmt"
	"log"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

func dumpFullCycles(bundle []rtdata.OpActivity) {
	var intvs rtdata.Interval
	// Check intervals

	// master-id to last cycle value
	prevCycleEnds := make(map[int]uint64)
	prevLastPid := make(map[int]int)

	var errs []error

	for _, act := range bundle {
		intvs = append(intvs, []uint64{
			act.StartCycle(),
			act.EndCycle(),
		})

		combinedCtx := act.Start.MasterIdValue()<<codec.RTCONTEXT_BITCOUNT +
			act.Start.Context
		opId := -1
		if act.IsOpRefValid() {
			opId = act.GetOp().OpId
		}
		if lastEndCycle, ok := prevCycleEnds[combinedCtx]; ok {
			if act.StartCycle() <= lastEndCycle {
				errs = append(errs,
					fmt.Errorf("error crossed for master-id:%v, context: %v, pktid: %v, op-id: %v\nlast cycle: %v, last packet-id: %v",
						act.Start.MasterIdValue(),
						act.Start.Context,
						act.Start.PacketID,
						opId,
						lastEndCycle,
						prevLastPid[combinedCtx],
					))
			}
		}
		// update the last one
		prevCycleEnds[combinedCtx] = act.EndCycle()
		prevLastPid[combinedCtx] = act.Start.PacketID
	}

	if len(errs) > 0 {
		fmt.Printf("#########################################\n")
		fmt.Printf("##### Start end crossed for op act  #####\n")
		fmt.Printf("#########################################\n")

		for _, err := range errs {
			fmt.Printf("%v\n", err)
		}
		fmt.Println()
	}

	// sort.Sort(intvs)
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
	if overlappedCount != 0 {
		log.Printf("warning: must be no overlap in full dump: but there is %v",
			overlappedCount,
		)
	}
}
