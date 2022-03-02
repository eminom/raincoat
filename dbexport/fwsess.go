package dbexport

import (
	"database/sql"
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

type FwSession struct {
	TableSession
}

func NewFwSession(db *sql.DB) *FwSession {
	return &FwSession{
		TableSession: NewTableSession(db, `insert into fw(
			idx, node_id, device_id, cluster_id, context_id, name,
			start_timestamp, end_timestamp, duration_timestamp,
			start_cycle, end_cycle, duration_cycle,
			packet_id, engine_type,
			vp_id, row_name,
			engine_id,
			tid
		) values(?, ?, ?, ?, ?, ?,
		         ?, ?, ?,
				 ?, ?, ?,
				 ?, ?,
				 ?, ?,
				 ?,
				 ?)`),
	}
}

func (fw *FwSession) AddFwTrace(idx, nodeID, devID, clusterID, ctxID int, name string,
	startTS, endTS, durTS uint64,
	startCy, endCy, durCy uint64,
	packetId int, engineType string,
	engineID int, rowName string,
) {
	//0:0:-1:2:ENGINE_TS:0:CQM Executable Launch0
	// row_name as name
	_, err := fw.stmt.Exec(idx, nodeID, devID, clusterID, ctxID, name,
		startTS, endTS, durTS,
		startCy, endCy, durCy,
		packetId, engineType,
		GetNextVpId(), rowName,
		engineID,
		fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v",
			nodeID, devID, ctxID, clusterID, engineType, engineID, rowName),
	)
	assert.Assert(err == nil, "Must be nil error: %v", err)
}
