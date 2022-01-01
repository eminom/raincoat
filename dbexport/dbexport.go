package dbexport

import (
	"database/sql"
	"log"
	"os"

	"git.enflame.cn/hai.bai/dmaster/rtinfo"
	_ "github.com/mattn/go-sqlite3"
)

type ExtractOpInfo func(rtinfo.OpActivity) (bool, string, string)

type AddOneTrace func(idx int, name string, ctxID int, startTS, endTS, durTS,
	startCy, endCy, durCy uint64, opId int, opName string, clusterID int)

type DbSession struct {
	targetName string
	finish     func()
	addTrace   AddOneTrace
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
		idx, name,
		context_id,
		start_timestamp, end_timestamp, duration_timestamp,
		start_cycle, end_cycle, duration_cycle,
		op_id, op_name,
		cluster_id) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	// layer_kind, kind, fusion_kind, description, device is none
	// meta is there but not for now
	// tid is 0:0:0:xxx
	// vp_id is 0 ~ max-1

	if err != nil {
		log.Fatal(err)
	}
	finish := func() {
		tx.Commit()
		db.Close()
	}
	addTrace := func(idx int, name string, ctxID int, startTS, endTS, durTS,
		startCy, endCy, durCy uint64, opId int, opName string, clusterID int) {
		stmt.Exec(idx, name, ctxID, startTS, endTS, durTS,
			startCy, endCy, durCy, opId, opName, clusterID)
	}
	return &DbSession{
		finish:   finish,
		addTrace: addTrace,
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
	for _, act := range bundle {
		///act.IsOpRefValid()
		if okToShow, _, name := extractor(act); okToShow {
			dtuOpCount++
			startHostTime, startOK := tm.MapToHosttime(act.StartCycle())
			endHostTime, endOK := tm.MapToHosttime(act.EndCycle())
			if startOK && endOK {
				dbs.addTrace(dbs.idx, name, act.Start.Context,
					startHostTime, endHostTime, endHostTime-startHostTime,
					act.StartCycle(), act.EndCycle(), act.EndCycle()-act.StartCycle(),
					act.GetOp().OpId, act.GetOp().OpName, act.DpfAct.Start.ClusterID,
				)
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
