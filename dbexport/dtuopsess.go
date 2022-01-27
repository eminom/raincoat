package dbexport

import (
	"database/sql"
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

const (
	DtuOpRowName = "DTU Op Re"
)

type DtuOpSession struct {
	TableSession
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

	_, err := dos.stmt.Exec(
		idx, nodeID, devID, clusterID, ctxID, name,
		startTS, endTS, durTS,
		startCy, endCy, durCy,
		opId, opName,
		// and the default
		vpId, moduleID,
		DtuOpRowName, fmt.Sprintf("%v:%v:%v:%v:%v",
			nodeID, devID, ctxID, clusterID, DtuOpRowName,
		))
	assert.Assert(err == nil, "Must be nil to carry on:%v", err)
}
