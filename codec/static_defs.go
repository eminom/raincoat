package codec

// Faster implementation if number of total events is small
var (
	dmaEventStringArray = []string{
		"ENG_START",
		"ENG_END",
		"VC_START",
		"VC_END",
	}
)

func ToDmaEventStr(evtID int) string {
	return dmaEventStringArray[evtID&3]
}
