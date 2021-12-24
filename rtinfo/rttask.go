package rtinfo

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/algo/linklist"
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/meta"
)

type RuntimeTask struct {
	TaskID         int
	ExecutableUUID uint64
	PgMask         int

	StartCycle uint64
	EndCycle   uint64
	Valid      bool
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

type RuntimeTaskManager struct {
	taskIdToTask  map[int]*RuntimeTask
	tsHead        *linklist.Lnk
	seq           []int
	execKnowledge *meta.ExecRaw
}

func LoadRuntimeTask(filename string) *RuntimeTaskManager {
	fin, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer fin.Close()

	dc := make(map[int]*RuntimeTask)
	var seq []int
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
		seq = append(seq, taskId)
	}
	sort.Ints(seq)
	return &RuntimeTaskManager{
		taskIdToTask: dc,
		tsHead:       linklist.NewLnkHead(),
		seq:          seq,
	}
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
				task.Valid = true
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
	for _, taskId := range r.seq {
		if r.taskIdToTask[taskId].Valid {
			fmt.Printf("%v\n", r.taskIdToTask[taskId].ToString())
		}
	}
	fmt.Println()
}

func (r *RuntimeTaskManager) LoadMeta(startPath string) {
	execKm := meta.NewExecRaw(startPath)
	for _, taskId := range r.seq {
		if r.taskIdToTask[taskId].Valid {
			if execKm.LoadMeta(r.taskIdToTask[taskId].ExecutableUUID) {
				r.taskIdToTask[taskId].MetaValid = true
			}
		}
	}
	r.execKnowledge = execKm
}
