package rtinfo

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/efintf/efconst"
	"git.enflame.cn/hai.bai/dmaster/efintf/sessintf"
	"git.enflame.cn/hai.bai/dmaster/meta"
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/misc/linklist"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

const MAX_BACKTRACE_TASK_COUNT = 6

var (
	ErrNoExecMeta = errors.New("no exec meta info for runtime")
)

var (
	ErrNoWildMatch = errors.New("no packet id found overall")
)

var (
	errNoStartTsEvent = errors.New("no start ts event")
)

type RuntimeTaskManagerBase struct {
	taskIdToTask   map[int]*rtdata.RuntimeTask // Full runtime task info, include cyccled ones and ones without cycles
	taskIdVec      []int                       // Full task id vec
	tsHead         *linklist.Lnk
	isOneSolidTask bool
}
type RuntimeTaskManager struct {
	RuntimeTaskManagerBase
	execKnowledge     *meta.ExecRaw
	orderedTaskVector []rtdata.OrderTask
	fullTaskVector    []rtdata.OrderTask
}

func NewRuntimeTaskManager(oneTask bool) *RuntimeTaskManager {
	return &RuntimeTaskManager{
		RuntimeTaskManagerBase: RuntimeTaskManagerBase{
			tsHead:         linklist.NewLnkHead(),
			isOneSolidTask: oneTask,
		},
	}
}

func (rtm *RuntimeTaskManagerBase) SelfClone() sessintf.ConcurEventSinker {
	assert.Assert(rtm.tsHead.ElementCount() == 0, "Must be empty")
	clonedTaskDc := make(map[int]*rtdata.RuntimeTask)
	for taskId, pToTask := range rtm.taskIdToTask {
		task := *pToTask // Copy
		clonedTaskDc[taskId] = &task
	}
	clonedTaskIdVec := make([]int, len(rtm.taskIdVec))
	copy(clonedTaskIdVec, rtm.taskIdVec)
	cloned := &RuntimeTaskManagerBase{
		taskIdToTask: clonedTaskDc,
		taskIdVec:    clonedTaskIdVec,
		tsHead:       linklist.NewLnkHead(),
	}
	return cloned
}

func (cloned *RuntimeTaskManagerBase) MergeTo(lhs interface{}) bool {
	master := lhs.(*RuntimeTaskManager)
	for taskId, pToTask := range cloned.taskIdToTask {
		if pToTask.CycleValid {
			pThisTask, ok := master.taskIdToTask[taskId]
			assert.Assert(ok, "must be there")
			pThisTask.StartCycle = pToTask.StartCycle
			pThisTask.EndCycle = pToTask.EndCycle
			pThisTask.CycleValid = true
		}
	}
	return true
}

func (rtm RuntimeTaskManagerBase) GetEngineTypeCodes() []codec.EngineTypeCode {
	return []codec.EngineTypeCode{codec.EngCat_TS}
}

// If there is an error, please propagate this event
func (rtm *RuntimeTaskManagerBase) DispatchEvent(evt codec.DpfEvent) error {
	if evt.Event == codec.TsLaunchCqmStart {
		rtm.tsHead.AppendAtTail(evt)
		return nil
	}
	if evt.Event == codec.TsLaunchCqmEnd {
		if start := rtm.tsHead.Extract(func(one interface{}) bool {
			un := one.(codec.DpfEvent)
			return un.Payload == evt.Payload && un.ClusterID == evt.ClusterID
		}); start != nil {
			startUn := start.(codec.DpfEvent)
			taskID := startUn.Payload
			if task, ok := rtm.GetTaskForId(taskID); !ok {
				// panic(fmt.Errorf("no start for cqm launch exec: task id(%v)", taskID))
				// make all these task invalid
				// The consequence is that
				// The task holds no executable information
				//  for there is no record on host site
				fmt.Fprintf(os.Stderr,
					"no start for cqm launch exec: task id(%v)\n", taskID)
			} else {
				rtm.updateTaskCycle(task, startUn.Cycle, evt.Cycle)
			}
			return nil
		}
		// No start is found
		return errNoStartTsEvent
	}

	// Do something for the rest of TS events ??
	return nil
}

func (rtm *RuntimeTaskManagerBase) updateTaskCycle(task *rtdata.RuntimeTask,
	startCycle, endCycle uint64) {
	if !rtm.isOneSolidTask {
		task.StartCycle = startCycle
		task.EndCycle = endCycle
		task.CycleValid = true
	}
}

func (rtm *RuntimeTaskManagerBase) GetTaskForId(taskId int) (
	*rtdata.RuntimeTask, bool) {
	if rtm.isOneSolidTask {
		taskId = efconst.SolidTaskID
	}
	task, ok := rtm.taskIdToTask[taskId]
	return task, ok
}

// For RuntimeTaskManager
func (rtm *RuntimeTaskManager) LoadRuntimeTask(
	infoReceiver efintf.InfoReceiver,
) bool {
	dc, taskSequentials, ok := infoReceiver.LoadTask()
	if !ok {
		return false
	}
	rtm.taskIdToTask, rtm.taskIdVec = dc, taskSequentials
	return true
}

func (r RuntimeTaskManager) DumpInfo() {
	if r.tsHead.ElementCount() > 0 {
		fmt.Fprintf(os.Stderr, "TS unmatched count: %v\n",
			r.tsHead.ElementCount())
	}
	fmt.Printf("# runtimetask (print the ones with cycle info only(TS event related)):\n")
	for _, taskId := range r.taskIdVec {
		if r.taskIdToTask[taskId].CycleValid {
			fmt.Printf("%v\n", r.taskIdToTask[taskId].ToString())
		}
	}
	fmt.Println()
}

func (r RuntimeTaskManager) GetExecRaw() *meta.ExecRaw {
	return r.execKnowledge
}

// After meta is loaded
// Ordered-task vector, element is placed in cycle orders
func (r *RuntimeTaskManager) BuildOrderInfo() {
	r.orderedTaskVector = r.createTaskSeq(true, true)
	r.fullTaskVector = r.createTaskSeq(false, true)
	log.Printf("%v order task has been built", len(r.orderedTaskVector))
	log.Printf("%v full order task has been built", len(r.fullTaskVector))
}

func (r RuntimeTaskManager) createTaskSeq(
	needCycleValid bool,
	needMetaValid bool,
) []rtdata.OrderTask {
	var orders []rtdata.OrderTask
	var check uint64 = 0
	if needCycleValid {
		check = 1
	}
	for _, task := range r.taskIdToTask {
		if (!needCycleValid || task.CycleValid) &&
			(!needMetaValid || task.MetaValid) {
			orders = append(orders, rtdata.NewOrderTask(
				task.StartCycle*check,
				task,
			))
		}
	}
	sort.Sort(rtdata.OrderTasks(orders))
	return orders
}

// LoadMeta will load executable raw from task info's executable-uuids
func (r *RuntimeTaskManager) LoadMeta(
	loader efintf.InfoReceiver,
) {
	if r.execKnowledge != nil {
		fmt.Printf("# warning: meta loaded already!!!!\n!!\n")
		return
	}
	execKm := meta.NewExecRaw(loader)
	for _, taskId := range r.taskIdVec {
		if r.taskIdToTask[taskId].CycleValid || true {
			if execKm.LoadMeta(r.taskIdToTask[taskId].ExecutableUUID) {
				r.taskIdToTask[taskId].MetaValid = true
			}
		}
	}
	execKm.LoadWildcard()
	execKm.DumpInfo()
	r.execKnowledge = execKm
}

func (r *RuntimeTaskManager) FindExecFor(execUuid uint64) metadata.ExecScope {
	exec, ok := r.execKnowledge.FindExecScope(execUuid)
	assert.Assert(ok, "Must be there")
	return exec
}

func (r *RuntimeTaskManager) LookupOpIdByPacketID(
	execUuid uint64,
	packetId int,
) (metadata.DtuOp, error) {
	if r.execKnowledge == nil {
		return metadata.DtuOp{}, ErrNoExecMeta
	}
	// Interception
	if efconst.IsWildcardExecuuid(execUuid) {
		return r.LookupOpOverall(packetId)
	}
	exec, ok := r.execKnowledge.FindExecScope(execUuid)
	if !ok {
		log.Printf("exec %016x is not loaded", exec)
		return metadata.DtuOp{}, ErrNoExecMeta
	}
	return exec.FindOp(packetId)
}

func (r *RuntimeTaskManager) LookupOpOverall(packetId int) (metadata.DtuOp, error) {
	var targetOp metadata.DtuOp
	var found = false
	r.execKnowledge.WalkExecScopes(func(es *metadata.ExecScope) bool {
		op, err := es.FindOp(packetId)
		if err == nil {
			found = true
			targetOp = op
			return false
		}
		return true
	})
	var err1 error
	if !found {
		err1 = ErrNoWildMatch
	}
	return targetOp, err1
}

func (r *RuntimeTaskManager) LookupDma(execUuid uint64, packetId int) (metadata.DmaOp, error) {
	if r.execKnowledge == nil {
		return metadata.DmaOp{}, ErrNoExecMeta
	}
	// Interception
	if efconst.IsWildcardExecuuid(execUuid) {
		return r.LookupDmaOverall(packetId)
	}
	exec, ok := r.execKnowledge.FindExecScope(execUuid)
	if !ok {
		log.Printf("exec %016x is not loaded", exec)
		return metadata.DmaOp{}, ErrNoExecMeta
	}
	return exec.FindDma(packetId)
}

func (r RuntimeTaskManager) LookupDmaOverall(packetId int) (metadata.DmaOp, error) {
	var targetDmaMeta metadata.DmaOp
	var found = false
	r.execKnowledge.WalkExecScopes(func(es *metadata.ExecScope) bool {
		if dmaMeta, err := es.FindDma(packetId); err == nil {
			found = true
			targetDmaMeta = dmaMeta
			return false
		}
		return true
	})
	var err1 error
	if !found {
		err1 = ErrNoWildMatch
	}
	return targetDmaMeta, err1
}

func (r RuntimeTaskManager) lowerBoundForTaskVec(cycle uint64) int {
	lz := len(r.orderedTaskVector)
	lo, hi := 0, lz
	vec := r.orderedTaskVector
	for lo < hi {
		md := (lo + hi) >> 1
		if cycle >= vec[md].StartCy {
			hi = md
		} else {
			lo = md + 1
		}
	}
	return lo
}

func (r RuntimeTaskManager) upperBoundForTaskVec(cycle uint64) int {
	lz := len(r.orderedTaskVector)
	lo, hi := 0, lz
	vec := r.orderedTaskVector
	for lo < hi {
		md := (lo + hi) >> 1
		if cycle < vec[md].StartCy {
			hi = md
		} else {
			lo = md + 1
		}
	}
	return lo
}

// GenerateDtuOps:  find dtu-op meta information for the Cqm Act
// Some ops (Cqm acts are combined into one)
func (rtm *RuntimeTaskManager) GenerateDtuOps(
	opActVec []rtdata.OpActivity,
	rule vgrule.EngineOrder,
) ([]rtdata.OpActivity, []rtdata.OpActivity) {
	// Each time we start processing a new session
	// We create a new object to do the math
	vec := rtm.orderedTaskVector
	for i := 0; i < len(vec); i++ {
		vec[i].CreateNewState()
	}

	bingoCount := 0
	unprocessedVec := []rtdata.OpActivity{}

	opState := NewOpXState()
	for i := 0; i < len(opActVec); i++ {
		curAct := &opActVec[i]
		start := curAct.StartCycle()
		idxStart := rtm.upperBoundForTaskVec(start)
		// backtrace for no more than MAX_BACKTRACE_TASK_COUNT
		found := false
		for j := idxStart - 1; j > idxStart-1-MAX_BACKTRACE_TASK_COUNT; j-- {
			if j < 0 || j >= len(vec) {
				continue
			}
			taskInOrder := vec[j]
			if !taskInOrder.IsValid() {
				continue
			}
			thisExecUuid := taskInOrder.GetExecUuid()
			if taskInOrder.AbleToMatchCqm(*curAct, rule) {
				if opInfo, err := rtm.LookupOpIdByPacketID(
					thisExecUuid,
					curAct.Start.PacketID); err == nil {
					// There is always a dtuop related to dbg op
					// and there is always a task
					taskInOrder.SuccessMatchDtuop(curAct.Start.PacketID)
					taskInOrder.SuccessMatchDtuop(curAct.End.PacketID)

					// Copy this result into op x-state
					cloneAct := *curAct
					cloneAct.SetOpRef(rtdata.NewOpRef(&opInfo,
						taskInOrder.GetRefToTask()))
					opState.AddOp(cloneAct)
					// curAct.SetOpRef(rtdata.NewOpRef(&opInfo, taskInOrder.GetRefToTask()))
					found = true
					break
				} else {
					// fmt.Printf("error: %v\n", err)
				}
			}
		}
		if found {
			bingoCount++
		} else {
			unprocessedVec = append(unprocessedVec, rtdata.OpActivity{
				DpfAct: curAct.DpfAct,
			})
		}
	}
	fmt.Printf("Dbg op/Dtu-op meta success matched count: %v out of %v\n",
		bingoCount,
		len(opActVec),
	)

	// OrderTasks(rtm.orderedTaskVector).DumpInfos(rtm)
	return opState.FinalizeOps(), unprocessedVec
}

func (rtm *RuntimeTaskManager) GenerateKernelActs(
	kernelActs []rtdata.KernelActivity,
	opBundles []rtdata.OpActivity,
	rule vgrule.EngineOrder) []rtdata.KernelActivity {
	vec := rtm.orderedTaskVector
	for i := 0; i < len(vec); i++ {
		vec[i].CreateNewState()
	}

	// TODO: refactor this
	// Retrieve Task Id Only
	var sipTaskBingoCount = 0
	for i, kernAct := range kernelActs {
		start := kernAct.StartCycle()
		idxStart := rtm.upperBoundForTaskVec(start)
		// backtrace for no more than MAX_BACKTRACE_TASK_COUNT
		found := false
		for j := idxStart - 1; j > idxStart-1-MAX_BACKTRACE_TASK_COUNT; j-- {
			if j < 0 || j >= len(vec) {
				continue
			}
			taskInOrder := vec[j]
			if !taskInOrder.IsValid() {
				continue
			}
			if taskInOrder.AbleToMatchSip(kernAct, rule) {
				kernelActs[i].RtInfo.TaskId = taskInOrder.GetTaskID()
				found = true
				break
			}
		}
		if found {
			sipTaskBingoCount++
		}
	}
	fmt.Printf("SIP task bingo %v out of %v\n", sipTaskBingoCount,
		len(kernelActs),
	)
	return GenerateKerenlActSeq(kernelActs, opBundles)
}

// Start from the first recorded task
func (r *RuntimeTaskManager) CookCqmEverSince(
	opActVec []rtdata.OpActivity,
	rule vgrule.EngineOrder,
) []rtdata.OpActivity {
	// Each time we start processing a new session
	// We create a new object to do the math

	validTaskMap := make(map[int]bool)
	for _, orderedTask := range r.orderedTaskVector {
		validTaskMap[orderedTask.GetTaskID()] = true
	}
	vec := r.fullTaskVector
	for i := 0; i < len(vec); i++ {
		vec[i].CreateNewState()
	}

	// If there is an ordered task vector, start from the very beginning
	// Let's assume that we do not miss any TS event from the start
	log.Printf("full vector size: %v", len(vec))
	startIdx := 0
	var firstTaskID = -1
	if len(r.orderedTaskVector) > 0 {
		startIdx = len(vec)
		firstTaskID = r.orderedTaskVector[0].GetTaskID()
		for i := 0; i < len(vec); i++ {
			if firstTaskID == vec[i].GetTaskID() {
				startIdx = i
				break
			}
		}
	}

	log.Printf("Ever since: [%v] taskid %v", startIdx, firstTaskID)
	bingoCount := 0
	unprocessedVec := []rtdata.OpActivity{}
	everSearched := false
	for i := 0; i < len(opActVec); i++ {
		curAct := &opActVec[i]
		found := false
		onceValid := false
	A100:
		for j := startIdx; j < len(vec); j++ {
			taskInOrder := vec[j]
			if validTaskMap[taskInOrder.GetTaskID()] {
				continue
			}
			everSearched = true
			thisExecUuid := taskInOrder.GetExecUuid()
			if taskInOrder.AbleToMatchCqm(*curAct, rule) {
				opInfo, err := r.LookupOpIdByPacketID(
					thisExecUuid,
					curAct.Start.PacketID)

				switch err {
				case nil:
					// There is always a dtuop related to dbg op
					// and there is always a task
					taskInOrder.SuccessMatchDtuop(curAct.Start.PacketID)
					taskInOrder.SuccessMatchDtuop(curAct.End.PacketID)
					curAct.SetOpRef(rtdata.NewOpRef(
						&opInfo,
						taskInOrder.GetRefToTask(),
					))
					found = true
					break A100
				case metadata.ErrValidPacketIdNoOp:
					onceValid = true
				}
			}
		}
		if found {
			bingoCount++
		} else {
			assert.Assert(!everSearched || onceValid, "must be valid for once")
			unprocessedVec = append(unprocessedVec, rtdata.OpActivity{
				DpfAct: curAct.DpfAct,
			})
		}
	}
	fmt.Printf("CookCqmEverSince: Success matched count: %v out of %v\n",
		bingoCount,
		len(opActVec),
	)
	fmt.Printf("  :and the rest unmatched is packet-valid-but-no-op\n")
	return unprocessedVec
}

// OvercookCqm:  find dtu-op meta information for the Cqm Act
func (r *RuntimeTaskManager) OvercookCqm(
	opActVec []rtdata.OpActivity,
	rule vgrule.EngineOrder,
) {
	// Each time we start processing a new session
	// We create a new object to do the math
	vec := r.orderedTaskVector
	for i := 0; i < len(vec); i++ {
		vec[i].CreateNewState()
	}

	bingoCount := 0
	noBingoCount := 0

	// Map from the first unmatched packet id to CQM engine index
	noMatchedPacketId := make(map[int]int)
	cqmUnmatched := make(map[int]map[int]bool)

	addUnmatchedToCqm := func(evt codec.DpfEvent) {
		assert.Assert(evt.EngineTypeCode == codec.EngCat_CQM, "Must be CQM")
		if _, ok := cqmUnmatched[evt.EngineIndex]; !ok {
			cqmUnmatched[evt.EngineIndex] = make(map[int]bool)
		}
		cqmUnmatched[evt.EngineIndex][evt.PacketID] = true
	}

	for i := 0; i < len(opActVec); i++ {
		curAct := &opActVec[i]
		found := false
		for j := len(vec) - 1; j >= 0; j-- {
			if j < 0 || j >= len(vec) {
				continue
			}
			taskInOrder := vec[j]
			if !taskInOrder.IsValid() {
				continue
			}
			if taskInOrder.AbleToMatchCqm(*curAct, rule) {
				thisExecUuid := taskInOrder.GetExecUuid()
				opInfo, err := r.LookupOpIdByPacketID(
					thisExecUuid,
					curAct.Start.PacketID)
				switch err {
				case nil:
					// There is always a dtuop related to dbg op
					// and there is always a task
					taskInOrder.SuccessMatchDtuop(curAct.Start.PacketID)
					taskInOrder.SuccessMatchDtuop(curAct.End.PacketID)
					curAct.SetOpRef(rtdata.NewOpRef(
						&opInfo,
						taskInOrder.GetRefToTask(),
					))
					found = true
				case metadata.ErrInvalidPacketId:
				case metadata.ErrValidPacketIdNoOp:
				default:
					assert.Assert(false, "not included")
				}
			}
		}
		if found {
			bingoCount++
		} else {
			noBingoCount++
			noMatchedPacketId[curAct.Start.PacketID] |=
				1 << curAct.Start.EngineIndex
			addUnmatchedToCqm(curAct.Start)
		}
	}
	fmt.Printf("Dbg op/Dtu-op meta wildcard success matched count: %v out of %v\n",
		bingoCount,
		noBingoCount+bingoCount,
	)

	fmt.Printf("no matched count: %v\n", len(noMatchedPacketId))
	limitedCc := 10
	cqmEngs := 0
	for pkt, cqmEng := range noMatchedPacketId {
		limitedCc--
		if limitedCc > 0 {
			fmt.Printf("no matched packet id: %v for Cqm(%v)\n",
				pkt, toCqmGroup(cqmEng))
		}
		cqmEngs |= cqmEng
	}
	fmt.Printf("in summary: %v\n", toCqmGroup(cqmEngs))
	if len(cqmUnmatched[0]) > 0 {
		limitedCc := 10
		fmt.Printf("For CQM ZERO:\n")
		for pkt := range cqmUnmatched[0] {
			limitedCc--
			if limitedCc > 0 {
				fmt.Printf("  unmatched packet: %v\n", pkt)
			} else if limitedCc == 0 {
				fmt.Printf("  ...\n")
			}
		}
	}
}

func toCqmGroup(engBitmap int) string {
	var rv string
	for i := 0; i < 31; i++ {
		if engBitmap&(1<<i) != 0 {
			rv += fmt.Sprintf("%v,", i)
		}
	}
	return rv
}

// WildCookCqm:  By all means
func (r *RuntimeTaskManager) WildCookCqm(
	opActVec []rtdata.OpActivity,
) []rtdata.OpActivity {
	var unmatchedVec []rtdata.OpActivity
	for i := 0; i < len(opActVec); i++ {
		curAct := &opActVec[i]
		if _, ok := r.execKnowledge.LookForWild(
			curAct.Start.PacketID,
			false); ok {
			// not change at all
		} else {
			unmatchedVec = append(unmatchedVec, rtdata.OpActivity{
				DpfAct: curAct.DpfAct,
			})
		}
	}

	if len(unmatchedVec) == 0 {
		log.Printf("all %v matched in wild match mode", len(unmatchedVec))
	} else {
		log.Printf("done wild cook, %v out of %v got unmatched",
			len(unmatchedVec),
			len(opActVec),
		)
		assert.Assert(false, "this is not expected")
	}
	return unmatchedVec
}

func (rtm *RuntimeTaskManager) DumpInfos(orderTask rtdata.OrderTasks) {
	fmt.Printf("# statistics for ordered-task\n")
	for _, task := range orderTask {
		execScope := rtm.FindExecFor(task.GetExecUuid())
		task.DumpStatusInfo(execScope)
	}
}

func (rtm *RuntimeTaskManager) CookDma(
	dmaActVec []rtdata.DmaActivity,
	algo vgrule.ActMatchAlgo,
) []rtdata.DmaActivity {
	vec := rtm.orderedTaskVector
	for i := 0; i < len(vec); i++ {
		vec[i].CreateNewState()
	}

	bingoCount := 0
	errDmaCount := 0
	cdmaErrCount, sdmaErrCount := 0, 0
	skippedDmaCount := 0
	const errPrintLimit = 10
	for i := 0; i < len(dmaActVec); i++ {
		curAct := &dmaActVec[i]
		startCy := curAct.StartCycle()
		idxStart := rtm.upperBoundForTaskVec(startCy)
		foundDmaMeta := false
	A100:
		for j := idxStart - 1; j >= 0 &&
			j > idxStart-1-MAX_BACKTRACE_TASK_COUNT; j-- {
			// taskInOrder := vec[j]
			taskInOrder := vec[j]
			if !taskInOrder.IsValid() {
				// SKIP. It is danger to skip
				// just because it is not setup correctly with meta.
				continue
			}
			if !taskInOrder.MatchXDMA(*curAct, algo) {
				continue
			}
			if dmaOp, err := rtm.LookupDma(taskInOrder.GetExecUuid(),
				curAct.Start.PacketID); err == nil {
				curAct.SetDmaRef(rtdata.NewDmaRef(&dmaOp,
					taskInOrder.GetRefToTask()))
				foundDmaMeta = true
				break A100
			}
		} // for backtrace all possible tasks
		if foundDmaMeta {
			bingoCount++
		} else {
			if !shallSkipErrDma(curAct.Start) {
				errDmaCount++
				if errDmaCount < errPrintLimit {
					fmt.Printf("error for no meta dma packet id = %v, %v\n",
						curAct.Start.PacketID,
						curAct.Start.EngineTy,
					)
				} else if errDmaCount == errPrintLimit {
					fmt.Printf("too many dma errors\n")
				}

				// statistics
				switch curAct.Start.EngineTypeCode {
				case codec.EngCat_SDMA:
					sdmaErrCount++
				case codec.EngCat_CDMA:
					cdmaErrCount++
				}
			} else {
				skippedDmaCount++
			}
		}
	}
	fmt.Printf("Dma meta set SUCCESS %v out of %v, error count: %v\n",
		bingoCount,
		len(dmaActVec),
		errDmaCount,
	)
	fmt.Printf("   %v are skipped\n", skippedDmaCount)
	fmt.Printf("   Cdma error: %v, Sdma error: %v\n",
		cdmaErrCount, sdmaErrCount)
	return nil
}

// For now only some SDMA are skipped
func shallSkipErrDma(evt codec.DpfEvent) bool {
	return evt.EngineTypeCode == codec.EngCat_SDMA && evt.PacketID == 0x5beaf
}
