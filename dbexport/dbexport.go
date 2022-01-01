package dbexport

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"git.enflame.cn/hai.bai/dmaster/rtinfo"
	_ "github.com/mattn/go-sqlite3"
)

type ExtractOpInfo func(rtinfo.OpActivity) (bool, string, string)

type AddOpTrace func(
	idx, nodeID, devID, clusterID, ctxID int, name string,
	startTS, endTS, durTS uint64,
	startCy, endCy, durCy uint64,
	opId int, opName string)

type DbSession struct {
	targetName string
	finish     func()
	addOpTrace AddOpTrace
	idx        int
}

func ifFileExist(file string) bool {
	stat, err := os.Stat(file)
	return nil == err && !stat.IsDir()
}

func NewDbSession(target string) (*DbSession, error) {
	if ifFileExist(target) {
		os.Remove(target)
	}

	db, err := sql.Open("sqlite3", target)
	if err != nil {
		return nil, err
	}

	sqlStmt := createDtuOpTable + `
	delete from dtu_op;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare(`insert into dtu_op(
		idx, node_id, device_id, cluster_id, context_id, name,
		start_timestamp, end_timestamp, duration_timestamp,
		start_cycle, end_cycle, duration_cycle,
		op_id, op_name,
		vp_id, module_id,
		row_name, tid)
		values(?, ?, ?, ?, ?, ?,
			   ?, ?, ?,
			   ?, ?, ?,
			   ?, ?,
			   ?, ?,
			   ?, ?)`)
	if err != nil {
		log.Fatal(err)
	}
	finish := func() {
		tx.Commit()
		db.Close()
	}
	addOpTrace := func(
		idx, nodeID, devID, clusterID, ctxID int, name string,
		startTS, endTS, durTS uint64,
		startCy, endCy, durCy uint64,
		opId int, opName string) {

		moduleID := 1
		vpId := idx

		stmt.Exec(
			idx, nodeID, devID, clusterID, ctxID, name,
			startTS, endTS, durTS,
			startCy, endCy, durCy,
			opId, opName,
			// and the default
			vpId, moduleID,
			"DTU Op", fmt.Sprintf("%v:%v:%v:%v:DTU Op",
				nodeID, devID, ctxID, clusterID,
			))
	}
	return &DbSession{
		finish:     finish,
		addOpTrace: addOpTrace,
	}, nil
}

func (dbs *DbSession) Close() {
	dbs.finish()
}

func (dbs *DbSession) DumpToEventTrace(
	bundle []rtinfo.OpActivity,
	tm *rtinfo.TimelineManager,
	extractor ExtractOpInfo,
	dumpWild bool,
) {
	dtuOpCount := 0
	convertToHostError := 0
	const nodeID = 0
	const deviceID = 0
	const clusterID = -1
	for _, act := range bundle {
		///act.IsOpRefValid()
		if okToShow, _, name := extractor(act); okToShow {
			dtuOpCount++
			startHostTime, startOK := tm.MapToHosttime(act.StartCycle())
			endHostTime, endOK := tm.MapToHosttime(act.EndCycle())
			if startOK && endOK {
				dbs.addOpTrace(dbs.idx, nodeID, deviceID, clusterID, act.Start.Context, name,
					startHostTime, endHostTime, endHostTime-startHostTime,
					act.StartCycle(), act.EndCycle(), act.EndCycle()-act.StartCycle(),
					act.GetOp().OpId, act.GetOp().OpName,
				)
				dbs.idx++
			} else {
				convertToHostError++
			}
		}
	}
	log.Printf("%v dtu-op record(s) have been traced into %v",
		dtuOpCount,
		dbs.targetName,
	)
}
