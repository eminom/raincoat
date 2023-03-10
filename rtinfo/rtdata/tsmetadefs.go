package rtdata

const TS_EVENT_DEFS = `
// TS DPF events
typedef enum {
  TS_CMD_PACKET_START = 0x1,
  TS_CMD_PACKET_END = 0x0,
  TS_PARSE_STREAM_START = 3,
  TS_PARSE_STREAM_END = 2,
  TS_READ_PACKET_START = 5,
  TS_READ_PACKET_END = 4,
  TS_PARSE_GDMA_START = 7,
  TS_PARSE_GDMA_END = 6,
  TS_ISR_START = 9,
  TS_ISR_END = 8,
  TS_REGISTER_WRITE_START = 11,
  TS_REGISTER_WRITE_END = 10,
  TS_MEMORY_WRITE_START = 13,
  TS_MEMORY_WRITE_END = 12,
  TS_REGISTER_WAIT_START = 15,
  TS_REGISTER_WAIT_END = 14,
  TS_MEMORY_WAIT_START = 17,
  TS_MEMORY_WAIT_END = 16,
  TS_VG_CONFIG_START = 19,
  TS_VG_CONFIG_END = 18,
  TS_PARSE_EDMA_START = 21,
  TS_PARSE_EDMA_END = 20,
  TS_CQM_EXECUTABLE_LAUNCH_START = 23,
  TS_CQM_EXECUTABLE_LAUNCH_END = 22,
  TS_HCVG_EXECUTABLE_LAUNCH_START = 25,
  TS_HCVG_EXECUTABLE_LAUNCH_END = 24,
  TS_VDEC_EXECUTABLE_LAUNCH_START = 27,
  TS_VDEC_EXECUTABLE_LAUNCH_END = 26,
  TS_WAIT_STREAM_START = 29,
  TS_WAIT_STREAM_END = 28,
  TS_RECORD_STREAM_START = 31,
  TS_RECORD_STREAM_END = 30
} TS_DPF_EVENT_T;`

var (
	tsEventNameKeeper EventNameKeeper = NewEventNameKeeper(TS_EVENT_DEFS)
)

func ToTSEventString(evtID int) (string, bool) {
	if name, ok := tsEventNameKeeper.nameDict[evtID]; ok {
		return name, true
	}
	return "", false
}
