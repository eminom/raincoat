package meta

import (
	"fmt"
	"log"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/efintf"
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
)

type PktToOpMap struct {
	pktToOp map[int]int
}

type ExecRaw struct {
	loader efintf.InfoReceiver
	bundle map[uint64]*metadata.ExecScope
	wilds  map[uint64]*metadata.ExecScope

	//~ Overall
	pkOpMap PktToOpMap
}

func (e *ExecRaw) LoadMeta(execUuid uint64) bool {
	if _, ok := e.bundle[execUuid]; ok {
		return true
	}

	exec := e.loader.LoadExecScope(execUuid)
	if exec != nil {
		e.bundle[execUuid] = exec
		log.Printf("meta for 0x%016x is loaded", execUuid)
		return true
	}
	return false
}

func (e *ExecRaw) LoadWildcard() {

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

	e.loader.LoadWildcards(within,
		func(execUuid uint64, es *metadata.ExecScope) {
			e.wilds[execUuid] = es
			fmt.Printf("exec 0x%016x is loaded for wildcard\n",
				execUuid)
			// For debug
			// printWithin(fmt.Sprintf("0x%016x", execUuid))
		})

	e.buildPktOpMap()
}

func (e ExecRaw) FindExecScope(execUuid uint64) (metadata.ExecScope, bool) {
	if rv, ok := e.bundle[execUuid]; ok {
		return *rv, true
	}
	return metadata.ExecScope{}, false
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

func (e ExecRaw) WalkExecScopes(walk func(exec *metadata.ExecScope) bool) {
	for _, es := range e.wilds {
		if !walk(es) {
			break
		}
	}
}

func (e *ExecRaw) buildPktOpMap() {
	pktToOp := make(map[int]int)
	e.WalkExecScopes(func(es *metadata.ExecScope) bool {
		es.IteratePktToOp(func(pktId, opId int) {
			pktToOp[pktId] = opId
		})
		return true
	})
	e.pkOpMap.pktToOp = pktToOp
}

func (e ExecRaw) GetOpIdForPacketId(packetId int) (int, bool) {
	rv, ok := e.pkOpMap.pktToOp[packetId]
	return rv, ok
}

func (er ExecRaw) GetSubOpMetaMap(execUuid uint64) map[int]map[int][]string {
	execScope, exists := er.bundle[execUuid]
	if !exists {
		return nil
	}
	// An executable scope "isa"
	// A task scope
	return execScope.GetSubOpMetaMap()
}

func NewExecRaw(loader efintf.InfoReceiver) *ExecRaw {
	return &ExecRaw{
		loader: loader,
		bundle: make(map[uint64]*metadata.ExecScope),
		wilds:  make(map[uint64]*metadata.ExecScope),
	}
}
