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
type DmaDetector struct {
	IdmaPrefetchCount int
}
type SipDetector struct{}
type TaskDetector struct{}

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
		CqmSleepStart,
		TsLaunchCqmStart:
		return true
	}
	return false
}

// No terminator for fw acts
func (FwPktDetector) IsStarterMark(evt DpfEvent) (bool, bool, bool) {
	return isFwInterested(evt.Event),
		(evt.Event&1) == 0 && isFwInterested(evt.Event+1),
		false
}

// We only care about the unmatched cqm executable start
func (FwPktDetector) IsRecyclable(evt DpfEvent) bool {
	return evt.Event == CqmExecutableStart
}

func (FwPktDetector) TestIfMatch(former, latter DpfEvent) bool {
	if former.EngineTypeCode != latter.EngineTypeCode ||
		former.Event-1 != latter.Event {
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

func (FwPktDetector) PurgePreviousEvents() bool { return false }

func (DbgPktDetector) GetEngineTypes() []EngineTypeCode {
	return []EngineTypeCode{
		EngCat_CQM,
		EngCat_GSYNC,
	}
}

// func (dbgD DbgPktDetector) IsTerminator(evt DpfEvent) bool {
// 	return dbgD.PurgeOnStepEnd && evt.Event == CqmEventDebugPacketStepEnd
// }

func (DbgPktDetector) IsTerminatorMark(evt DpfEvent) bool {
	return evt.Event == CqmEventDebugPacketStepEnd
}

func (DbgPktDetector) IsStarterMark(evt DpfEvent) (bool, bool, bool) {
	return evt.Event == CqmEventOpStart,
		evt.Event == CqmEventOpEnd,
		evt.Event == CqmEventDebugPacketStepEnd
}

// We only care about the unmatched cqm executable start
func (DbgPktDetector) IsRecyclable(DpfEvent) bool {
	return false
}

func (DbgPktDetector) TestIfMatch(former, latter DpfEvent) bool {
	return former.EngineTypeCode == latter.EngineTypeCode &&
		former.PacketID+1 == latter.PacketID
}

func (DbgPktDetector) PurgePreviousEvents() bool { return false }

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
// And there is no terminator for Dma acts
func (dmaD *DmaDetector) IsStarterMark(evt DpfEvent) (bool, bool, bool) {
	evtCode := evt.Event & 3
	mustIgnore := GetDmaVcId(evt.Event) == 16 && evt.EngineTypeCode == EngCat_SDMA
	if mustIgnore {
		dmaD.IdmaPrefetchCount++
	}
	return !mustIgnore && (evtCode == DmaVcExecStart || evtCode == DmaBusyStart),
		!mustIgnore && (evtCode == DmaVcExecEnd || evtCode == DmaBusyEnd),
		false
}

func (DmaDetector) IsRecyclable(DpfEvent) bool {
	return false
}

const VC_BITCOUNT = 6

func getVcVal(v int) int {
	// Shift 1 to elide the start/end flag
	// and event bits are reduced from 2 bits to 1 bit
	// Plus the Vc bits, to form the mask bit (1+VC_BITCOUNT)
	return (v >> 1) & ((1 << (VC_BITCOUNT + 1)) - 1)
}

func GetDmaVcId(evtVal int) int {
	const mask = (1 << VC_BITCOUNT) - 1
	return (evtVal >> 2) & mask
}

func getXdmaEvt(dpf DpfEvent) int {
	return (dpf.Event >> 1) & 1
}

func (DmaDetector) TestIfMatch(former, latter DpfEvent) bool {
	return former.EngineTypeCode == latter.EngineTypeCode &&
		former.PacketID == latter.PacketID &&
		getXdmaEvt(former) == getXdmaEvt(latter) &&
		getVcVal(former.Event) == getVcVal(latter.Event)
}

func (DmaDetector) PurgePreviousEvents() bool { return true }

func (SipDetector) GetEngineTypes() []EngineTypeCode {
	return []EngineTypeCode{
		EngCat_SIP,
	}
}

// Only two events
//
func (SipDetector) IsStarterMark(evt DpfEvent) (bool, bool, bool) {
	return evt.Event == 1,
		evt.Event == 0,
		false
}

func (SipDetector) IsRecyclable(DpfEvent) bool {
	return false
}

// ALL SIP, Engine Type must be the same
// 		former.EngineTypeCode == latter.EngineTypeCode
// Packet id is now all zero for SIP events
//   	former.PacketID == latter.PacketID &&
func (SipDetector) TestIfMatch(former, latter DpfEvent) bool {
	// Always match the closest begin
	return true
}

func (SipDetector) PurgePreviousEvents() bool { return false }

func (TaskDetector) GetEngineTypes() []EngineTypeCode {
	return []EngineTypeCode{
		EngCat_TS,
	}
}
func (TaskDetector) IsStarterMark(evt DpfEvent) (bool, bool, bool) {
	return evt.Flag == 1 && evt.Event == TsLaunchCqmStart,
		evt.Flag == 1 && evt.Event == TsLaunchCqmEnd,
		false
}

func (TaskDetector) IsRecyclable(DpfEvent) bool {
	return false
}

func (TaskDetector) TestIfMatch(former, latter DpfEvent) bool {
	return true // always match
}

func (TaskDetector) PurgePreviousEvents() bool { return false }
