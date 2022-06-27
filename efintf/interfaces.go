package efintf

import (
	"git.enflame.cn/hai.bai/dmaster/efintf/affinity"
	"git.enflame.cn/hai.bai/dmaster/meta/dtuarch"
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/topsdev/mimic/mimicdefs"
)

type RingBufferLoader interface {
	LoadRingBufferContent(cid int, ringbufferIdx int) []byte
	GetRingBufferCount() int
	GetInputName() string
}

type CpuOpTraceLoader interface {
	GetCpuOpTraceSeq() []rtdata.CpuOpAct
}

type TaskLoader interface {
	LoadTask(oneSolid bool) (map[int]*rtdata.RuntimeTask, []int, bool)
}

type ArchTypeGet interface {
	GetArchType() dtuarch.ArchType
}

type InfoReceiver interface {
	TaskLoader
	ArchTypeGet
	GetCdmaAffinity() affinity.CdmaAffinitySet
	LoadTimepoints() ([]rtdata.HostTimeEntry, bool)
	LoadExecScope(execUuid uint64) *metadata.ExecScope
	LoadWildcards(checkExist func(str string) bool, notifyNew func(uint64, *metadata.ExecScope))
	ExtractHostInfo() *mimicdefs.HostInfo
}
