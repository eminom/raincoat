package codec

import (
	"errors"
	"fmt"
)

var (
	errMalFormattedError = errors.New("mal-formatted")
	errDpfItemDecodeErr  = errors.New("decode error")
)

const (
	MASTERVALUE_BITCOUNT = 10
	MASTERVALUE_COUNT    = 1 << MASTERVALUE_BITCOUNT

	RTCONTEXT_BITCOUNT = 4
	RTCONTEXT_COUNT    = 1 << RTCONTEXT_BITCOUNT
	// RTCONTEXT_BITMASK  = (1 << RTCONTEXT_BITCOUNT) - 1
)

type DpfEvent struct {
	RawValue [4]uint32

	Flag     int
	PacketID int
	Event    int
	Context  int
	Payload  int
	Cycle    uint64

	EngineTypeCode EngineTypeCode
	EngineUniqIdx  int
	EngineIndex    int
	ClusterID      int

	OffsetIndex int // Offset int the DPF buffer from the beginning
}

// This field is at-most 31-bit width
func (d DpfEvent) DpfSyncIndex() int {
	return int(d.RawValue[0] >> 1)
}

func (d DpfEvent) DpfSyncIndexMasked() int {
	val := int(d.RawValue[0])
	// 4 27 1
	val &= ^(0xF << 28)
	return val >> 1
}

func (d DpfEvent) DpfSyncProcId() int {
	val := int(d.RawValue[0])
	return 0xf & (val >> 28)
}

func (d DpfEvent) MasterIdValue() int {
	return int(d.RawValue[1]) & ((1 << MASTERVALUE_BITCOUNT) - 1)
}

func (d DpfEvent) IsOfEngine(engineTypeCode EngineTypeCode) bool {
	return d.EngineTypeCode == engineTypeCode
}

func toStartEndStr(evtFlag int) string {
	switch evtFlag {
	case 0:
		return "end"
	case 1:
		return "start"
	}
	return "unk"
}

func dmaEventTypeToString(evt int) string {
	if evt&2 == 2 {
		return "VC"
	}
	return "ENG"
}

func (d DpfEvent) ToString() string {
	if d.Flag == 0 {
		switch d.EngineTypeCode {
		case EngCat_PCIE:
			return fmt.Sprintf("%-10s %-10v ts=%v",
				d.EngineTypeCode,
				d.DpfSyncIndex(),
				d.Cycle)
		case EngCat_TS:
			return fmt.Sprintf("%-6s %-2v %-2v %-2v stream=%v %v pid=%v ts=%-14d",
				d.EngineTypeCode, d.ClusterID, d.EngineIndex, d.Context,
				d.Event>>1, toStartEndStr(d.Event&1), d.PacketID, d.Cycle,
			)
		case EngCat_CDMA, EngCat_SDMA:
			return fmt.Sprintf(
				"%-6s %-2v %-2v %-2v event=%-4v evt=%-9v vc=%-3v pid=%v  ts=%-14d",
				d.EngineTypeCode, d.ClusterID, d.EngineIndex, d.Context,
				d.Event,
				ToDmaEventStr(d.Event),
				GetDmaVcId(d.Event),
				d.PacketID, d.Cycle)
		}
		return fmt.Sprintf("%-10s %-2v %-2v %-2v event=%-4v pid=%v ts=%-14d",
			d.EngineTypeCode, d.ClusterID, d.EngineIndex, d.Context, d.Event, d.PacketID, d.Cycle)
	}
	return fmt.Sprintf("%-6s %-2v %-5v event=%-3v payload=%v ts=%-14d",
		d.EngineTypeCode, d.ClusterID, d.EngineIndex, d.Event, d.Payload, d.Cycle)
}

func (d DpfEvent) RawRepr() string {
	return fmt.Sprintf("[%08x: %08x %08x %08x %08x]",
		d.OffsetIndex*16,
		d.RawValue[0], d.RawValue[1],
		d.RawValue[2], d.RawValue[3])
}

func copyFrom(vals []uint32) [4]uint32 {
	var buff [4]uint32
	copy(buff[:], vals)
	return buff
}

// helper API for format V1
func (decoder *DecodeMaster) createFormatV1(vals []uint32) (DpfEvent, error) {
	ts := uint64(vals[2]) + uint64(vals[3])<<32
	// flag_ : 1;  // should be always 0
	// event_ : 8;
	// packet_id_ : 23;
	event := (vals[0] >> 1) & 0xFF
	packet_id := (vals[0] >> 9)
	engIdx, engUniqIdx, ctx, clusterID, ok := decoder.GetEngineInfo(vals[1])
	if !ok {
		return DpfEvent{}, errDpfItemDecodeErr
	}
	return DpfEvent{
		RawValue:       copyFrom(vals),
		Flag:           0,
		PacketID:       int(packet_id),
		Event:          int(event),
		Context:        ctx,
		EngineUniqIdx:  engUniqIdx,
		EngineTypeCode: decoder.EngUniqueIndexToTypeName(engUniqIdx),
		EngineIndex:    engIdx,
		Cycle:          ts,
		ClusterID:      clusterID,
	}, nil
}

// helper API for format V2
func (decoder *DecodeMaster) createFormatV2(vals []uint32) (DpfEvent, error) {
	// flag_ : 1; // should always be 1
	// event_ : 7;
	// payload_ : 24;
	ts := uint64(vals[2]) + uint64(vals[3])<<32
	event := (vals[0] >> 1) & 0x7F
	payload := (vals[0] >> 8)
	engineIdx, engUniqIdx, clusterID, ok := decoder.GetEngineInfoV2(vals[1])
	if !ok {
		return DpfEvent{}, errDpfItemDecodeErr
	}
	return DpfEvent{
		RawValue:       copyFrom(vals),
		Flag:           1,
		Event:          int(event),
		Payload:        int(payload),
		EngineUniqIdx:  engUniqIdx,
		EngineTypeCode: decoder.EngUniqueIndexToTypeName(engUniqIdx),
		EngineIndex:    engineIdx,
		Cycle:          ts,
		ClusterID:      clusterID,
	}, nil
}

func (decoder *DecodeMaster) NewDpfEvent(
	vals []uint32,
	offsetIdx int,
) (dpf DpfEvent, err error) {
	if len(vals) != 4 {
		panic(errMalFormattedError)
	}
	dpf, err = func() (DpfEvent, error) {
		if vals[0]&1 == 0 {
			return decoder.createFormatV1(vals)
		}
		return decoder.createFormatV2(vals)
	}()
	dpf.OffsetIndex = offsetIdx
	return
}
