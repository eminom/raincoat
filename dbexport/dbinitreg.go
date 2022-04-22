package dbexport

type DbTabGen struct {
	initCmd []string
}

var (
	dbTab DbTabGen
)

func (d *DbTabGen) Append(cmd string) {
	d.initCmd = append(d.initCmd, cmd)
}

//Forward
func RegisterTabInitCommand(cmd string) {
	dbTab.Append(cmd)
}

func GetRegisteredInitCmds() []string {
	return dbTab.initCmd
}
