package dbexport

import (
	"database/sql"
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

type DmaSession struct {
	TableSession
}

func NewDmaSession(db *sql.DB) *DmaSession {
	return &DmaSession{
		TableSession: NewTableSession(db, `insert into memcpy(
			idx, node_id, device_id, cluster_id, context_id, name,
			start_timestamp, end_timestamp, duration_timestamp,
			start_cycle, end_cycle, duration_cycle,
			packet_id, engine_type,
			vp_id, row_name,
			tiling_mode,
			engine_id,
			vc,
			tid
		) values(?, ?, ?, ?, ?, ?,
		         ?, ?, ?,
				 ?, ?, ?,
				 ?, ?,
				 ?, ?,
				 ?,
				 ?,
				 ?,
				 ?)`),
	}
}

func (dmaS *DmaSession) AddDmaTrace(idx, nodeID, devID, clusterID, ctxID int,
	name string,
	startTS, endTS, durTS uint64,
	startCy, endCy, durCy uint64,
	packetId int, engineType string,
	tilingMode string,
	engineID int,
	vc int) {
	//0:0:-1:2:ENGINE_TS:0:CQM Executable Launch0
	// row_name as name
	rowName := name
	_, err := dmaS.stmt.Exec(idx, nodeID, devID, clusterID, ctxID, name,
		startTS, endTS, durTS,
		startCy, endCy, durCy,
		packetId, engineType,
		GetNextVpId(), rowName,
		tilingMode,
		engineID,
		vc,
		fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v",
			nodeID, devID, ctxID, clusterID, engineType, engineID, name),
	)
	assert.Assert(err == nil, "Must be nil error: %v", err)
}
