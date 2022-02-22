package main

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type normalDumper struct {
	taskToNameMap map[int]map[string]int
}

func NewNormalDumper() *normalDumper {
	return &normalDumper{
		taskToNameMap: make(map[int]map[string]int),
	}
}

func (d *normalDumper) GetPidAndName(act rtdata.OpActivity) (bool, string, string, string) {
	if act.IsOpRefValid() {

		taskId := act.GetTask().TaskID
		var nameMap map[string]int
		if nm, ok := d.taskToNameMap[taskId]; ok {
			nameMap = nm
		} else {
			nameMap = make(map[string]int)
			d.taskToNameMap[taskId] = nameMap
		}

		opMeta := act.GetOp()
		opNameWithOpId := fmt.Sprintf("%v.%v", opMeta.OpName, opMeta.OpId)
		uniqueOpName := opNameWithOpId
		var subSeq int
		var ok bool
		if subSeq, ok = nameMap[opNameWithOpId]; ok {
			subSeq++
			uniqueOpName += fmt.Sprintf(".%v", subSeq)
		}
		nameMap[opNameWithOpId] = subSeq
		return true,
			act.GetTask().ToShortString(), // Process name
			opMeta.OpName, // Thread name
			uniqueOpName // Thread name with op id
	}
	return false, "Unknown Task", "Unk", "Unk"
}

type wildInDumper struct {
	notWildInCount int
}

func (w *wildInDumper) GetPidAndName(act rtdata.OpActivity) (bool, string, string, string) {
	//+ act.GetTask().ToShortString(),
	if act.IsOpRefValid() {
		rawOpName := act.GetOp().OpName
		opId := act.GetOp().OpId
		return true, "Wild In",
			rawOpName,
			fmt.Sprintf("%v.%v", rawOpName, opId)
	}
	w.notWildInCount++
	return false, "", "", ""
}

type wildOutDumper struct {
	subSampleCc int
}

func (w *wildOutDumper) GetPidAndName(act rtdata.OpActivity) (bool, string, string, string) {
	//+ act.GetTask().ToShortString(),
	w.subSampleCc++
	if w.subSampleCc%17 == 0 {
		return true, "Wild Out", "some op", "some op"
	}
	return false, "", "", ""
}
