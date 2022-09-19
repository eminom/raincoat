package vgrule

import (
	"math/bits"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/efintf/affinity"
)

type EngineMaterIdStringMap interface {
	CheckoutEngineString(int) string
	MayCheckoutEngineString(int) (string, bool)
}

type EngineOrder interface {
	EngineMaterIdStringMap
	GetEngineOrderIndex(codec.DpfEvent) int
	GetCqmEngineOrder(codec.DpfEvent) int
	GetSipEngineOrder(codec.DpfEvent) int
	GetCdmaPgBitOrder(codec.DpfEvent) int
	GetSdmaPgBitOrder(codec.DpfEvent) int
	MapPgMaskBitsToCdmaEngineMask(pgMask int) int
	MapPgMaskBitsToSdmaEngineMask(pgMask int) int
	GetMaxPgOrderIndex() int
	GetMaxMasterId() int
}

type MasterValueDecoder interface {
	DecodeMasterValue(int) (codec.EngineTypeCode, int, int)
}

type ActMatchAlgo interface {
	EngineOrder
	MasterValueDecoder
	GetChannelNum() int
	MapToChan(masterValue, ctx int) int
	DecodeChan(chNum int) (int, int)
}

type doradoRule struct {
	codec.ArchDispatcher
	mDecoder     MasterValueDecoder
	cdmaAffinity affinity.CdmaAffinitySet
}

func NewDoradoRule(decoder MasterValueDecoder,
	cdmaAffinity affinity.CdmaAffinitySet) *doradoRule {
	return &doradoRule{
		ArchDispatcher: codec.MakeDoradoCollectDispatch(),
		mDecoder:       decoder,
		cdmaAffinity:   cdmaAffinity,
	}
}

func (a doradoRule) DecodeMasterValue(val int) (codec.EngineTypeCode, int, int) {
	return a.mDecoder.DecodeMasterValue(val)
}

func (a doradoRule) GetChannelNum() int {
	return codec.MASTERVALUE_COUNT * codec.RTCONTEXT_COUNT
}

func (a doradoRule) MapToChan(masterValue, ctx int) int {
	return masterValue<<codec.RTCONTEXT_BITCOUNT + ctx
}

func (a doradoRule) DecodeChan(chNum int) (int, int) {
	return chNum >> codec.RTCONTEXT_BITCOUNT,
		chNum & (1<<codec.RTCONTEXT_BITCOUNT - 1)
}

/*
For now only 5 kinds of engines are bound to Pg
*/
func (a doradoRule) GetEngineOrderIndex(dpf codec.DpfEvent) int {
	switch dpf.EngineTypeCode {
	case codec.EngCat_CDMA:
		return a.GetCdmaPgBitOrder(dpf)
	case codec.EngCat_GSYNC,
		codec.EngCat_CQM:
		return a.GetCqmEngineOrder(dpf)
	case codec.EngCat_SIP:
		return a.GetSipEngineOrder(dpf)
	case codec.EngCat_SDMA:
		return a.GetSdmaPgBitOrder(dpf)
	}
	// negative shift is not allowed in Golang
	return -1
}

/*
Get engine order for CQM/GSYNC
6 pg, 6 CQM/GSYNC (master engine)
*/
func (a doradoRule) GetCqmEngineOrder(dpf codec.DpfEvent) int {
	return dpf.EngineIndex + a.CqmPerC*dpf.ClusterID
}

func (a doradoRule) GetCdmaPgBitOrder(dpf codec.DpfEvent) int {
	return a.cdmaAffinity.GetCdmaIdxToPg(dpf.ClusterID, dpf.EngineIndex)
}

/*
  SIP 0~3  pgbit    1   ---> into 0
  SIP 4~7  pgbit   10   ---> into 1
  SIP 8~11 pgbit  100   ---> into 2
*/
func (a doradoRule) GetSipEngineOrder(dpf codec.DpfEvent) int {
	return dpf.EngineIndex/a.SipPerPg + a.SipPgGroupPerCluster*dpf.ClusterID
}

/*
12 SIPs per cluster
0~3   pg 0
4~7   pg 1
8~11  pg 2
Going to fit into the 1 of the 6 bits (6 PGs for dorado)
*/
func (a doradoRule) GetSdmaPgBitOrder(dpf codec.DpfEvent) int {
	return dpf.EngineIndex/a.SipPerPg + a.SipPgGroupPerCluster*dpf.ClusterID
}

func (a doradoRule) MapPgMaskBitsToCdmaEngineMask(pgMask int) int {
	// For now(Until Jan.16.2022)
	// Only single PG is implemented
	// SO TODO
	if bits.OnesCount(uint(pgMask)) == 1 {
		return pgMask
	}
	// Could be more
	assert.Assert(false, "No a single bit pg-mask")
	return 0
}

func (a doradoRule) MapPgMaskBitsToSdmaEngineMask(pgMask int) int {
	// For now(Until Jan.16.2022)
	// Only single PG is implemented
	// SO TODO
	if bits.OnesCount(uint(pgMask)) == 1 {
		return pgMask
	}
	// Could be more
	assert.Assert(false, "No a single bit pg-mask")
	return 0
}
