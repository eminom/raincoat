package meta

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrValidPacketIdNoOp = errors.New("known packet but no op")
	ErrInvalidPacketId   = errors.New("invalid packet id")
)

type DtuOp struct {
	OpName string
	OpId   int
}

type ExecScope struct {
	execUuid  uint64
	pktIdToOp map[int]int
	opMap     map[int]DtuOp
}

func (es ExecScope) DumpToOstream(fout *os.File) {
	fmt.Fprintf(fout, "# Packet to op id map\n")
	for pktId, opId := range es.pktIdToOp {
		fmt.Fprintf(fout, "%v %v\n", pktId, opId)
	}
	//TODO: Dump OP MAP
}

func (es ExecScope) FindOp(packetId int) (DtuOp, error) {
	if opId, ok := es.pktIdToOp[packetId]; ok {
		if rv, ok1 := es.opMap[opId]; ok1 {
			return rv, nil
		}
		return DtuOp{}, ErrValidPacketIdNoOp
	}
	return DtuOp{}, ErrInvalidPacketId
}

func (es ExecScope) CheckOpMapStatus(opMap map[int]bool) {
	matchedCount := 0
	for opId := range opMap {
		if _, ok := es.opMap[opId]; ok {
			matchedCount++
		}
	}
	fmt.Printf(
		"  all op(from pkt-mapped) %v, matched: %v out of %v\n",
		len(opMap),
		matchedCount,
		len(es.opMap),
	)
}

func (es ExecScope) MapPacketIdToOpId(packetId int) (int, bool) {
	if opId, ok := es.pktIdToOp[packetId]; ok {
		return opId, true
	}
	return 0, false
}

func (es ExecScope) IteratePacketId(cb func(pktId int)) {
	for pktId := range es.pktIdToOp {
		cb(pktId)
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

func testOpMetaFileName(execUuid uint64, prefix string, suffixes []SuffixConf) (
	string, func() FormatFetcher, bool) {
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

func loadOpMap(execUuid uint64, prefix string) map[int]DtuOp {
	inputName, creator, fileOK := testOpMetaFileName(
		execUuid,
		prefix,
		opFileSuffixes,
	)
	if !fileOK {
		return nil
	}
	fin, err := os.Open(inputName)
	if err != nil {
		log.Printf("error load dtu-op file for %016x: %v\n",
			execUuid,
			err)
		return nil
	}
	defer fin.Close()
	var fetcher FormatFetcher = creator()
	opMap := make(map[int]DtuOp)
	scan := bufio.NewScanner(fin)
	for {
		if !scan.Scan() {
			break
		}
		opIdStr, opName := fetcher.FetchOpIdOpName(scan.Text())
		opId, _ := strconv.Atoi(opIdStr)
		if _, ok := opMap[opId]; ok {
			panic(fmt.Errorf("duplicate op id for exec 0x%016x with op-id=%v",
				execUuid, opId,
			))
		}
		opMap[opId] = DtuOp{
			OpId:   opId,
			OpName: opName,
		}
	}
	return opMap
}

func loadExecScope(execUuid uint64, prefix string) *ExecScope {
	opMap := loadOpMap(execUuid, prefix)
	pktMap := loadPktToOpMap(execUuid, prefix)
	if opMap != nil && pktMap != nil {
		return &ExecScope{
			execUuid:  execUuid,
			opMap:     opMap,
			pktIdToOp: pktMap,
		}
	}
	return nil
}

type ExecRaw struct {
	startPath string
	bundle    map[uint64]*ExecScope
	wilds     map[uint64]*ExecScope
}

func (e *ExecRaw) LoadMeta(execUuid uint64) bool {
	if _, ok := e.bundle[execUuid]; ok {
		return true
	}
	exec := loadExecScope(
		execUuid,
		e.startPath,
	)
	if exec != nil {
		e.bundle[execUuid] = exec
		log.Printf("meta for 0x%016x is loaded", execUuid)
		return true
	}
	return false
}

func (e *ExecRaw) LoadWildcard() {
	entries, err := os.ReadDir(e.startPath)
	if err != nil {
		log.Printf("error readdir: %v", err)
		return
	}
	pat := regexp.MustCompile(`0x[a-f\d]+_dtuop\.dumptxt`)

	within := func(name string) bool {
		found := false
		for execUuid := range e.bundle {
			prefix := fmt.Sprintf("0x%016x", execUuid)[:10]
			if strings.HasPrefix(name, prefix) {
				found = true
				break
			}
		}
		return found
	}

	printWithin := func(str string) {
		for execUuid := range e.bundle {
			prefix := fmt.Sprintf("0x%0x16", execUuid)[:10]
			fmt.Printf("  %v to %v\n", prefix, str)
		}
	}
	_ = printWithin

	for _, entry := range entries {
		if !pat.MatchString(entry.Name()) {
			continue
		}
		if !within(entry.Name()) {
			execUuid, err := strconv.ParseUint(entry.Name()[2:10], 16, 64)
			if err != nil {
				panic(err)
			}
			execUuid <<= 32
			es := loadExecScope(execUuid, e.startPath)
			if es != nil {
				e.wilds[execUuid] = es
				fmt.Printf("exec 0x%016x is loaded for wildcard\n",
					execUuid)
				// For debug
				// printWithin(fmt.Sprintf("0x%016x", execUuid))
			}
		} else {
			// log.Printf("%v is skipped", entry.Name())
		}
	}
}

func (e ExecRaw) FindExecScope(execUuid uint64) (ExecScope, bool) {
	if rv, ok := e.bundle[execUuid]; ok {
		return *rv, true
	}
	return ExecScope{}, false
}

func (e ExecRaw) LookForWild(packetId int, inWild bool) (uint64, bool) {
	// assert.Assert(len(e.wilds) > 0, "Must be greater than zero")
	var mapToSearch = e.wilds
	if !inWild {
		mapToSearch = e.bundle
	}
	for execUuid, es := range mapToSearch {
		if _, ok := es.MapPacketIdToOpId(packetId); ok {
			return execUuid, true
		}
	}
	return 0, false
}

func (e ExecRaw) DumpInfo() {
	fmt.Printf("%v exec loaded in all\n", len(e.bundle))
	fmt.Printf("%v wildcard exec loaded in all\n", len(e.wilds))
}

func NewExecRaw(startPath string) *ExecRaw {
	return &ExecRaw{
		startPath: startPath,
		bundle:    make(map[uint64]*ExecScope),
		wilds:     make(map[uint64]*ExecScope),
	}
}