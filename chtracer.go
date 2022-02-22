package main

import (
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type normalDumper struct{}

func (normalDumper) GetPidAndName(act rtdata.OpActivity) (bool, string, string) {
	if act.IsOpRefValid() {
		opMeta := act.GetOp()
		opName := fmt.Sprintf("%v.%v", opMeta.OpName, opMeta.OpId)
		return true,
			act.GetTask().ToShortString(),
			opName
	}
	return false, "Unknown Task", "Unk"
}

type wildInDumper struct {
	notWildInCount int
}

func (w *wildInDumper) GetPidAndName(act rtdata.OpActivity) (bool, string, string) {
	//+ act.GetTask().ToShortString(),
	if act.IsOpRefValid() {
		return true, "Wild In",
			act.GetOp().OpName
	}
	w.notWildInCount++
	return false, "", ""
}

type wildOutDumper struct {
	subSampleCc int
}

func (w *wildOutDumper) GetPidAndName(act rtdata.OpActivity) (bool, string, string) {
	//+ act.GetTask().ToShortString(),
	w.subSampleCc++
	if w.subSampleCc%17 == 0 {
		return true, "Wild Out", "some op"
	}
	return false, "", ""
}
