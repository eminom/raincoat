package metadata

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrValidPacketIdNoOp = errors.New("known packet but no op")
	ErrInvalidPacketId   = errors.New("invalid packet id")
)

type ExecScope struct {
	execUuid  uint64
	pktIdToOp map[int]int
	opMap     map[int]DtuOp
}

func NewExecScope(execUuid uint64,
	pktIdToOp map[int]int, opMap map[int]DtuOp) *ExecScope {
	return &ExecScope{
		execUuid,
		pktIdToOp,
		opMap,
	}
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

func (es ExecScope) GetExecUuid() uint64 {
	return es.execUuid
}

func (es ExecScope) IteratePktToOp(cb func(int, int)) {
	for pkt, opId := range es.pktIdToOp {
		cb(pkt, opId)
	}
}
