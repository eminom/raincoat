package rtinfo

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/codec"
)

type DevCycleTime struct {
	dpfSyncIndex int // dpfSyncIndex is at most 31 bit-wide
	devCycle     uint64
}

type HostTimeEntry struct {
	cid          int
	hosttime     uint64
	dpfSyncIndex int
}

type DevCycleAligned struct {
	DevCycleTime
	cid      int
	hosttime uint64
}

func (d DevCycleAligned) ToString() string {
	return fmt.Sprintf("%v %v %v", d.dpfSyncIndex, d.hosttime, d.devCycle)
}

type TimelineManager struct {
	cycles     []DevCycleTime
	hosttp     []HostTimeEntry
	alignedVec []DevCycleAligned
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
		if alignedVec[md].devCycle >= targetCycle {
			hi = md
		} else {
			lo = 1 + md
		}
	}
	bound := lo
	if bound < lz && alignedVec[bound].devCycle > targetCycle {
		bound--
	}
	if bound < lz-1 && bound >= 0 {
		hostStart, hostClose := alignedVec[bound].hosttime,
			alignedVec[bound+1].hosttime
		assert.Assert(hostStart < hostClose, "must be valid host time span")
		hostSpan := hostClose - hostStart

		cyStart, cyClose := alignedVec[bound].devCycle,
			alignedVec[bound+1].devCycle
		assert.Assert(cyStart < cyClose, "cycle must valid")
		cySpan := cyClose - cyStart
		hostAligned := hostStart +
			ratioMapTo(targetCycle-cyStart, cySpan, hostSpan)
		return hostAligned, true
	}
	return 0, false
}

func (tm *TimelineManager) PutEvent(evt codec.DpfEvent) {
	devCy := DevCycleTime{
		dpfSyncIndex: evt.DpfSyncIndex(),
		devCycle:     evt.Cycle,
	}
	tm.cycles = append(tm.cycles, devCy)
}

func (tm *TimelineManager) AlignToHostTimeline() {
	tm.trimHostWrapped()
	tm.trimWrappedSyncIndex()
	// if the host-timeline is in mono ascending

	timeMap := make(map[int]HostTimeEntry)
	for _, v := range tm.hosttp {
		timeMap[v.dpfSyncIndex] = v
	}

	var alignedVec []DevCycleAligned

	// tm.cycles are already in the right order
	for _, v := range tm.cycles {
		if host, ok := timeMap[v.dpfSyncIndex]; ok {
			_ = host
			alignedVec = append(
				alignedVec,
				DevCycleAligned{
					DevCycleTime: DevCycleTime{
						dpfSyncIndex: v.dpfSyncIndex,
						devCycle:     v.devCycle,
					},
					cid:      host.cid,
					hosttime: host.hosttime,
				},
			)
		}
	}

	log.Printf("time sync %v poinst are established", len(alignedVec))
	tm.alignedVec = alignedVec
}

func (tm *TimelineManager) trimHostWrapped() {
	lz := len(tm.hosttp)
	i := 0
	for i < lz {
		j := i
		for j < lz-1 && tm.hosttp[j].dpfSyncIndex < tm.hosttp[j+1].dpfSyncIndex {
			j++
		}
		if j+1 < lz {
			trimmed := (j + 1) - i
			log.Printf("warning: trimming host %v items, for sync index %v",
				trimmed,
				tm.hosttp[j].dpfSyncIndex,
			)
			tm.hosttp = tm.hosttp[j+1:]
			lz, i = len(tm.hosttp), 0
			continue
		}
		break
	}
	log.Printf("after trimmig, %v host timepoints remain", len(tm.hosttp))
}

func (tm *TimelineManager) Verify() bool {
	lz := len(tm.alignedVec)
	indexErrCount := 0
	hostErrCount := 0
	cycleErrCount := 0
	for i := 0; i < lz-1; i++ {
		if tm.alignedVec[i].dpfSyncIndex >= tm.alignedVec[i+1].dpfSyncIndex {
			indexErrCount++
		}
		if tm.alignedVec[i].hosttime >= tm.alignedVec[i+1].hosttime {
			hostErrCount++
		}
		if tm.alignedVec[i].devCycle >= tm.alignedVec[i+1].devCycle {
			cycleErrCount++
		}
	}
	return cycleErrCount == 0 && hostErrCount == 0 && indexErrCount == 0
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
	i := 0
	for i < lz {
		j := i
		for j < lz-1 && tm.cycles[j].dpfSyncIndex < tm.cycles[j+1].dpfSyncIndex {
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
	log.Printf("after trimming, %v dev cycles remains", len(tm.cycles))
}

func xsplit(a string) []string {
	rv := []string{}
	lz := len(a)
	i := 0
	for i < lz {
		for i < lz && unicode.IsSpace(rune(a[i])) {
			i++
		}
		j := i
		for j < lz && !unicode.IsSpace(rune(a[j])) {
			j++
		}
		if j-i > 0 {
			rv = append(rv, a[i:j])
		}
		i = j
	}
	return rv
}

// LoadTimepoints for
func (tm *TimelineManager) LoadTimepoints(path string) bool {
	fin, err := os.Open(path)
	if err != nil {
		log.Printf("error load timepoints: %v\n", err)
		return false
	}
	defer fin.Close()
	scan := bufio.NewScanner(fin)
	for {
		if !scan.Scan() {
			break
		}
		text := scan.Text()
		vs := xsplit(text)
		if len(vs) != 3 {
			panic(fmt.Errorf("error timepoints file content: %v"+
				", split len is %v",
				text,
				len(vs),
			))
		}
		var decodeErr error
		cidUint, decodeErr := strconv.ParseInt(vs[0], 10, 32)
		if decodeErr != nil {
			panic(decodeErr)
		}
		cid := int(cidUint)
		hosttime, decodeErr := strconv.ParseUint(vs[1], 10, 64)
		if decodeErr != nil {
			panic(decodeErr)
		}
		indexUint, decodeErr := strconv.ParseInt(vs[2], 10, 32)
		if decodeErr != nil {
			panic(decodeErr)
		}
		dpfSyncIndex := int(indexUint)

		tm.hosttp = append(tm.hosttp, HostTimeEntry{
			cid:          cid,
			hosttime:     hosttime,
			dpfSyncIndex: dpfSyncIndex,
		})
	}
	log.Printf("in all: %v timepoint(s) have been loaded", len(tm.hosttp))
	return true
}
