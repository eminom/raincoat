package rtdata

import (
	"fmt"
)

type RuntimeTask struct {
	TaskID         int
	ExecutableUUID uint64
	PgMask         int

	StartCycle uint64
	EndCycle   uint64
	CycleValid bool
	MetaValid  bool
}

func (r RuntimeTask) ToString() string {
	return fmt.Sprintf("Task(%v) %016x %v,[%v,%v]",
		r.TaskID,
		r.ExecutableUUID,
		r.PgMask,
		r.StartCycle,
		r.EndCycle,
	)
}

func (r RuntimeTask) ToShortString() string {
	hex := fmt.Sprintf("%016x", r.ExecutableUUID)[:8]
	return fmt.Sprintf("PG %v Task %v Exec %v",
		r.PgMask,
		r.TaskID,
		hex,
	)
}

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

func (r RuntimeTask) MatchSip(sipOrder int) bool {
	return r.PgMask&(1<<sipOrder) != 0
}
