package inspector

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

const ExtraCond = " and name like \"DMA VC%%\""

// const ExtraCond = " and name like \"DMA BUSY%%\""

// packet id distribution
func InspectMemcpy(out io.Writer, db *sql.DB, engineType string) map[int]int {
	rows, err := db.Query(
		fmt.Sprintf("SELECT distinct(tiling_mode) FROM memcpy where engine_type=\"%v\""+ExtraCond,
			engineType),
	)
	if err != nil {
		panic(err)
	}
	hasEmpty := false
	var tilingSet = make(map[string]bool)
	for rows.Next() {
		var tilingMode string
		err = rows.Scan(&tilingMode)
		if err != nil {
			panic(err)
		}
		if tilingMode == "" {
			hasEmpty = true
			continue
		}
		tilingSet[tilingMode] = true
	}
	rows.Close()

	//
	tilingCount := make(map[string]int)
	for tiling := range tilingSet {
		rows, err = db.Query(
			fmt.Sprintf("SELECT COUNT(*) FROM memcpy WHERE tiling_mode=\"%v\" AND engine_type=\"%v\""+ExtraCond,
				tiling,
				engineType),
		)
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			var count int
			err = rows.Scan(&count)
			if err != nil {
				panic(err)
			}
			tilingCount[tiling] = count
		}
		rows.Close()
	}

	var tiles []string
	for t := range tilingSet {
		tiles = append(tiles, t)
	}
	sort.Strings(tiles)
	fmt.Fprintf(out, "### Engine type for %v ###\n", engineType)
	for _, t := range tiles {
		fmt.Fprintf(out, "%20v\t%v\n", t, tilingCount[t])
	}

	fmt.Fprintf(out, "### Engine type for %v in details ###\n", engineType)
	pidSet := (func() map[int]bool {
		cmd := fmt.Sprintf("SELECT distinct(packet_id) FROM memcpy WHERE engine_type=\"%v\""+ExtraCond, engineType)
		rows, err := db.Query(cmd)
		if err != nil {
			panic(err)
		}
		pidSet := make(map[int]bool)
		for rows.Next() {
			var pktId int
			err = rows.Scan(&pktId)
			if err != nil {
				panic(err)
			}
			pidSet[pktId] = true
		}
		rows.Close()
		return pidSet
	})()

	pidCount := make(map[int]int)
	for pid := range pidSet {
		cmd :=
			fmt.Sprintf("SELECT count(*) FROM memcpy WHERE packet_id=\"%v\" and engine_type=\"%v\""+ExtraCond,
				pid, engineType)
		rows, err := db.Query(cmd)
		if err != nil {
			panic(err)
		}
		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				panic(err)
			}
		}
		pidCount[pid] = count
	}

	if hasEmpty {
		cmd := fmt.Sprintf(
			"SELECT distinct(packet_id) FROM memcpy WHERE (tiling_mode is null or TRIM(tiling_mode)=\"\") and engine_type=\"%v\""+ExtraCond,
			engineType)
		// fmt.Fprintf(out, "cmd is %v\n", cmd)
		rows, err = db.Query(cmd)
		if err != nil {
			panic(err)
		}

		pids := make(map[int]bool)
		for rows.Next() {
			var packetId int
			err = rows.Scan(&packetId)
			if err != nil {
				panic(err)
			}
			pids[packetId] = true
		}
		if len(pids) > 0 {
			fmt.Fprintf(out, "# Packet without meta found, distinct(pid) is #\n")
			printCc := 5
			for pid := range pids {
				fmt.Fprintf(out, "%v\n", pid)
				printCc--
				if printCc < 0 {
					break
				}
			}
			if printCc < 0 {
				fmt.Fprintf(out, ".... too many\n")
			}
		} else {
			fmt.Fprintf(out, "all DMA meta restored\n")
		}

		// for pid := range pids {
		// 	rows, err = db.Query(
		// 		fmt.Sprintf("SELECT count(*) from memcpy where (TRIM(tiling_mode)=\"\" or tiling_mode is null) and engine_type=\"%v\" and packet_id=%v",
		// 			engineType, pid),
		// 	)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	for rows.Next() {
		// 		var count int
		// 		err = rows.Scan(&count)
		// 		if err != nil {
		// 			panic(err)
		// 		}
		// 		fmt.Fprintf(out, "%v counts to %v\n", pid, count)
		// 	}
		//  rows.Close()
		// }
	}
	fmt.Fprintf(out, "\n")
	return pidCount
}

func GetDistinctEngineTypes(db *sql.DB) map[string]bool {
	rows, err := db.Query(
		"SELECT DISTINCT(engine_type) FROM memcpy",
	)
	if err != nil {
		panic(err)
	}
	rv := make(map[string]bool)
	for rows.Next() {
		var engineStr string
		err = rows.Scan(&engineStr)
		if err != nil {
			panic(err)
		}
		rv[engineStr] = true
	}
	return rv
}

func InspectXdma(out io.Writer, targetName string) map[string]map[int]int {
	db, err := sql.Open("sqlite3", targetName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	engines := GetDistinctEngineTypes(db)
	var engineArr []string
	for e := range engines {
		engineArr = append(engineArr, e)
	}

	distri := make(map[string]map[int]int)
	sort.Strings(engineArr)
	for _, engine := range engineArr {
		dis := InspectMemcpy(out, db, engine)
		distri[engine] = dis
	}
	return distri
}

// return true:  two map are exactly the same
func compareDicts(uObj, vObj string, lhs, rhs map[string]map[int]int, out io.Writer) bool {
	isSame := true

	allEngine := make(map[string]bool)
	for k := range lhs {
		allEngine[k] = true
	}
	for k := range rhs {
		allEngine[k] = true
	}
	for k := range allEngine {
		fmt.Fprintf(out, "# For engine(%v)\n", k)
		if _, ok := lhs[k]; !ok {
			fmt.Fprintf(out, "  %v is missing for %v\n", k, uObj)
			isSame = false
			continue
		}
		if _, ok := rhs[k]; !ok {
			fmt.Fprintf(out, "  %v is missing for %v\n", k, vObj)
			isSame = false
			continue
		}
		lset, rset := lhs[k], rhs[k]
		allPids := make(map[int]bool)
		for p := range lset {
			allPids[p] = true
		}
		for p := range rset {
			allPids[p] = true
		}
		diffCount := 0
		sameCount := 0
		for p := range allPids {
			if lset[p] != rset[p] {
				diffCount++
				if diffCount < 5 {
					fmt.Fprintf(out, "  %v Pid(%v) diffs in %v, %v\n", k, p, lset[p], rset[p])
				}
			} else {
				sameCount++
			}
		}
		if diffCount == 0 && sameCount > 0 {
			fmt.Fprintf(out, "  Entries distribution are the same, %v in all\n",
				sameCount)
		} else {
			fmt.Fprintf(out, "  Same entries count: %v\n", sameCount)
			fmt.Fprintf(out, "  Diff entries count: %v\n", diffCount)
		}

		if diffCount > 0 {
			isSame = false
		}
	}
	return isSame
}

func countElement(distr map[string]map[int]int) int {
	rv := 0
	for _, dis := range distr {
		for _, cc := range dis {
			rv += cc
		}
	}
	return rv
}

type CmpRecord struct {
	TextLog    string
	Filename   string
	EntryCount int
}

func InspectMain(files []string, out io.Writer) {
	var disArr []map[string]map[int]int

	var inspectRes []CmpRecord
	for i := 0; i < len(files); i++ {
		textOut := bytes.NewBuffer(nil)
		dis := InspectXdma(textOut, files[i])
		disArr = append(disArr, dis)
		inspectRes = append(inspectRes, CmpRecord{
			TextLog:    textOut.String(),
			Filename:   files[i],
			EntryCount: countElement(dis),
		})
	}

	anyDiff := false
	n := len(disArr)
	for i := 0; i < n; i++ {
		u := files[i]
		us := disArr[i]
		for j := i + 1; j < n; j++ {
			v := files[j]
			vs := disArr[j]
			textOut := bytes.NewBuffer(nil)
			fmt.Fprintf(textOut, "# Compare \"%v\" to \"%v\"\n", u, v)
			isEqual := compareDicts(u, v, us, vs, textOut)
			if isEqual {
				fmt.Fprintf(out, "# \"%v\" and \"%v\" are the same, in %v entries\n", u, v, inspectRes[i].EntryCount)
			} else {
				anyDiff = true
				fmt.Fprint(out, textOut.String())
			}
		}
	}

	if anyDiff {
		for _, res := range inspectRes {
			fmt.Fprintf(out, "# Res for %v\n", res.Filename)
			fmt.Fprint(out, res.TextLog)
		}
	}
}
