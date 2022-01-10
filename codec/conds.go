package codec

func isDebugOpPacket(evt DpfEvent) bool {
	switch evt.EngineTypeCode {
	case EngCat_GSYNC, EngCat_CQM:
		return evt.Event == CqmEventOpStart ||
			evt.Event == CqmEventDebugPacketStepStart
	}
	return false
}

type FwPktDetector struct{}
type DbgPktDetector struct{}

// Implementations

func (FwPktDetector) GetEngineTypes() []EngineTypeCode {
	return []EngineTypeCode{
		EngCat_CQM,
		EngCat_GSYNC,
		EngCat_TS,
	}
}

func isFwInterested(event int) bool {
	switch event {
	case CqmEventCmdPacketStart,
		CqmEventOpStart,
		CqmDbgPacketStepStart,
		CqmExecutableStart,
		TsLaunchCqmStart:
		return true
	}
	return false
}

func (FwPktDetector) IsStarterMark(evt DpfEvent) (bool, bool) {
	return isFwInterested(evt.Event),
		(evt.Event&1) == 0 && isFwInterested(evt.Event+1)
}

func (FwPktDetector) TestIfMatch(former, latter DpfEvent) bool {
	if former.EngineTypeCode != latter.EngineTypeCode ||
		former.Event-1 != latter.Event || // event in pairs
		former.ClusterID != latter.ClusterID {
		return false
	}
	if isDebugOpPacket(former) {
		return former.PacketID+1 == latter.PacketID
	}
	// if former.EngineTypeCode == EngCat_CQM &&
	// 	former.Event == CqmEventCmdPacketStart {
	// 	return former.PacketID == latter.PacketID
	// }
	// Or the rest of the event must be of the same packet-ID??
	return former.PacketID == latter.PacketID
}

func (DbgPktDetector) GetEngineTypes() []EngineTypeCode {
	return []EngineTypeCode{
		EngCat_CQM,
		EngCat_GSYNC,
	}
}

func (DbgPktDetector) IsStarterMark(evt DpfEvent) (bool, bool) {
	return evt.Event == CqmEventOpStart, evt.Event == CqmEventOpEnd
}

func (DbgPktDetector) TestIfMatch(former, latter DpfEvent) bool {
	return former.EngineTypeCode == latter.EngineTypeCode &&
		former.PacketID+1 == latter.PacketID &&
		former.ClusterID == latter.ClusterID
}
