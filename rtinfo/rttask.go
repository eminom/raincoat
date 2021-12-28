package rtinfo

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/algo/linklist"
	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/meta"
)

var (
	ErrNoExecMeta = errors.New("no exec meta info for runtime")
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

type RuntimeTaskManager struct {
	taskIdToTask map[int]*RuntimeTask
	taskIdVec    []int
	tsHead       *linklist.Lnk

	execKnowledge     *meta.ExecRaw
	orderedTaskVector []OrderTask
}

func NewRuntimeTaskManager() *RuntimeTaskManager {
	return &RuntimeTaskManager{
		tsHead: linklist.NewLnkHead(),
	}
}

func (self *RuntimeTaskManager) LoadRuntimeTask(filename string) bool {
	fin, err := os.Open(filename)
	if err != nil {
		log.Printf("error load runtime info:%v\n", err)
		return false
	}
	defer fin.Close()

	dc := make(map[int]*RuntimeTask)
	var taskSequentials []int
	scan := bufio.NewScanner(fin)
	for {
		if !scan.Scan() {
			break
		}
		line := scan.Text()
		vs := strings.Split(line, " ")
		taskId, err := strconv.Atoi(vs[0])
		if err != nil {
			log.Printf("error read '%v'", line)
			continue
		}

		if _, ok := dc[taskId]; ok {
			panic("error runtimetask: duplicate task id")
		}

		hxVal := vs[1]
		if strings.HasPrefix(hxVal, "0x") || strings.HasPrefix(hxVal, "0X") {
			hxVal = hxVal[2:]
		}
		exec, err := strconv.ParseUint(hxVal, 16, 64)
		if err != nil {
			log.Printf("error exec: %v", vs[1])
		}
		pgMask, err := strconv.Atoi(vs[2])
		if err != nil {
			log.Printf("error read pg mask: %v", err)
		}
		dc[taskId] = &RuntimeTask{
			TaskID:         taskId,
			ExecutableUUID: exec,
			PgMask:         pgMask,
		}
		taskSequentials = append(taskSequentials, taskId)
	}
	sort.Ints(taskSequentials)
	// update to self
	self.taskIdToTask, self.taskIdVec = dc, taskSequentials
	return true
}

func (r *RuntimeTaskManager) CollectTsEvent(evt codec.DpfEvent) {
	if evt.Event == codec.TsLaunchCqmStart {
		r.tsHead.AppendNode(evt)
		return
	}
	if evt.Event == codec.TsLaunchCqmEnd {
		if start := r.tsHead.Extract(func(one interface{}) bool {
			un := one.(codec.DpfEvent)
			return un.Payload == evt.Payload
		}); start != nil {
			startUn := start.(codec.DpfEvent)
			taskID := startUn.Payload
			if task, ok := r.taskIdToTask[taskID]; !ok {
				panic(fmt.Errorf("no start for cqm launch exec"))
			} else {
				task.StartCycle = startUn.Cycle
				task.EndCycle = evt.Cycle
				task.CycleValid = true
			}
		}
		return
	}
}

func (r RuntimeTaskManager) DumpInfo() {
	if r.tsHead.ElementCount() > 0 {
		fmt.Fprintf(os.Stderr, "TS unmatched count: %v\n",
			r.tsHead.ElementCount())
	}
	fmt.Printf("# runtimetask:\n")
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
	var orders []OrderTask
	for _, task := range r.taskIdToTask {
		if task.CycleValid && task.MetaValid {
			orders = append(orders, NewOrderTask(
				task.StartCycle,
				task,
			))
		}
	}
	sort.Sort(OrderTasks(orders))
	r.orderedTaskVector = orders

	log.Printf("%v order task has been built", len(orders))
}

// LoadMeta will load executable raw from task info's executable-uuids
func (r *RuntimeTaskManager) LoadMeta(startPath string) {
	execKm := meta.NewExecRaw(startPath)
	for _, taskId := range r.taskIdVec {
		if r.taskIdToTask[taskId].CycleValid {
			if execKm.LoadMeta(r.taskIdToTask[taskId].ExecutableUUID) {
				r.taskIdToTask[taskId].MetaValid = true
			}
		}
	}
	execKm.LoadWildcard()
	execKm.DumpInfo()
	r.execKnowledge = execKm
}

func (r *RuntimeTaskManager) FindExecFor(execUuid uint64) meta.ExecScope {
	exec, ok := r.execKnowledge.FindExecScope(execUuid)
	assert.Assert(ok, "Must be there")
	return exec
}

func (r *RuntimeTaskManager) LookupOpIdByPacketID(
	execUuid uint64,
	packetId int,
) (meta.DtuOp, error) {
	if r.execKnowledge == nil {
		return meta.DtuOp{}, ErrNoExecMeta
	}
	exec, ok := r.execKnowledge.FindExecScope(execUuid)
	if !ok {
		log.Printf("exec %016x is not loaded", exec)
		return meta.DtuOp{}, ErrNoExecMeta
	}
	return exec.FindOp(packetId)
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

// CookCqm:  find dtu-op meta information for the Cqm Act
func (r *RuntimeTaskManager) CookCqm(dtuBundle []CqmActBundle) []CqmActBundle {
	// Each time we start processing a new session
	// We create a new object to do the math
	vec := r.orderedTaskVector
	for i := 0; i < len(vec); i++ {
		vec[i].taskState = NewOrderTaskState()
	}

	bingoCount := 0
	unprocessedVec := []CqmActBundle{}
	for i := 0; i < len(dtuBundle); i++ {
		curAct := &dtuBundle[i]
		start := curAct.StartCycle()
		idxStart := r.upperBoundForTaskVec(start)
		// backtrace for no more than 5
		const maxBacktraceTaskCount = 2
		found := false
		for _, j := range []int{idxStart - 1} {
			if j < 0 || j >= len(vec) {
				continue
			}
			taskInOrder := vec[j]
			if !taskInOrder.IsValid() {
				continue
			}
			thisExecUuid := taskInOrder.refToTask.ExecutableUUID
			if taskInOrder.AbleToMatchCqm(*curAct) {
				if opInfo, err := r.LookupOpIdByPacketID(
					thisExecUuid,
					curAct.Start.PacketID); err == nil {
					// There is always a dtuop related to dbg op
					// and there is always a task
					taskInOrder.SuccessMatchDtuop(curAct.Start.PacketID)
					taskInOrder.SuccessMatchDtuop(curAct.End.PacketID)
					curAct.opRef = OpRef{
						dtuOp:     &opInfo,
						refToTask: taskInOrder.refToTask,
					}
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
			unprocessedVec = append(unprocessedVec, CqmActBundle{
				DpfAct: curAct.DpfAct,
			})
		}
	}
	fmt.Printf("Dbg op/Dtu-op meta success matched count: %v out of %v\n",
		bingoCount,
		len(dtuBundle),
	)

	fmt.Printf("statistics for ordered-task\n")
	for _, v := range vec {
		execScope := r.FindExecFor(v.refToTask.ExecutableUUID)
		v.DumpInfo(execScope)
	}

	return unprocessedVec
}

// CookCqm:  find dtu-op meta information for the Cqm Act
func (r *RuntimeTaskManager) OvercookCqm(
	dtuBundle []CqmActBundle,
) {
	// Each time we start processing a new session
	// We create a new object to do the math
	vec := r.orderedTaskVector
	for i := 0; i < len(vec); i++ {
		vec[i].taskState = NewOrderTaskState()
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

	for i := 0; i < len(dtuBundle); i++ {
		curAct := &dtuBundle[i]
		found := false
		for j := len(vec) - 1; j >= 0; j-- {
			if j < 0 || j >= len(vec) {
				continue
			}
			taskInOrder := vec[j]
			if !taskInOrder.IsValid() {
				continue
			}
			if taskInOrder.AbleToMatchCqm(*curAct) {
				thisExecUuid := taskInOrder.refToTask.ExecutableUUID
				opInfo, err := r.LookupOpIdByPacketID(
					thisExecUuid,
					curAct.Start.PacketID)
				switch err {
				case nil:
					// There is always a dtuop related to dbg op
					// and there is always a task
					taskInOrder.SuccessMatchDtuop(curAct.Start.PacketID)
					taskInOrder.SuccessMatchDtuop(curAct.End.PacketID)
					curAct.opRef = OpRef{
						dtuOp:     &opInfo,
						refToTask: taskInOrder.refToTask,
					}
					found = true
				case meta.ErrInvalidPacketId:
				case meta.ErrValidPacketIdNoOp:
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

// CookCqm:  find dtu-op meta information for the Cqm Act
func (r *RuntimeTaskManager) WildCookCqm(
	dtuBundle []CqmActBundle,
) []CqmActBundle {
	var rv []CqmActBundle
	log.Printf("start wild cook for %v item(s)", len(dtuBundle))
	for i := 0; i < len(dtuBundle); i++ {
		curAct := &dtuBundle[i]
		// start := curAct.StartCycle()
		if _, ok := r.execKnowledge.LookForWild(curAct.Start.PacketID, false); ok {
			rv = append(rv, CqmActBundle{
				DpfAct: curAct.DpfAct,
			})
		}
	}
	log.Printf("done wild cook, get %v item(s)", len(rv))
	if len(dtuBundle) == len(rv) {
		log.Printf("all matched in wild match mode")
	}
	return rv
}
