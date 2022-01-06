package rtinfo

import (
	"fmt"
	"log"
	"os"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type TimelineManager struct {
	cycles     []rtdata.DevCycleTime
	hosttp     []rtdata.HostTimeEntry
	alignedVec []rtdata.DevCycleAligned
}

func NewTimelineManager() *TimelineManager {
	return &TimelineManager{}
}

func ratioMapTo(a, b uint64, c uint64) uint64 {
	return uint64(float64(a) / float64(b) * float64(c))
}

// helper for find the legal span for cycle to belong to
func (tm *TimelineManager) MapToHosttime(targetCycle uint64) (uint64, bool) {
	alignedVec := tm.alignedVec
	lz := len(alignedVec)
	lo, hi := 0, lz
	for lo < hi {
		md := (lo + hi) >> 1
		if alignedVec[md].DevCycle >= targetCycle {
			hi = md
		} else {
			lo = 1 + md
		}
	}
	bound := lo
	if bound < lz && alignedVec[bound].DevCycle > targetCycle {
		bound--
	}
	if bound < lz-1 && bound >= 0 {
		hostStart, hostClose := alignedVec[bound].Hosttime,
			alignedVec[bound+1].Hosttime
		assert.Assert(hostStart < hostClose, "must be valid host time span")
		hostSpan := hostClose - hostStart

		cyStart, cyClose := alignedVec[bound].DevCycle,
			alignedVec[bound+1].DevCycle
		assert.Assert(cyStart < cyClose, "cycle must valid")
		cySpan := cyClose - cyStart
		hostAligned := hostStart +
			ratioMapTo(targetCycle-cyStart, cySpan, hostSpan)
		return hostAligned, true
	}
	return 0, false
}

func (tm *TimelineManager) DispatchEvent(evt codec.DpfEvent) error {
	devCy := rtdata.DevCycleTime{
		DpfSyncIndex: evt.DpfSyncIndex(),
		DevCycle:     evt.Cycle,
	}
	tm.cycles = append(tm.cycles, devCy)
	return nil
}

func (tm *TimelineManager) AlignToHostTimeline() {
	tm.trimHostWrapped()
	tm.trimWrappedSyncIndex()
	// if the host-timeline is in mono ascending

	timeMap := make(map[int]rtdata.HostTimeEntry)
	for _, v := range tm.hosttp {
		timeMap[v.DpfSyncIndex] = v
	}

	var alignedVec []rtdata.DevCycleAligned

	// tm.cycles are already in the right order
	for _, v := range tm.cycles {
		if host, ok := timeMap[v.DpfSyncIndex]; ok {
			_ = host
			alignedVec = append(
				alignedVec,
				rtdata.DevCycleAligned{
					DevCycleTime: rtdata.DevCycleTime{
						DpfSyncIndex: v.DpfSyncIndex,
						DevCycle:     v.DevCycle,
					},
					Cid:      host.Cid,
					Hosttime: host.Hosttime,
				},
			)
		}
	}

	log.Printf("time sync %v poinst are established", len(alignedVec))
	tm.alignedVec = alignedVec
}

func (tm *TimelineManager) trimHostWrapped() {
	lz := len(tm.hosttp)
	originalHostTpLen := lz
	i := 0
	for i < lz {
		j := i
		for j < lz-1 && tm.hosttp[j].DpfSyncIndex < tm.hosttp[j+1].DpfSyncIndex {
			j++
		}
		if j+1 < lz {
			trimmed := (j + 1) - i
			log.Printf("warning: trimming host %v items, for sync index %v",
				trimmed,
				tm.hosttp[j].DpfSyncIndex,
			)
			tm.hosttp = tm.hosttp[j+1:]
			lz, i = len(tm.hosttp), 0
			continue
		}
		break
	}
	log.Printf("after trimmig, %v host timepoints remain from %v",
		len(tm.hosttp),
		originalHostTpLen,
	)
}

func (tm *TimelineManager) Verify() bool {
	lz := len(tm.alignedVec)
	indexErrCount := 0
	hostErrCount := 0
	cycleErrCount := 0
	for i := 0; i < lz-1; i++ {
		if tm.alignedVec[i].DpfSyncIndex >= tm.alignedVec[i+1].DpfSyncIndex {
			indexErrCount++
		}
		if tm.alignedVec[i].Hosttime >= tm.alignedVec[i+1].Hosttime {
			hostErrCount++
		}
		if tm.alignedVec[i].DevCycle >= tm.alignedVec[i+1].DevCycle {
			cycleErrCount++
		}
	}
	return cycleErrCount == 0 && hostErrCount == 0 && indexErrCount == 0
}

func (tm *TimelineManager) GetEngineTypeCodes() []codec.EngineTypeCode {
	return []codec.EngineTypeCode{codec.EngCat_PCIE}
}

func (tm *TimelineManager) DumpInfo() {
	fout, err := os.Create("syncpoints.txt")
	if err != nil {
		log.Printf("error syncpoints.txt:%v", err)
		return
	}
	defer fout.Close()
	for _, v := range tm.alignedVec {
		fmt.Fprintf(fout, "%v\n", v.ToString())
	}
}

// Sometimes, the real world is so crude,
// There are duplicated DPF sync index in the same session
// So we have to to do something
func (tm *TimelineManager) trimWrappedSyncIndex() {
	lz := len(tm.cycles)
	originalLen := lz
	i := 0
	for i < lz {
		j := i
		for j < lz-1 && tm.cycles[j].DpfSyncIndex < tm.cycles[j+1].DpfSyncIndex {
			j++
		}
		if j+1 < lz {
			trimmed := (j + 1) - i
			log.Printf("%v dpf sync indice are trimmed", trimmed)
			tm.cycles = tm.cycles[:j+1]
			// reset
			lz, i = len(tm.cycles), 0
			continue
		}
		break
	}
	log.Printf("after trimming, %v dev cycles remains from %v",
		len(tm.cycles),
		originalLen,
	)
}

// LoadTimepoints for
func (tm *TimelineManager) LoadTimepoints(
	infoReceiver efintf.InfoReceiver,
) bool {
	tmVec, ok := infoReceiver.LoadTimepoints()
	if !ok {
		return false
	}
	tm.hosttp = tmVec
	return true
}
