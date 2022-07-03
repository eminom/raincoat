package archtarget

func NewDoradoArchTarget() ArchTarget {
	return ArchTarget{
		CdmaPerC:    4,
		SdmaPerC:    12,
		SipPerC:     12,
		CqmPerC:     3,
		GsyncPerC:   3,
		ClusterPerD: 2,
		MaxMasterId: 1024,
	}
}

func NewDoradoArchPgTarget() ArchPgTarget {
	baseTarget := NewDoradoArchTarget()
	const DoradoSipPerPg = 4
	SipPgGroupPerCluster := baseTarget.SipPerC / DoradoSipPerPg
	return ArchPgTarget{
		ArchTarget:           baseTarget,
		SipPerPg:             DoradoSipPerPg,
		SipPgGroupPerCluster: SipPgGroupPerCluster,
		MaxPgOrderIndex:      6,
	}
}
