package dbexport

import (
	"database/sql"
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

type KernelSession struct {
	TableSession
}

func NewKernelSession(db *sql.DB) *KernelSession {
	return &KernelSession{
		TableSession: NewTableSession(db,
			`insert into kernel(
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

func (kernS *KernelSession) AddKernelTrace(idx, nodeID, devID, clusterID, ctxID int,
	name string,
	startTS, endTS, durTS uint64,
	startCy, endCy, durCy uint64,
	packetId int, engineType string,
	engineID int) {
	//0:0:-1:2:ENGINE_SIP:0:SIP BUSY
	// And SIP BUSY only so far.
	// row_name as name
	rowName := name
	_, err := kernS.stmt.Exec(idx, nodeID, devID, clusterID, ctxID, name,
		startTS, endTS, durTS,
		startCy, endCy, durCy,
		packetId, engineType,
		GetNextVpId(), rowName,
		engineID,
		fmt.Sprintf("%v:%v:%v:%v:%v:%v:%v",
			nodeID, devID, ctxID, clusterID, engineType, engineID, name),
	)
	assert.Assert(err == nil, "Must be nil error: %v", err)
}
