package codec

const (
	CqmEventOpStart = 9
	CqmEventOpEnd   = 8

	CqmEventCmdPacketStart = 7
	CqmEventCmdPacketEnd   = 6
)

const (
	TsLaunchCqmStart = 23
	TsLaunchCqmEnd   = 22
)

func IsCqmOpEvent(evt DpfEvent) bool {
	return (evt.EngineTypeCode == EngCat_CQM || evt.EngineTypeCode == EngCat_GSYNC) &&
		(evt.Event == CqmEventOpStart || evt.Event == CqmEventOpEnd)
}
