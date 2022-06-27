package affinity

type CdmaAffinitySet interface {
	GetCdmaIdxToPg(cid int, eid int) int
}
