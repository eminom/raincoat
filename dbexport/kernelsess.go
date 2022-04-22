package dbexport

import (
	"database/sql"
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

const (
	createKernelTable = `
	CREATE TABLE kernel(idx INT,name TEXT,node_id INT,description TEXT,context_id INT,
		start_timestamp INT,end_timestamp INT,duration_timestamp INT,
		start_cycle INT,end_cycle INT,duration_cycle INT,packet_id INT,
		device_id INT,cluster_id INT,engine_id INT,engine_type TEXT,
		op_id INT,op_name TEXT,vp_id INT,row_name TEXT,tid TEXT);`
)

func init() {
	RegisterTabInitCommand(createKernelTable)
}

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
	engineID int,
	rowName string) {
	//0:0:-1:2:ENGINE_SIP:0:SIP BUSY
	// And SIP BUSY only so far.
	// row_name as name
	_, err := kernS.stmt.Exec(idx, nodeID, devID, clusterID, ctxID, name,
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
