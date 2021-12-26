package decmconf

import "path/filepath"

type DecmConf struct {
	startupPath string
}

func NewDecmConf(startup string) DecmConf {
	return DecmConf{
		startupPath: startup,
	}
}

func (d DecmConf) GetRuntimeTaskPath() string {
	return filepath.Join(d.startupPath, "runtime_task.txt")
}

func (d DecmConf) GetMetaStartupPath() string {
	return d.startupPath
}

func (d DecmConf) GetTimepointsPath() string {
	return filepath.Join(d.startupPath, "timepoints.txt")
}
