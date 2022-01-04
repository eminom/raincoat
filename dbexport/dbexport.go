package dbexport

import (
	"database/sql"
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
	string, DatabaseChannelType)

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

	dtuOpCount int
	fwOpCount  int
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

	sqlStmt := createHeaderTable + "\n" +
		createDtuOpTable + "\n" +
		createFwTable + `
	delete from dtu_op;
	`
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
	hs.Close()
	// And finally , close DB handle
	dbs.dbObject.Close()
}

func (dbs *DbSession) DumpToEventTrace(
	bundle []rtdata.OpActivity,
	tm *rtinfo.TimelineManager,
	extractor ExtractOpInfo,
	dumpWild bool,
) {
	fw := NewFwSession(dbs.dbObject)
	// Finalize fw traces
	defer fw.Close()
	dos := NewDtuOpSession(dbs.dbObject)
	defer dos.Close()

	dtuOpCount := 0
	convertToHostError := 0
	const nodeID = 0
	const deviceID = 0
	const clusterID = -1
	for _, act := range bundle {
		///act.IsOpRefValid()
		if okToShow, _, name, whereToGo := extractor(act); okToShow {
			dtuOpCount++
			startHostTime, startOK := tm.MapToHosttime(act.StartCycle())
			endHostTime, endOK := tm.MapToHosttime(act.EndCycle())
			if startOK && endOK {
				switch whereToGo {
				case DbChannel_DtuOp:
					dos.AddDtuOp(dbs.idx, nodeID, deviceID, clusterID, act.Start.Context, name,
						startHostTime, endHostTime, endHostTime-startHostTime,
						act.StartCycle(), act.EndCycle(), act.EndCycle()-act.StartCycle(),
						act.GetOp().OpId, act.GetOp().OpName,
					)
					dbs.dtuOpCount++
				case DbChannel_FW:
					packetID, contextID := 0, -1
					if act.Start.EngineTypeCode == codec.EngCat_CQM {
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
				}
				dbs.idx++
			} else {
				convertToHostError++
			}
		}
	}

	log.Printf("%v dtu-op record(s) have been traced into %v",
		dtuOpCount,
		dbs.targetName,
	)
}
