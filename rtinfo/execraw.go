package rtinfo

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type DtuOp struct {
	OpName string
	OpId   int
}

type ExecScope struct {
	pktIdToOp map[int]int
	opMap     map[int]DtuOp
}

func loadPktToOpMap(execUuid uint64, prefix string) map[int]int {
	inputName := fmt.Sprintf("0x%016x", execUuid)[:10] + "_pkt2op.dumptxt"
	inputName = filepath.Join(prefix, inputName)
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
	inputName := fmt.Sprintf("0x%016x", execUuid)[:10] + "_dtuop.dumptxt"
	inputName = filepath.Join(prefix, inputName)
	fin, err := os.Open(inputName)
	if err != nil {
		log.Printf("error load dtu-op file for %016x: %v\n",
			execUuid,
			err)
		return nil
	}
	defer fin.Close()
	opMap := make(map[int]DtuOp)
	scan := bufio.NewScanner(fin)
	for {
		if !scan.Scan() {
			break
		}
		text := scan.Text()
		lz := len(text)
		i := 0
		for i < lz && text[i] >= '0' && text[i] <= '9' {
			i++
		}
		opId, _ := strconv.Atoi(text[:i])
		i += 1
		nameStart := i
		for i < lz && text[i] != ' ' {
			i++
		}
		opName := text[nameStart:i]
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

func LoadExecScope(execUuid uint64, prefix string) *ExecScope {
	opMap := loadOpMap(execUuid, prefix)
	pktMap := loadPktToOpMap(execUuid, prefix)
	if opMap != nil && pktMap != nil {
		return &ExecScope{
			opMap:     opMap,
			pktIdToOp: pktMap,
		}
	}
	return nil
}

type ExecRaw struct {
	startPath string
	bundle    map[uint64]*ExecScope
}

func (e *ExecRaw) LoadMeta(execUuid uint64) bool {
	if _, ok := e.bundle[execUuid]; ok {
		return true
	}
	exec := LoadExecScope(
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

func NewExecRaw(startPath string) *ExecRaw {
	return &ExecRaw{
		startPath: startPath,
		bundle:    make(map[uint64]*ExecScope),
	}
}
