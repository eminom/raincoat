package codec

func IsDebugOpPacket(evt DpfEvent) bool {
	switch evt.EngineTypeCode {
	case EngCat_GSYNC, EngCat_CQM:
		return evt.Event == CqmEventOpStart
	}
	return false
}
