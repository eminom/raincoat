package dbexport

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/rtinfo"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	_ "github.com/mattn/go-sqlite3"
)

type DatabaseChannelType string

const (
	DbChannel_DtuOp DatabaseChannelType = "channel.DtuOp"
	DbChannel_FW    DatabaseChannelType = "channel.FW"
)

type ExtractOpInfo func(rtdata.OpActivity) (bool, string,
	string)

type AddOpTrace func(
	idx, nodeID, devID, clusterID, ctxID int, name string,
	startTS, endTS, durTS uint64,
	startCy, endCy, durCy uint64,
	opId int, opName string)

type AddFwTrace func(idx, nodeID, devID, clusterID, ctxID int, name string,
	startTS, endTS, durTS uint64,
	startCy, endCy, durCy uint64,
	packetId int, engineType string,
	engineID int,
)

type DbSession struct {
	targetName string
	dbObject   *sql.DB
	idx        int

	dtuOpCount    int
	fwOpCount     int
	dmaOpCount    int
	kernelOpCount int
}

func ifFileExist(file string) bool {
	stat, err := os.Stat(file)
	return nil == err && !stat.IsDir()
}

func NewDbSession(target string) (*DbSession, error) {
	if ifFileExist(target) {
		os.Remove(target)
	}

	db, err := sql.Open("sqlite3", target)
	if err != nil {
		return nil, err
	}

	sqlStmt := getDbInitSchema()
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil, err
	}

	return &DbSession{
		dbObject: db,
	}, nil
}

func (dbs *DbSession) Close() {
	log.Printf("finish db session")
	hs := NewHeaderSess(dbs.dbObject)
	// Finalize headers, not until the end do we know the count
	hs.AddHeader("dtu_op", "1.0",
		TableCategory_DTUOpActivity, dbs.dtuOpCount, "ns")
	hs.AddHeader("fw", "1.0",
		TableCategory_DTUFwActivity, dbs.fwOpCount, "ns")
	hs.AddHeader("memcpy", "1.0",
		TableCategory_DTUMemcpyActivity, dbs.dmaOpCount, "ns")
	hs.AddHeader("kernel", "1.0",
		TableCategory_DTUKernelActivity, dbs.kernelOpCount, "ns")
	hs.Close()
	// And finally , close DB handle
	dbs.dbObject.Close()
}

func (dbs *DbSession) DumpDtuOps(
	coords rtdata.Coords,
	bundle []rtdata.OpActivity,
	tm *rtinfo.TimelineManager,
	extractor ExtractOpInfo,
) {
	dos := NewDtuOpSession(dbs.dbObject)
	defer dos.Close()
	dtuOpCount, convertToHostError := 0, 0
	nodeID, deviceID := coords.NodeID, coords.DeviceID
	const clusterID = -1
	for _, act := range bundle {
		if okToShow, _, name := extractor(act); okToShow {
			dtuOpCount++
			startHostTime, startOK := tm.MapToHosttime(act.StartCycle())
			endHostTime, endOK := tm.MapToHosttime(act.EndCycle())
			if startOK && endOK {
				dos.AddDtuOp(dbs.idx, nodeID, deviceID, clusterID, act.Start.Context, name,
					startHostTime, endHostTime, endHostTime-startHostTime,
					act.StartCycle(), act.EndCycle(), act.EndCycle()-act.StartCycle(),
					act.GetOp().OpId, act.GetOp().OpName,
				)
				dbs.dtuOpCount++
				dbs.idx++
			} else {
				convertToHostError++
			}
		}
	}
	if convertToHostError > 0 {
		fmt.Printf("error: DTU-Op convert-time error: %v\n", convertToHostError)
	}
	log.Printf("%v DTU-OPs have been traced into %v",
		dtuOpCount,
		dbs.targetName,
	)
}

func (dbs *DbSession) DumpFwActs(
	coords rtdata.Coords,
	bundle []rtdata.FwActivity,
	tm *rtinfo.TimelineManager,
) {
	fw := NewFwSession(dbs.dbObject)
	defer fw.Close()

	fwActCount, convertToHostError := 0, 0
	nodeID, deviceID := coords.NodeID, coords.DeviceID
	for _, act := range bundle {

		getName := func(act rtdata.FwActivity) string {
			switch act.Start.EngineTypeCode {
			case codec.EngCat_TS:
				str, _ := rtdata.ToTSEventString(act.Start.Event)
				return str
			case codec.EngCat_CQM, codec.EngCat_GSYNC:
				str, _ := rtdata.ToCQMEventString(act.Start.Event)
				return str
			}
			return fmt.Sprintf("Engine(%v)", act.Start.EngineTy)
		}
		name := getName(act)

		fwActCount++
		startHostTime, startOK := tm.MapToHosttime(act.StartCycle())
		endHostTime, endOK := tm.MapToHosttime(act.EndCycle())
		if startOK && endOK {
			packetID, contextID := 0, -1
			switch act.Start.EngineTypeCode {
			case codec.EngCat_CQM, codec.EngCat_GSYNC:
				packetID = act.Start.PacketID
				contextID = act.Start.Context
			}
			fw.AddFwTrace(dbs.idx, nodeID, deviceID, act.Start.ClusterID, contextID, name,
				startHostTime, endHostTime, endHostTime-startHostTime,
				act.StartCycle(), act.EndCycle(), act.EndCycle()-act.StartCycle(),
				packetID, act.Start.EngineTy,
				act.Start.EngineIndex,
			)
			dbs.fwOpCount++
			dbs.idx++
		} else {
			convertToHostError++
		}

	}
	if convertToHostError > 0 {
		fmt.Printf("error: FW ACT convert-time error: %v\n", convertToHostError)
	}
	log.Printf("%v FW record(s) have been traced into %v",
		fwActCount,
		dbs.targetName,
	)
}

func (dbs *DbSession) DumpDmaActs(
	coords rtdata.Coords,
	bundle []rtdata.DmaActivity,
	tm *rtinfo.TimelineManager,
) {
	dmaS := NewDmaSession(dbs.dbObject)
	defer dmaS.Close()

	dmaActCount, convertToHostError := 0, 0
	nodeID, deviceID := coords.NodeID, coords.DeviceID

	const TimeErrDisplayLimit = 10
	timeErrCount := 0
	for _, act := range bundle {

		dmaActCount++
		startHostTime, startOK := tm.MapToHosttime(act.StartCycle())
		endHostTime, endOK := tm.MapToHosttime(act.EndCycle())
		if startOK && endOK {
			packetID, contextID := act.Start.PacketID, act.Start.Context

			name, _ := rtdata.ToDmaEventString(act.Start.Event)
			tilingMode := act.Start.EngineTy // Unknown tiling mode(Slice,Transpose, etc)
			if act.IsDmaMetaRefValid() {
				dmaMeta := act.GetDmaMeta()
				tilingMode = dmaMeta.DmaOpString
			}

			dmaS.AddDmaTrace(dbs.idx, nodeID, deviceID, act.Start.ClusterID,
				contextID, name,
				startHostTime, endHostTime, endHostTime-startHostTime,
				act.StartCycle(), act.EndCycle(), uint64(act.Duration()),
				packetID, act.Start.EngineTy,
				tilingMode,
				act.GetEngineIndex(),
				act.GetVcId(),
			)
			dbs.dmaOpCount++
			dbs.idx++
		} else {
			convertToHostError++
			timeErrCount++
			if timeErrCount < TimeErrDisplayLimit {
				if !startOK {
					fmt.Printf("start_cycle: %v\n", act.StartCycle())
				}
				if !endOK {
					fmt.Printf("end_cycle: %v\n", act.EndCycle())
				}
			} else if timeErrCount == TimeErrDisplayLimit {
				fmt.Printf("too many DtoH time convert error\n")
			}
		}
	}
	if convertToHostError > 0 {
		fmt.Printf("error: DMA ACT convert-time error: %v\n", convertToHostError)
	}
	log.Printf("%v DMA ACT record(s) have been traced into %v",
		dmaActCount,
		dbs.targetName,
	)

}

func (dbs *DbSession) DumpKernelActs(
	coords rtdata.Coords,
	bundle []rtdata.KernelActivity,
	tm *rtinfo.TimelineManager,
) {
	dmaS := NewKernelSession(dbs.dbObject)
	defer dmaS.Close()
	nodeID, deviceID := coords.NodeID, coords.DeviceID
	for _, act := range bundle {
		startHostTime, startOK := tm.MapToHosttime(act.StartCycle())
		endHostTime, endOK := tm.MapToHosttime(act.EndCycle())
		if startOK && endOK {
			packetID, contextID := act.Start.PacketID, act.Start.Context
			name := "SIP BUSY"
			dmaS.AddKernelTrace(
				dbs.idx, nodeID, deviceID, act.Start.ClusterID,
				contextID, name,
				startHostTime, endHostTime, endHostTime-startHostTime,
				act.StartCycle(), act.EndCycle(), uint64(act.Duration()),
				packetID, act.Start.EngineTy,
				act.GetEngineIndex(),
			)
			dbs.kernelOpCount++
			dbs.idx++
		}
	}
	log.Printf("%v SIP ACT record(s) have been traced into %v",
		len(bundle),
		dbs.targetName,
	)

}
