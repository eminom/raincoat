package dbexport

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"git.enflame.cn/hai.bai/dmaster/rtinfo"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	_ "github.com/mattn/go-sqlite3"
)

type ExtractOpInfo func(rtdata.OpActivity) (bool, string, string)

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

type TableSession struct {
	stmt      *sql.Stmt
	tx        *sql.Tx
	cmdString string
}

func NewTableSession(db *sql.DB, cmdString string) TableSession {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare(cmdString)
	if err != nil {
		log.Fatal(err)
	}
	return TableSession{
		stmt:      stmt,
		tx:        tx,
		cmdString: cmdString,
	}
}

func (tabSess *TableSession) Close() {
	tabSess.tx.Commit()
	tabSess.stmt.Close()
}

func (tabSess *TableSession) Exec(args ...interface{}) {
	tabSess.stmt.Exec(args...)
}

type HeaderSession struct {
	TableSession
}

func NewHeaderSess(db *sql.DB) *HeaderSession {
	return &HeaderSession{
		TableSession: NewTableSession(db,
			`insert into header(table_name, version, category,
			count, time_unit) values(?, ?, ?, ?, ?)`),
	}
}

func (hs *HeaderSession) AddHeader(tableName string, version string, category string,
	count int, timeUnit string) {
	hs.stmt.Exec(tableName,
		version, category, fmt.Sprintf("%v", count),
		timeUnit)
}

type DtuOpSession struct {
	TableSession
	vid        int32
	dtuOpCount int
}

func NewDtuOpSession(db *sql.DB) *DtuOpSession {
	return &DtuOpSession{
		TableSession: NewTableSession(db, `insert into dtu_op(
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
				   ?, ?)`),
	}
}

func (dos *DtuOpSession) GetNextVid() int {
	return int(atomic.AddInt32(&dos.vid, 1))
}

func (dos *DtuOpSession) AddDtuOp(
	idx, nodeID, devID, clusterID, ctxID int, name string,
	startTS, endTS, durTS uint64,
	startCy, endCy, durCy uint64,
	opId int, opName string) {

	moduleID := 1
	vpId := dos.GetNextVid()

	dos.stmt.Exec(
		idx, nodeID, devID, clusterID, ctxID, name,
		startTS, endTS, durTS,
		startCy, endCy, durCy,
		opId, opName,
		// and the default
		vpId, moduleID,
		"DTU Op", fmt.Sprintf("%v:%v:%v:%v:DTU Op",
			nodeID, devID, ctxID, clusterID,
		))
	dos.dtuOpCount++

}

func NewDbSession(target string) (*DbSession, error) {
	if ifFileExist(target) {
		os.Remove(target)
	}

	db, err := sql.Open("sqlite3", target)
	if err != nil {
		return nil, err
	}

	sqlStmt := createHeaderTable + "\n" +
		createDtuOpTable + `
	delete from dtu_op;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil, err
	}

	// Prepare a header for you

	hs := NewHeaderSess(db)
	dos := NewDtuOpSession(db)

	finishDbWork := func() {
		// Finalize Dtu-ops
		dos.Close()

		// Finalize headers
		hs.AddHeader("dtu_op", "1.0",
			TableCategory_DTUOpActivity, dos.dtuOpCount, "ns")
		hs.Close()
		db.Close()
	}

	return &DbSession{
		finish:     finishDbWork,
		addOpTrace: dos.AddDtuOp,
	}, nil
}

func (dbs *DbSession) Close() {
	dbs.finish()
}

func (dbs *DbSession) DumpToEventTrace(
	bundle []rtdata.OpActivity,
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
