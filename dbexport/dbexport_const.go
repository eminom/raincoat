package dbexport

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

const (
	createHeaderTable = `
	CREATE TABLE header(table_name TEXT,version TEXT,category TEXT,count INT,time_unit TEXT);`
)

const (
	createFwTable = `
	CREATE TABLE fw(idx INT,name TEXT,node_id INT,description TEXT,context_id INT,
		start_timestamp INT,end_timestamp INT,duration_timestamp INT,
		start_cycle INT,end_cycle INT,duration_cycle INT,
		packet_id INT,device_id INT,cluster_id INT,engine_id INT,
		engine_type TEXT,args TEXT,vp_id INT,row_name TEXT,tid TEXT);`
)

const (
	TableCategory_DTUOpActivity = "DTUOpActivity"
	TableCategory_DTUFwActivity = "DTUFwActivity"
)
