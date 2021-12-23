package codec

import (
	"errors"
	"fmt"
)

var (
	errMalFormattedError = errors.New("mal-formatted")
	errDpfItemDecodeErr  = errors.New("decode error")
)

type DpfItem struct {
	RawVale [4]uint32

	Flag     int
	PacketID int
	Event    int
	Context  int
	Payload  int
	Cycle    uint64

	EngineTy      string
	EngineUniqIdx int
	EngineIndex   int
}

func (d DpfItem) ToString() string {
	if d.Flag == 0 {
		switch d.EngineTy {
		case ENGINE_PCIE:
			return fmt.Sprintf("%-10v %-10v ts=%08x",
				d.EngineTy, d.RawVale[0]>>1, d.Cycle)
		}
		return fmt.Sprintf("%-10v %-2v %-2v event=%-4v pid=%v ts=%08x",
			d.EngineTy, d.EngineIndex, d.Context, d.Event, d.PacketID, d.Cycle)
	}
	return fmt.Sprintf("%-6v %-2v event=%v payload=%v ts=%08x",
		d.EngineTy, d.EngineIndex, d.Event, d.Payload, d.Cycle)
}

func copyFrom(vals []uint32) [4]uint32 {
	var buff [4]uint32
	copy(buff[:], vals)
	return buff
}

// helper API for format V1
func (decoder *DecodeMaster) createFormatV1(vals []uint32) (DpfItem, error) {
	ts := uint64(vals[2]) + uint64(vals[3])<<32
	// flag_ : 1;  // should be always 0
	// event_ : 8;
	// packet_id_ : 23;
	event := (vals[0] >> 1) & 0xFF
	packet_id := (vals[0] >> 9)
	engIdx, ctx, engUniqIdx, ok := decoder.GetEngineInfo(vals[1])
	if !ok {
		return DpfItem{}, errDpfItemDecodeErr
	}

	return DpfItem{
		RawVale:       copyFrom(vals),
		Flag:          0,
		PacketID:      int(packet_id),
		Event:         int(event),
		Context:       ctx,
		EngineUniqIdx: engUniqIdx,
		EngineTy:      decoder.EngUniqueIndexToTypeName(engUniqIdx),
		EngineIndex:   engIdx,
		Cycle:         ts,
	}, nil
}

// helper API for format V2
func (decoder *DecodeMaster) createFormatV2(vals []uint32) (DpfItem, error) {
	// flag_ : 1; // should always be 1
	// event_ : 7;
	// payload_ : 24;
	ts := uint64(vals[2]) + uint64(vals[3])<<32
	event := (vals[0] >> 1) & 0x7F
	payload := (vals[0] >> 8)
	engineIdx, engUniqIdx, ok := decoder.GetEngineInfoV2(vals[1])
	if !ok {
		return DpfItem{}, errDpfItemDecodeErr
	}
	return DpfItem{
		RawVale:       copyFrom(vals),
		Flag:          1,
		Event:         int(event),
		Payload:       int(payload),
		EngineUniqIdx: engUniqIdx,
		EngineTy:      decoder.EngUniqueIndexToTypeName(engUniqIdx),
		EngineIndex:   engineIdx,
		Cycle:         ts,
	}, nil
}

func (decoder *DecodeMaster) NewDpfItem(vals []uint32) (DpfItem, error) {
	if len(vals) != 4 {
		panic(errMalFormattedError)
	}
	if vals[0]&1 == 0 {
		return decoder.createFormatV1(vals)
	}
	return decoder.createFormatV2(vals)
}

type DpfItems []DpfItem

func (d DpfItems) Len() int {
	return len(d)
}

func (d DpfItems) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d DpfItems) Less(i, j int) bool {
	lhs, rhs := d[i], d[j]
	if lhs.PacketID != rhs.PacketID {
		return lhs.PacketID < rhs.PacketID
	}
	if lhs.Event != rhs.Event {
		return lhs.Event < rhs.Event
	}
	return lhs.Cycle < rhs.Cycle
}
