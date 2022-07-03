package archtarget

type ArchTarget struct {
	CdmaPerC    int
	SdmaPerC    int
	SipPerC     int
	CqmPerC     int
	GsyncPerC   int
	ClusterPerD int
	MaxMasterId int
}

type ArchPgTarget struct {
	ArchTarget
	SipPerPg             int
	SipPgGroupPerCluster int
	MaxPgOrderIndex      int
}

func (at ArchTarget) GetCdmaCount() int {
	return at.ClusterPerD * at.CdmaPerC
}

func (at ArchTarget) GetMaxMasterId() int {
	return at.MaxMasterId
}

func (ad ArchTarget) GetClusterCount() int {
	return ad.ClusterPerD
}

func (ad ArchPgTarget) GetMaxPgOrderIndex() int {
	return ad.MaxPgOrderIndex
}
