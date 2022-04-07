package rtdata

type CommandInfo struct {
	Command        string
	StartTs, EndTs uint64
}

type VersionInfo struct {
	SdkVersion      string
	FrameworkVer    string
	ProfileDataName string
	ProfileDataType string
	ProfileDataVer  string
}

type HostInfo struct {
	CommandInfo CommandInfo
	VerInfo     VersionInfo
}
