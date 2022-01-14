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
type DmaDetector struct{}

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

func (DmaDetector) GetEngineTypes() []EngineTypeCode {
	return []EngineTypeCode{
		EngCat_CDMA,
		EngCat_SDMA,
	}
}

// Master Word for CDMA/SDMA
// bit0: flag b'0
// bit1-2:  event
// bit3-7:  vc id (5bit)
// bit8: b'0
// bit9~31(23 bit packet id)
func (DmaDetector) IsStarterMark(evt DpfEvent) (bool, bool) {
	return evt.Event&3 == DmaVcExecStart, evt.Event&3 == DmaVcExecEnd
}

func getVcVal(v int) int {
	const VC_BITCOUNT = 6
	// Shift 1 to elide the start/end flag
	// and event bits are reduced from 2 bits to 1 bit
	// Plus the Vc bits, to form the mask bit (1+VC_BITCOUNT)
	return (v >> 1) & ((1 << (VC_BITCOUNT + 1)) - 1)
}

func (DmaDetector) TestIfMatch(former, latter DpfEvent) bool {
	return former.EngineTypeCode == latter.EngineTypeCode &&
		former.PacketID == latter.PacketID &&
		getVcVal(former.Event) == getVcVal(latter.Event) &&
		former.ClusterID == latter.ClusterID
}
