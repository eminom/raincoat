package dbexport

import (
	"database/sql"
	"fmt"
	"log"
)

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

func (dos *DtuOpSession) AddDtuOp(
	idx, nodeID, devID, clusterID, ctxID int, name string,
	startTS, endTS, durTS uint64,
	startCy, endCy, durCy uint64,
	opId int, opName string) {

	moduleID := 1
	vpId := GetNextVpId()

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

type FwSession struct {
	TableSession
	fwOpCount int
}

func NewFwSession(db *sql.DB) *FwSession {
	return &FwSession{
		TableSession: NewTableSession(db, `insert into fw(
			idx, node_id, device_id, cluster_id, context_id, name,
			start_timestamp, end_timestamp, duration_timestamp,
			start_cycle, end_cycle, duration_cycle,
			packet_id, engine_type,
			vp_id, row_name
		) values(?, ?, ?, ?, ?, ?,
		         ?, ?, ?,
				 ?, ?, ?,
				 ?, ?,
				 ?, ?)`),
	}
}

func (fw *FwSession) AddFwTrace(idx, nodeID, devID, clusterID, ctxID int, name string,
	startTS, endTS, durTS uint64,
	startCy, endCy, durCy uint64,
	packetId int, engineType string) {
	fw.stmt.Exec(idx, nodeID, devID, clusterID, ctxID, name,
		startTS, endTS, durTS,
		startCy, endCy, durCy,
		packetId, engineType,
		GetNextVpId(), name,
	)
}
