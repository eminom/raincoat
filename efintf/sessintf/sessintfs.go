package sessintf

import "git.enflame.cn/hai.bai/dmaster/codec"

type EventSinker interface {
	GetEngineTypeCodes() []codec.EngineTypeCode
	DispatchEvent(codec.DpfEvent) error
}

type ConcurEventSinker interface {
	EventSinker
	SelfClone() ConcurEventSinker
	MergeTo(interface{}) bool
}
