package rtdata

const CQM_EVENT_DEFS = `
typedef enum {
	CQM_SLEEP_START = 0x1,
	CQM_SLEEP_END = 0x0,
	CQM_EXECUTABLE_START = 0x3,
	CQM_EXECUTABLE_END = 0x2,
	CQM_LOOP_TASK_START = 0x5,
	CQM_LOOP_TASK_END = 0x4,
	CQM_CMD_PACKET_START = 0x7,
	CQM_CMD_PACKET_END = 0x6,
	CQM_DBG_PACKET_OP_START = 0x9,
	CQM_DBG_PACKET_OP_END = 0x8,
	CQM_DBG_PACKET_STEP_START = 0xb,
	CQM_DBG_PACKET_STEP_END = 0xa,
	CQM_SIGNAL_COUNTER = 0xd,
	CQM_WAIT_COUNTER = 0xc,
	CQM_WRITE_MEMORY = 0xf,
	CQM_WAIT_MEMORY = 0xe,
  } CQM_DPF_EVENT_T;`

var (
	cqmEventNameKeeper EventNameKeeper = NewEventNameKeeper(CQM_EVENT_DEFS)
)

func ToCQMEventString(evtID int) (string, bool) {
	if name, ok := cqmEventNameKeeper.nameDict[evtID]; ok {
		return name, true
	}
	return "", false
}
