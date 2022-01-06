package codec

const (
	CqmEventOpStart = 9
	CqmEventOpEnd   = 8

	CqmEventCmdPacketStart = 7
	CqmEventCmdPacketEnd   = 6

	CqmExecutableStart = 3
	CqmExecutableEnd   = 2

	CqmDbgPacketStepStart = 0xb
	CqmDbgPacketStepEnd   = 0xa
)

const (
	TsLaunchCqmStart = 23
	TsLaunchCqmEnd   = 22
)

const (
	DmaBusyStart   = 0
	DmaBusyEnd     = 1
	DmaVcExecStart = 2
	DmaVcExecEnd   = 3
)

func IsCqmOpEvent(evt DpfEvent) bool {
	return (evt.EngineTypeCode == EngCat_CQM || evt.EngineTypeCode == EngCat_GSYNC) &&
		(evt.Event == CqmEventOpStart || evt.Event == CqmEventOpEnd)
}
