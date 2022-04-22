package dbexport

import (
	"database/sql"
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

const (
	createMemcpyTable = `
	CREATE TABLE memcpy(idx INT,name TEXT,node_id INT,description TEXT,context_id INT,
		start_timestamp INT,end_timestamp INT,duration_timestamp INT,
		start_cycle INT,end_cycle INT,duration_cycle INT,
		packet_id INT,device_id INT,cluster_id INT,engine_id INT,
		engine_type TEXT,op_id INT,op_name TEXT,
		src_addr INT,dst_addr INT,src_size INT,dst_size INT,
		direction TEXT,tiling_mode TEXT,vc INT,
		args TEXT,vp_id INT,row_name TEXT,tid TEXT);`
)

func init() {
	RegisterTabInitCommand(createMemcpyTable)
}

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
