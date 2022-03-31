package efintf

import (
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type RingBufferLoader interface {
	LoadRingBufferContent(cid int, ringbufferIdx int) []byte
	GetRingBufferCount() int
}

type TaskLoader interface {
	LoadTask(oneSolid bool) (map[int]*rtdata.RuntimeTask, []int, bool)
}

type InfoReceiver interface {
	TaskLoader
	LoadTimepoints() ([]rtdata.HostTimeEntry, bool)
	LoadExecScope(execUuid uint64) *metadata.ExecScope
	LoadWildcards(checkExist func(str string) bool, notifyNew func(uint64, *metadata.ExecScope))
}
