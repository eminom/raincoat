package codec

func IsDebugOpPacket(evt DpfEvent) bool {
	switch evt.EngineTypeCode {
	case EngCat_GSYNC, EngCat_CQM:
		return evt.Event == CqmEventOpStart
	}
	return false
}

func FirmwareEventFilter(evt DpfEvent) bool {
	switch evt.Event {
	case CqmEventCmdPacketStart,
		CqmEventOpStart,
		CqmDbgPacketStepStart,
		CqmExecutableStart,
		TsLaunchCqmStart:
		return true
	}
	return false
}
