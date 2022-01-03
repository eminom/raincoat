package efintf

import (
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type InfoReceiver interface {
	LoadRingBufferContent(cid int) []byte
	LoadTask() (map[int]*rtdata.RuntimeTask, []int, bool)
	LoadTimepoints() ([]rtdata.HostTimeEntry, bool)
	LoadExecScope(execUuid uint64) *metadata.ExecScope
	LoadWildcards(checkExist func(str string) bool, notifyNew func(uint64, *metadata.ExecScope))
}
