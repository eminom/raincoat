package dbexport

import "strings"

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
		createCommandTable,
		createVersionTable,
		createPlatform,
	}
	tablesCreate = append(tablesCreate, GetRegisteredInitCmds()...)
	return strings.Join(tablesCreate, "\n") + `
	delete from dtu_op;
	`
}
