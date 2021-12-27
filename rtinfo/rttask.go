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
	errNoMeta       = errors.New("no meta info for runtime")
	errNoSuchPacket = errors.New("no such packet")
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
	return fmt.Sprintf("PG %v Task %v",
		r.PgMask,
		r.TaskID,
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
		return meta.DtuOp{}, errNoMeta
	}
	exec, ok := r.execKnowledge.FindExecScope(execUuid)
	if !ok {
		log.Printf("exec %016x is not loaded", exec)
		return meta.DtuOp{}, errNoMeta
	}
	if dtuOp, ok := exec.FindOp(packetId); ok {
		return dtuOp, nil
	}
	return meta.DtuOp{}, fmt.Errorf("no packet %v in %016x",
		packetId,
		execUuid,
	)
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
func (r *RuntimeTaskManager) CookCqm(dtuBundle []CqmActBundle) {
	// Each time we start processing a new session
	// We create a new object to do the math
	vec := r.orderedTaskVector
	for i := 0; i < len(vec); i++ {
		vec[i].taskState = NewOrderTaskState()
	}

	bingoCount := 0
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
}
