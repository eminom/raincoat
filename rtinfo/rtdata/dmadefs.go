package rtdata

const DMA_EVENT_DEFS = `
typedef enum {
  DMA_BUSY_START = 0,
  DMA_BUSY_END = 1,
  DMA_VC_EXEC_START = 2,
  DMA_VC_EXEC_END = 3,
} DMA_DPF_EVENT_T;`

var (
	dmaEventNameKeeper EventNameKeeper = NewEventNameKeeper(DMA_EVENT_DEFS)
)

func ToDmaEventString(evtID int) (string, bool) {
	if name, ok := dmaEventNameKeeper.nameDict[evtID&3]; ok {
		return name, true
	}
	return "Unknown.1x.dma", false
}
