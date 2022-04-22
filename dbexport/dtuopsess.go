package dbexport

import (
	"database/sql"
	"fmt"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

const (
	DtuOpRowName = "DTU Op Re"
)

const (
	createDtuOpTable = `
	CREATE TABLE dtu_op(idx INT,name TEXT,node_id INT,description TEXT,
		context_id INT,start_timestamp INT,
		end_timestamp INT,duration_timestamp INT,
		start_cycle INT,end_cycle INT,duration_cycle INT,
		device TEXT,op_id INT,op_name TEXT,kind TEXT,fusion_kind TEXT,
		input_shape TEXT,output_shape TEXT,layer_kind TEXT,
		layer_name TEXT,
		module_id INT,module_name TEXT,meta TEXT,device_id INT,
		cluster_id INT,vp_id INT,row_name TEXT,tid TEXT);`
)

func init() {
	RegisterTabInitCommand(createDtuOpTable)
}

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
	opId int, opName string, rowName string) {

	moduleID := 1
	vpId := GetNextVpId()

	_, err := dos.stmt.Exec(
		idx, nodeID, devID, clusterID, ctxID, name,
		startTS, endTS, durTS,
		startCy, endCy, durCy,
		opId, opName,
		// and the default
		vpId, moduleID,
		rowName, fmt.Sprintf("%v:%v:%v:%v:%v",
			nodeID, devID, ctxID, clusterID, rowName,
		))
	assert.Assert(err == nil, "Must be nil to carry on:%v", err)
}
