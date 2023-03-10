package rtdata

import (
	"fmt"
	"log"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/efintf/efconst"
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
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
}

type OrderTaskStated struct {
	OrderTask
	taskState OrderTaskState
}

func NewOrderTask(startCycle uint64, task *RuntimeTask) OrderTask {
	return OrderTask{
		StartCy:   startCycle,
		refToTask: task,
	}
}

func NewOrderTaskStated(origin OrderTask) OrderTaskStated {
	var rv OrderTaskStated
	rv.OrderTask = origin
	rv.taskState = NewOrderTaskState()
	return rv
}

func (ot OrderTask) GetRefToTask() *RuntimeTask {
	return ot.refToTask
}

func (ot OrderTask) GetTaskID() int {
	return ot.refToTask.TaskID
}

func (ot OrderTask) GetExecUuid() uint64 {
	return ot.refToTask.ExecutableUUID
}

func (ot OrderTask) AbleToMatchCqm(evt codec.DpfEvent, a vgrule.EngineOrder) bool {
	if efconst.IsAllZeroPgMask(ot.refToTask.PgMask) {
		return true
	}
	return ot.refToTask.MatchCqm(a.GetCqmEngineOrder(evt))
}

func (ot OrderTask) AbleToMatchSip(evt codec.DpfEvent, a vgrule.EngineOrder) bool {
	if efconst.IsAllZeroPgMask(ot.refToTask.PgMask) {
		return true
	}
	return ot.refToTask.MatchSip(a.GetSipEngineOrder(evt))
}

func (ot OrderTask) AbleToMatchXDMA(dmaEvt codec.DpfEvent, a vgrule.EngineOrder) bool {
	pgMask := ot.refToTask.PgMask
	if efconst.IsAllZeroPgMask(pgMask) {
		return true
	}
	if dmaEvt.IsOfEngine(codec.EngCat_CDMA) {
		return a.MapPgMaskBitsToCdmaEngineMask(pgMask)&
			(1<<a.GetCdmaPgBitOrder(dmaEvt)) != 0
	} else if dmaEvt.IsOfEngine(codec.EngCat_SDMA) {
		return a.MapPgMaskBitsToSdmaEngineMask(pgMask)&
			(1<<a.GetSdmaPgBitOrder(dmaEvt)) != 0
	}
	return false
}

func (ot *OrderTaskStated) SuccessMatchDtuop(packetId int) {
	ot.taskState.matchedPacketIdMap[packetId]++
}

// A task is an instance of executable
// So all matched packets must belong to the same executable
func (ot OrderTaskStated) DumpStatusInfo(exec metadata.ExecScope) {
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
	matchedUniqueOpIdSet := make(map[int]bool)
	for pktId := range ot.taskState.matchedPacketIdMap {
		opId, ok := exec.MapPacketIdToOpId(pktId)
		assert.Assert(ok, "must be valid packet id for %v", pktId)
		matchedUniqueOpIdSet[opId] = true
		matchedPacketIdCount++
	}
	exec.CheckOpMapStatus(matchedUniqueOpIdSet)
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
	if o[i].StartCy != o[j].StartCy {
		return o[i].StartCy < o[j].StartCy
	}
	return o[i].refToTask.TaskID < o[j].refToTask.TaskID
}
