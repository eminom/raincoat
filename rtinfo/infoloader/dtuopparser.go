package infoloader

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"

	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
)

type DtuOpMapLoader interface {
	LoadOpMap(filename string) map[int]metadata.DtuOp
}

type compatibleFetcher struct{}

func (c compatibleFetcher) LoadOpMap(filename string) map[int]metadata.DtuOp {
	fin, err := os.Open(filename)
	if err != nil {
		log.Printf("error load dtu-op file for %016x: %v\n",
			filename,
			err)
		return nil
	}
	defer fin.Close()
	opMap := make(map[int]metadata.DtuOp)
	scan := bufio.NewScanner(fin)
	for {
		if !scan.Scan() {
			break
		}
		opIdStr, opName := c.fetchOpIdOpName(scan.Text())
		opId, _ := strconv.Atoi(opIdStr)
		if _, ok := opMap[opId]; ok {
			panic(fmt.Errorf("duplicate op id for exec 0x%016x with op-id=%v",
				filename, opId,
			))
		}
		opMap[opId] = metadata.DtuOp{
			OpId:   opId,
			OpName: opName,
		}
	}
	return opMap
}

func (compatibleFetcher) fetchOpIdOpName(text string) (string, string) {
	vs := XSplit(text, 3)
	return vs[0], vs[1]
}

type nuevoModeFetcher struct{}

func (c nuevoModeFetcher) LoadOpMap(filename string) map[int]metadata.DtuOp {
	fin, err := os.Open(filename)
	if err != nil {
		log.Printf("error load dtu-op file for %016x: %v\n",
			filename,
			err)
		return nil
	}
	defer fin.Close()
	opMap := make(map[int]metadata.DtuOp)
	scan := bufio.NewScanner(fin)
	for {
		if !scan.Scan() {
			break
		}
		opIdStr, opName := c.fetchOpIdOpName(scan.Text())
		opId, _ := strconv.Atoi(opIdStr)
		if _, ok := opMap[opId]; ok {
			panic(fmt.Errorf("duplicate op id for exec 0x%016x with op-id=%v",
				filename, opId,
			))
		}
		opMap[opId] = metadata.DtuOp{
			OpId:   opId,
			OpName: opName,
		}
	}
	return opMap
}

func (nuevoModeFetcher) fetchOpIdOpName(text string) (string, string) {
	vs := XSplit(text, 4)
	return vs[0], vs[2]
}

func newCompatibleOpLoader() DtuOpMapLoader {
	return compatibleFetcher{}
}

func newNuevoOpLoader() DtuOpMapLoader {
	return nuevoModeFetcher{}
}
