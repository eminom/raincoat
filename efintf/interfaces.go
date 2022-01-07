package efintf

import (
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type RingBufferLoader interface {
	LoadRingBufferContent(cid int) []byte
	HasMore() bool
}

type InfoReceiver interface {
	LoadTask() (map[int]*rtdata.RuntimeTask, []int, bool)
	LoadTimepoints() ([]rtdata.HostTimeEntry, bool)
	LoadExecScope(execUuid uint64) *metadata.ExecScope
	LoadWildcards(checkExist func(str string) bool, notifyNew func(uint64, *metadata.ExecScope))
}
