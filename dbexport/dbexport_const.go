package dbexport

import "strings"

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

const (
	createKernelTable = `
	CREATE TABLE kernel(idx INT,name TEXT,node_id INT,description TEXT,context_id INT,
		start_timestamp INT,end_timestamp INT,duration_timestamp INT,
		start_cycle INT,end_cycle INT,duration_cycle INT,packet_id INT,
		device_id INT,cluster_id INT,engine_id INT,engine_type TEXT,
		op_id INT,op_name TEXT,vp_id INT,row_name TEXT,tid TEXT);`
)

const (
	createCommandTable = `
	CREATE TABLE command(command TEXT,start_timestamp INT64,end_timestamp INT64);
`
)

const (
	createVersionTable = `
	CREATE TABLE version(sdk_version TEXT,framework_version TEXT,
		profile_data_name TEXT,
		profile_data_type TEXT,profile_data_version TEXT);`
)

const (
	createPlatform = `
	CREATE TABLE platform(product TEXT,
		platform TEXT,os_name TEXT,
		os_version TEXT,os_release TEXT,
		host_name TEXT,arch TEXT,cpu_model TEXT,cpu_vendor TEXT,
		distribution_name TEXT);`
)

const (
	TableCategory_DTUOpActivity     = "DTUOpActivity"
	TableCategory_DTUFwActivity     = "DTUFwActivity"
	TableCategory_DTUMemcpyActivity = "DTUMemcpyActivity"
	TableCategory_DTUKernelActivity = "DTUKernelActivity"
	TableCategory_CommandInfo       = "CommandInfo"
	TableCategory_VersionInfo       = "SotwareVersionInfo"
	TableCategroy_Platform          = "PlatformInfo"
)

func getDbInitSchema() string {

	tablesCreate := []string{
		createHeaderTable,
		createDtuOpTable,
		createFwTable,
		createKernelTable,
		createCommandTable,
		createVersionTable,
		createPlatform,
		createMemcpyTable,
	}

	return strings.Join(tablesCreate, "\n") + `
	delete from dtu_op;
	`
}
