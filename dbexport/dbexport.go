package dbexport

import (
	"database/sql"
	"log"
	"os"

	"git.enflame.cn/hai.bai/dmaster/rtinfo"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	_ "github.com/mattn/go-sqlite3"
)

type ExtractOpInfo func(rtdata.OpActivity) (bool, string, string)

type AddOpTrace func(
	idx, nodeID, devID, clusterID, ctxID int, name string,
	startTS, endTS, durTS uint64,
	startCy, endCy, durCy uint64,
	opId int, opName string)

type AddFwTrace func(
	idx, nodeID, devID, clusterID, ctxID int, name string,
	startTS, endTS, durTS uint64,
	startCy, endCy, durCy uint64,
	packetId int, engineType string,
)

type DbSession struct {
	targetName string
	finish     func()
	addOpTrace AddOpTrace
	addFwTrace AddFwTrace
	idx        int
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

	// Prepare a header for you

	hs := NewHeaderSess(db)
	dos := NewDtuOpSession(db)
	fw := NewFwSession(db)

	finishDbWork := func() {
		// Finalize Dtu-ops
		dos.Close()

		// Finalize fw traces
		fw.Close()

		// Finalize headers, not until the end do we know the count
		hs.AddHeader("dtu_op", "1.0",
			TableCategory_DTUOpActivity, dos.dtuOpCount, "ns")
		hs.AddHeader("fw", "1.0",
			TableCategory_DTUFwActivity, fw.fwOpCount, "ns")
		hs.Close()

		// And finally , close DB handle
		db.Close()
	}

	return &DbSession{
		finish:     finishDbWork,
		addOpTrace: dos.AddDtuOp,
		addFwTrace: fw.AddFwTrace,
	}, nil
}

func (dbs *DbSession) Close() {
	dbs.finish()
}

func (dbs *DbSession) DumpToEventTrace(
	bundle []rtdata.OpActivity,
	tm *rtinfo.TimelineManager,
	extractor ExtractOpInfo,
	dumpWild bool,
) {
	dtuOpCount := 0
	convertToHostError := 0
	const nodeID = 0
	const deviceID = 0
	const clusterID = -1
	for _, act := range bundle {
		///act.IsOpRefValid()
		if okToShow, _, name := extractor(act); okToShow {
			dtuOpCount++
			startHostTime, startOK := tm.MapToHosttime(act.StartCycle())
			endHostTime, endOK := tm.MapToHosttime(act.EndCycle())
			if startOK && endOK {
				dbs.addOpTrace(dbs.idx, nodeID, deviceID, clusterID, act.Start.Context, name,
					startHostTime, endHostTime, endHostTime-startHostTime,
					act.StartCycle(), act.EndCycle(), act.EndCycle()-act.StartCycle(),
					act.GetOp().OpId, act.GetOp().OpName,
				)
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
