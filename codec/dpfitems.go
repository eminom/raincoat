package codec

import (
	"errors"
	"fmt"
)

var (
	malFormattedError = errors.New("mal-formatted")
	decodeErr         = errors.New("decode error")
)

type DpfItem struct {
	RawVale [4]uint32

	Flag     int
	PacketID int
	Event    int
	Context  int
	Payload  int
	Cycle    uint64

	EngineTy    string
	EngineIndex int
}

func (d DpfItem) ToString() string {
	if d.Flag == 0 {
		return fmt.Sprintf("%-10v %-2v %-2v event=%-4v pid=%v ts=%08x",
			d.EngineTy, d.EngineIndex, d.Context, d.Event, d.PacketID, d.Cycle)
	}
	return fmt.Sprintf("%-6v %-2v event=%v payload=%v ts=%08x",
		d.EngineTy, d.EngineIndex, d.Event, d.Payload, d.Cycle)
}

func (decoder *DecodeMaster) NewDpfItem(vals []uint32) (DpfItem, error) {
	if len(vals) != 4 {
		panic(malFormattedError)
	}
	ts := uint64(vals[2]) + uint64(vals[3])<<32
	if vals[0]&1 == 0 {
		// uint32_t flag_ : 1;  // should be always 0
		// uint32_t event_ : 8;
		// uint32_t packet_id_ : 23;
		event := (vals[0] >> 1) & 0xFF
		packet_id := (vals[0] >> 9)
		engIdx, ctx, engTy, ok := decoder.GetEngineInfo(vals[1])
		if !ok {
			return DpfItem{}, decodeErr
		}

		return DpfItem{
			Flag:        0,
			PacketID:    int(packet_id),
			Event:       int(event),
			Context:     ctx,
			EngineTy:    engTy,
			EngineIndex: engIdx,
			Cycle:       ts,
		}, nil

	}
	// uint32_t flag_ : 1;  // should be always 1
	// uint32_t event_ : 7;
	// uint32_t payload_ : 24;
	event := (vals[0] >> 1) & 0x7F
	payload := (vals[0] >> 8)
	engineIdx, engTy, ok := decoder.GetEngineInfoV2(vals[1])
	if !ok {
		return DpfItem{}, decodeErr
	}
	return DpfItem{
		Flag:        1,
		Event:       int(event),
		Payload:     int(payload),
		EngineTy:    engTy,
		EngineIndex: engineIdx,
	}, nil
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
