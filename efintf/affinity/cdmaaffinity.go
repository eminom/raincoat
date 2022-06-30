package affinity

import "git.enflame.cn/hai.bai/dmaster/efintf/archtarget"

type CdmaAffinity struct {
	PgIndex int
	Cid     int
	Eid     int
}

type DoradoCdmaAffinityDefault struct{}

func (DoradoCdmaAffinityDefault) GetCdmaIdxToPg(cid int, eid int) int {
	// (0, 0), (0, 1), (0, 2), pg 012 if single
	// (1, 0), (1, 1), (1, 2), pg 345 if single
	if eid == 3 {
		// (0, 3), the fourth CDMA engine is mapped to (3pg pg000111)
		// (1, 3), the fourth CDMA engine is mapped to (3pg pg111000)
		return cid * 3
	}
	return cid*3 + eid
}

type DoradoCdmaAffinity struct {
	toPgIdx    []int
	archTarget archtarget.ArchTarget
}

func NewDoradoCdmaAffinity(
	cdmaAffinity []CdmaAffinity,
	arch archtarget.ArchTarget) DoradoCdmaAffinity {
	affinityMap := make([]int, arch.CdmaPerC*arch.ClusterPerD)

	def := DoradoCdmaAffinityDefault{}
	for cid := 0; cid < arch.ClusterPerD; cid++ {
		for eid := 0; eid < arch.CdmaPerC; eid++ {
			affinityMap[cid*arch.CdmaPerC+eid] = def.GetCdmaIdxToPg(cid, eid)
		}
	}

	for _, cdma := range cdmaAffinity {
		idx := arch.CdmaPerC*cdma.Cid + cdma.Eid
		affinityMap[idx] = cdma.PgIndex
	}
	return DoradoCdmaAffinity{
		toPgIdx:    affinityMap,
		archTarget: arch,
	}
}

func (c2p DoradoCdmaAffinity) GetCdmaIdxToPg(cid int, eid int) int {
	idx := cid*c2p.archTarget.CdmaPerC + eid
	return c2p.toPgIdx[idx]
}

func NewDoradoCdmaAffinityVersioning(target archtarget.ArchTarget) DoradoCdmaAffinity {
	return NewDoradoCdmaAffinity([]CdmaAffinity{
		{0, 0, 2}, {1, 0, 0}, {2, 0, 3}, {3, 1, 1}, {4, 1, 0}, {5, 1, 3},
	}, target)
}
