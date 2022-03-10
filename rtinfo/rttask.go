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

// var (
// 	errNoStartTsEvent = errors.New("no start ts event")
// )

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

	tempInternal RuntimeTaskManInternal
}

func newStatedOrderTaskVector(origin []rtdata.OrderTask) []rtdata.OrderTaskStated {
	lz := len(origin)
	out := make([]rtdata.OrderTaskStated, lz)
	for i, v := range origin {
		out[i] = rtdata.NewOrderTaskStated(v)
	}
	return out
}

type RuntimeTaskManInternal struct {
	subOpInformation map[string]map[string][]string
}

func NewRuntimeTaskManager(oneTask bool) *RuntimeTaskManager {
	return &RuntimeTaskManager{
		RuntimeTaskManagerBase: RuntimeTaskManagerBase{
			tsHead:         linklist.NewLnkHead(),
			isOneSolidTask: oneTask,
		},
	}
}

func (rtm *RuntimeTaskManagerBase) ProcessTaskActVector(
	taskActVec rtdata.TaskActivityVec) {
	for _, taskAct := range taskActVec {
		taskID := taskAct.Start.Payload
		if task, ok := rtm.GetTaskForId(taskID); !ok {
			fmt.Fprintf(os.Stderr,
				"no host site cqm launch exec info: task id(%v)\n", taskID)
		} else {
			rtm.updateTaskCycle(task, taskAct.Start.Cycle, taskAct.End.Cycle)
		}
	}
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
	bingoCount := 0
	terminatorCount := 0
	unprocessedVec := []rtdata.OpActivity{}

	opState := NewOpXState()

	lookupOpMeta := func(execUuid uint64, packetId int) bool {
		_, err := rtm.LookupOpIdByPacketID(execUuid, packetId)
		return err == nil
	}

	for i := 0; i < len(opActVec); i++ {
		curAct := &opActVec[i]

		// terminator, fit and quit
		isTerminator := (codec.DbgPktDetector{}).IsTerminatorMark(curAct.Start)
		var exhaustiveMatcher MatchExtraConds = lookupOpMeta
		if isTerminator {
			terminatorCount++
			exhaustiveMatcher = nil
		}
		taskInOrder, found := rtm.locateTask(
			curAct.Start, rule, MatchToCqm{},
			exhaustiveMatcher,
		)
		if found {
			if isTerminator {
				opState.CombineOps(taskInOrder.GetTaskID(),
					curAct.ContextId())
				continue
			}
			thisExecUuid := taskInOrder.GetExecUuid()
			if opInfo, err := rtm.LookupOpIdByPacketID(
				thisExecUuid,
				curAct.Start.PacketID); err == nil {
				// There is always a dtuop related to dbg op
				// and there is always a task
				// TODO: Statistics can be done later
				// taskInOrder.SuccessMatchDtuop(curAct.Start.PacketID)
				// taskInOrder.SuccessMatchDtuop(curAct.End.PacketID)

				// Copy this result into op x-state
				cloneAct := *curAct
				cloneAct.SetOpRef(rtdata.NewOpRef(&opInfo,
					taskInOrder.GetRefToTask()))
				opState.AddOp(cloneAct)
				// curAct.SetOpRef(rtdata.NewOpRef(&opInfo, taskInOrder.GetRefToTask()))
			}
		}
		// Test again
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
		len(opActVec)-terminatorCount,
	)

	// OrderTasks(rtm.orderedTaskVector).DumpInfos(rtm)
	return opState.FinalizeOps(), unprocessedVec
}

/*
 * One task, one executable activity
 */
func (rtm *RuntimeTaskManager) GenerateTaskOps(
	fwActs []rtdata.FwActivity,
	rule vgrule.EngineOrder,
) map[int]rtdata.FwActivity {
	var taskIdToActivity = make(map[int]rtdata.FwActivity)
	for _, fwAct := range fwActs {
		if fwAct.Start.EngineTypeCode == codec.EngCat_CQM &&
			fwAct.Start.Event == codec.CqmExecutableStart {
			task, found := rtm.locateTask(
				fwAct.Start,
				rule,
				MatchToCqm{},
				nil,
			)
			if found {
				taskId := task.GetTaskID()
				if _, ok := taskIdToActivity[taskId]; !ok {
					taskIdToActivity[taskId] = fwAct
				} else {
					fmt.Fprintf(
						os.Stderr,
						"error duplicate executable to task for task id: %v\n",
						taskId)
				}
			}
		}
	}
	return taskIdToActivity
}

func (rtm *RuntimeTaskManager) locateTask(
	evt codec.DpfEvent,
	rule vgrule.EngineOrder,
	matcher MatchPhysicalEngine,
	extraMatch MatchExtraConds,
) (rv rtdata.OrderTask, found bool) {
	idxStart := rtm.upperBoundForTaskVec(evt.Cycle)
	found = false
	vec := rtm.orderedTaskVector
	for j := idxStart - 1; j > idxStart-1-MAX_BACKTRACE_TASK_COUNT; j-- {
		if j < 0 || j >= len(vec) {
			continue
		}
		if !vec[j].IsValid() {
			continue
		}
		thatTask := vec[j]
		if matcher.DoMatchTo(evt, thatTask, rule) &&
			(extraMatch == nil || extraMatch(thatTask.GetExecUuid(), evt.PacketID)) {
			// Copy result
			rv = thatTask
			found = true
			break
		}
	}
	return
}

func (rtm *RuntimeTaskManager) GenerateKernelActs(
	kernelActs []rtdata.KernelActivity,
	opBundles []rtdata.OpActivity,
	rule vgrule.EngineOrder) []rtdata.KernelActivity {
	var sipTaskBingoCount = 0
	for i, kernAct := range kernelActs {
		assert.Assert(kernAct.Start.EngineTypeCode == codec.EngCat_SIP, "must be sip")
		taskObj, found := rtm.locateTask(
			kernAct.Start, rule, MatchToSip{}, nil)
		if found {
			kernelActs[i].RtInfo.TaskId = taskObj.GetTaskID()
			sipTaskBingoCount++
		}
	}
	fmt.Printf("SIP task bingo %v out of %v\n", sipTaskBingoCount,
		len(kernelActs),
	)
	return GenerateKerenlActSeq(kernelActs, opBundles, rtm)
}

func (rtm *RuntimeTaskManager) QuerySubOpName(
	taskId int,
	opId int,
	entityId int,
	subIdx int,
) (string, bool) {
	if _, ok := rtm.taskIdToTask[taskId]; ok {
		// execUuid := task.ExecutableUUID
		// TODO: find sub op thru execKnowledge
		// if rtm.execKnowledge == nil {
		// 	return "", false
		// }
		if rtm.tempInternal.subOpInformation == nil {
			rtm.tempInternal.subOpInformation = metadata.JsonLoader{}.LoadInfo("subops.json")
		}

		if rtm.tempInternal.subOpInformation != nil {
			that := rtm.tempInternal.subOpInformation[fmt.Sprintf("%v", opId)]
			if that != nil {
				entity := fmt.Sprintf("%v", entityId)
				subVec := that[entity]
				if subIdx >= 0 && subIdx < len(subVec) {
					return subVec[subIdx], true
				}
			}
		}
	}
	return "", false
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
	vec := newStatedOrderTaskVector(r.fullTaskVector)

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
			if taskInOrder.AbleToMatchCqm(curAct.Start, rule) {
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
			// assert.Assert(!everSearched || onceValid, "must be valid for once")
			_ = everSearched
			_ = onceValid
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
	vec := newStatedOrderTaskVector(r.orderedTaskVector)

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
			if taskInOrder.AbleToMatchCqm(curAct.Start, rule) {
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

func (rtm *RuntimeTaskManager) DumpInfos(orderTask []rtdata.OrderTaskStated) {
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
	vec := newStatedOrderTaskVector(rtm.orderedTaskVector)

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
			if !taskInOrder.AbleToMatchXDMA(curAct.Start, algo) {
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
