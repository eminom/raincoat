package codec

type DpfItems []DpfEvent

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
