package infoloader

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/efintf/affinity"
	"git.enflame.cn/hai.bai/dmaster/meta/dtuarch"
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/topsdev/mimic/mimicdefs"
)

type taskFileLoader struct {
	startupPath string
}

func (d taskFileLoader) GetRuntimeTaskPath() string {
	return filepath.Join(d.startupPath, "runtime_task.txt")
}

func (d taskFileLoader) LoadTask(oneSolid bool) (dc map[int]*rtdata.RuntimeTask,
	taskSequentials []int,
	ok bool,
) {
	return loadTaskFromFile(d.GetRuntimeTaskPath())
}

type metaFileLoader struct {
	efintf.TaskLoader
	startupPath string
}

func NewMetaFileLoader(startup string, oneTaskMode bool) *metaFileLoader {
	return &metaFileLoader{
		TaskLoader: (func() efintf.TaskLoader {
			if oneTaskMode {
				return OneSolidTaskLoader{}
			}
			return taskFileLoader{startupPath: startup}
		})(),
		startupPath: startup,
	}
}

func (metaFileLoader) GetArchType() dtuarch.ArchType {
	return dtuarch.EnflameUnknownArch
}

func (metaFileLoader) GetCdmaAffinity() affinity.CdmaAffinitySet {
	return affinity.NewDoradoCdmaAffinityDefault()
}

func (d metaFileLoader) GetMetaStartupPath() string {
	return d.startupPath
}

func (d metaFileLoader) GetTimepointsPath() string {
	return filepath.Join(d.startupPath, "timepoints.txt")
}

func loadTaskFromFile(filename string) (
	dc map[int]*rtdata.RuntimeTask,
	taskSequentials []int,
	ok bool,
) {
	fin, err := os.Open(filename)
	if err != nil {
		log.Printf("error load runtime info from \"%v\":%v\n", filename, err)
		return
	}
	defer fin.Close()

	// dc: Full runtime task info
	// including the ones with cycle info and the ones without cycle info
	dc = make(map[int]*rtdata.RuntimeTask)
	scan := bufio.NewScanner(fin)
	for {
		if !scan.Scan() {
			break
		}
		line := scan.Text()
		vs := strings.Split(line, " ")
		taskId, err := strconv.Atoi(vs[0])
		if err != nil {
			log.Printf("error read '%v'", line)
			continue
		}

		if _, ok := dc[taskId]; ok {
			panic("error runtimetask: duplicate task id")
		}

		hxVal := vs[1]
		if strings.HasPrefix(hxVal, "0x") || strings.HasPrefix(hxVal, "0X") {
			hxVal = hxVal[2:]
		}
		exec, err := strconv.ParseUint(hxVal, 16, 64)
		if err != nil {
			log.Printf("error exec: %v", vs[1])
		}
		pgMask, err := strconv.Atoi(vs[2])
		if err != nil {
			log.Printf("error read pg mask: %v", err)
		}
		dc[taskId] = &rtdata.RuntimeTask{
			RuntimeTaskBase: rtdata.RuntimeTaskBase{TaskID: taskId,
				ExecutableUUID: exec,
				PgMask:         pgMask,
			},
		}
		taskSequentials = append(taskSequentials, taskId)
	}
	sort.Ints(taskSequentials)
	ok = true
	return
}

func (d metaFileLoader) LoadTimepoints() (hosttp []rtdata.HostTimeEntry, ok bool) {
	filename := d.GetTimepointsPath()
	fin, err := os.Open(filename)
	if err != nil {
		log.Printf("error load timepoints: %v\n", err)
		return
	}
	defer fin.Close()

	scan := bufio.NewScanner(fin)
	for {
		if !scan.Scan() {
			break
		}
		text := scan.Text()
		vs := strings.Fields(text)
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

		hosttp = append(hosttp, rtdata.HostTimeEntry{
			Cid:          cid,
			Hosttime:     hosttime,
			DpfSyncIndex: dpfSyncIndex,
		})
	}
	log.Printf("in all: %v timepoint(s) have been loaded", len(hosttp))
	ok = true
	return
}

func (metaFileLoader) ExtractHostInfo() *mimicdefs.HostInfo {
	return nil
}

func (d metaFileLoader) LoadWildcards(checkExist func(str string) bool,
	notfiyNew func(uint64, *metadata.ExecScope)) {
	entries, err := os.ReadDir(d.startupPath)
	if err != nil {
		log.Printf("error readdir: %v", err)
		return
	}
	// only load for dtuop now(not including dtuop dumped in PbMODE)
	//(TODO:baihai)
	pat := regexp.MustCompile(`0x[a-f\d]+_dtuop\.dumptxt`)
	for _, entry := range entries {
		if !pat.MatchString(entry.Name()) {
			continue
		}
		if !checkExist(entry.Name()) {
			execUuid, err := strconv.ParseUint(entry.Name()[2:10], 16, 64)
			if err != nil {
				panic(err)
			}
			execUuid <<= 32
			es := d.LoadExecScope(execUuid)
			if es != nil {
				notfiyNew(execUuid, es)
			}
		} else {
			// log.Printf("%v is skipped", entry.Name())
		}
	}
}

func isFileExist(name string) bool {
	info, err := os.Stat(name)
	// log.Printf("test for (%v)", name)
	return nil == err && !info.IsDir()
}

func testMetaFileName(execUuid uint64, prefix string, suffixes []string) (string, bool) {
	mark := fmt.Sprintf("0x%016x", execUuid)[:10]
	for _, suf := range suffixes {
		if inputName := filepath.Join(prefix, mark+suf); isFileExist(inputName) {
			// log.Printf("bingo for %v", inputName)
			return inputName, true
		}
	}
	return "", false
}

func testOpMetaFileName(execUuid uint64, prefix string, suffixes []OpInfoSuffixConf) (
	string, func() DtuOpMapLoader, bool) {
	mark := fmt.Sprintf("0x%016x", execUuid)[:10]
	for _, suf := range suffixes {
		if inputName := filepath.Join(prefix,
			mark+suf.suffixName); isFileExist(inputName) {
			// log.Printf("bingo for %v", inputName)
			return inputName, suf.fetcherCreator, true
		}
	}
	return "", nil, false
}

func testDmaMetaFileName(execUuid uint64, prefix string, suffixes []DmaInfoSuffixConf) (
	string, func() DmaOpFormatFetcher, bool) {
	mark := fmt.Sprintf("0x%016x", execUuid)[:10]
	for _, suf := range suffixes {
		if inputName := filepath.Join(prefix,
			mark+suf.suffixName); isFileExist(inputName) {
			return inputName, suf.fetcherCreator, true
		}
	}
	return "dma-meta-file-not-exist", nil, false
}

func loadPktToOpMap(execUuid uint64, prefix string) map[int]int {
	inputName, fileOK := testMetaFileName(
		execUuid,
		prefix,
		pkt2opFileSuffixes,
	)
	if !fileOK {
		return nil
	}
	fin, err := os.Open(inputName)
	if err != nil {
		log.Printf("error load pkt-to-op map file for %016x: %v",
			execUuid,
			err)
		return nil
	}
	defer fin.Close()
	pktIdToOp := make(map[int]int)
	scan := bufio.NewScanner(fin)
	for {
		if !scan.Scan() {
			break
		}
		vs := strings.Split(scan.Text(), " ")
		pktId, err := strconv.Atoi(vs[0])
		if err != nil {
			panic(err)
		}

		opid, err := strconv.Atoi(vs[1])
		if err != nil {
			panic(err)
		}

		if _, ok := pktIdToOp[pktId]; ok {
			panic(fmt.Errorf("duplicate packet id for %016x pkt-id %v",
				execUuid, pktId))
		}
		pktIdToOp[pktId] = opid

	}
	return pktIdToOp
}

func loadOpMap(execUuid uint64, prefix string) map[int]metadata.DtuOp {
	inputName, dtuOpInfLoaderCreator, fileOK := testOpMetaFileName(
		execUuid,
		prefix,
		GetOpFileSuffixes(),
	)
	if !fileOK {
		return nil
	}
	return dtuOpInfLoaderCreator().LoadOpMap(inputName)
}

func loadDmaMap(execUuid uint64, prefix string) metadata.DmaInfoMap {
	inputName, dmaInfoLoaderCreator, fileOK := testDmaMetaFileName(
		execUuid,
		prefix,
		GetDmaInfoFileSuffix(),
	)
	if !fileOK {
		return metadata.DmaInfoMap{}
	}
	return dmaInfoLoaderCreator().FetchDmaOpDict(inputName)
}

func (d metaFileLoader) LoadExecScope(execUuid uint64) *metadata.ExecScope {
	opMap := loadOpMap(execUuid, d.startupPath)
	pktMap := loadPktToOpMap(execUuid, d.startupPath)
	dmaMap := loadDmaMap(execUuid, d.startupPath)
	if opMap != nil && pktMap != nil {
		return metadata.NewExecScope(
			execUuid,
			pktMap,
			opMap,
			dmaMap,
			make(map[int][]metadata.SubOpMeta),
			make(map[string]metadata.ModThunk), // leave it empty for file loader
		)
	}
	return nil
}

type bufferLoader struct {
	rawfilenames []string
}

func NewContentBufferLoader(rawfilenames ...string) *bufferLoader {
	var files []string
	files = append(files, rawfilenames...)
	return &bufferLoader{rawfilenames: files}
}

func (bl bufferLoader) GetRingBufferCount() int {
	return len(bl.rawfilenames)
}

func (bl bufferLoader) LoadRingBufferContent(cid int, rbIdx int) []byte {
	name := bl.rawfilenames[rbIdx]
	chunk, err := os.ReadFile(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %v\n", name)
		os.Exit(1)
	}
	return chunk
}

func (bl bufferLoader) GetInputName() string {
	if len(bl.rawfilenames) > 0 {
		return bl.rawfilenames[0]
	}
	return "uname"
}
