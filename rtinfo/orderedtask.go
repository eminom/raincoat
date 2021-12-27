package rtinfo

import (
	"fmt"
	"log"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/meta"
)

type OrderTaskState struct {
	matchedPacketIdMap map[int]int
}

func NewOrderTaskState() OrderTaskState {
	return OrderTaskState{
		matchedPacketIdMap: make(map[int]int),
	}
}

type OrderTask struct {
	StartCy   uint64
	refToTask *RuntimeTask
	taskState OrderTaskState
}

func NewOrderTask(startCycle uint64, task *RuntimeTask) OrderTask {
	return OrderTask{
		StartCy:   startCycle,
		refToTask: task,
	}
}

func (ot OrderTask) AbleToMatchCqm(cqm CqmActBundle) bool {
	return ot.refToTask.MatchCqm(cqm)
}

func (ot *OrderTask) SuccessMatchDtuop(packetId int) {
	ot.taskState.matchedPacketIdMap[packetId]++
}

func (ot OrderTask) DumpInfo(exec meta.ExecScope) {
	assert.Assert(ot.refToTask != nil, "must not be nil, created from start")
	if !ot.IsValid() {
		log.Printf("not a valid task")
	}
	fmt.Printf("Task ID: %v, Exec %0x16\n",
		ot.refToTask.TaskID,
		ot.refToTask.ExecutableUUID,
	)
	fmt.Printf("  Matched distinct dtuop packet count: %v\n",
		len(ot.taskState.matchedPacketIdMap),
	)
	matchedPacketIdCount := 0
	// a packet id with a task scope must belong to the same executable
	for pktId := range ot.taskState.matchedPacketIdMap {
		ok := exec.IsValidPacketId(pktId)
		assert.Assert(ok, "must be valid packet id for %v", pktId)
		matchedPacketIdCount++
	}
	missedPacketMap := make(map[int]bool)
	exec.IteratePacketId(func(pkt int) {
		if _, ok := ot.taskState.matchedPacketIdMap[pkt]; !ok {
			missedPacketMap[pkt] = true
		}
	})
	if len(missedPacketMap) > 0 {
		fmt.Printf("  error: %v packets missing for this task/exec\n",
			len(missedPacketMap),
		)
	}
}

func (ot OrderTask) IsValid() bool {
	return ot.refToTask.CycleValid && ot.refToTask.MetaValid
}

type OrderTasks []OrderTask

func (o OrderTasks) Len() int {
	return len(o)
}

func (o OrderTasks) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o OrderTasks) Less(i, j int) bool {
	return o[i].StartCy < o[j].StartCy
}
