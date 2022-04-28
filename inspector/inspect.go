package inspector

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

func InspectMemcpy(out io.Writer, db *sql.DB, engineType string) {
	rows, err := db.Query(
		fmt.Sprintf("SELECT distinct(tiling_mode) FROM memcpy where engine_type=\"%v\"",
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

	tilingCount := make(map[string]int)
	for tiling := range tilingSet {
		rows, err = db.Query(
			fmt.Sprintf("SELECT COUNT(*) FROM memcpy WHERE tiling_mode=\"%v\" AND engine_type=\"%v\"",
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

	if hasEmpty {
		cmd := fmt.Sprintf(
			"SELECT distinct(packet_id) FROM memcpy WHERE tiling_mode is null and engine_type=\"%v\"",
			engineType)
		fmt.Fprintf(out, "cmd is %v\n", cmd)
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
		} else {
			fmt.Fprintf(out, "all DMA meta restored\n")
		}

		// for pid := range pids {
		// 	rows, err = db.Query(
		// 		fmt.Sprintf("SELECT count(*) from memcpy where tiling_mode=\"\" and engine_type=\"%v\" and packet_id=%v",
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
		// }
	}
	fmt.Fprintf(out, "\n")
}

func GetDistinctEngineTypes(db *sql.DB) map[string]bool {
	rows, err := db.Query(
		"SELECT distinct(engine_type) FROM memcpy",
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

func InspectMain(out io.Writer, targetName string) {
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
	sort.Strings(engineArr)
	for _, engine := range engineArr {
		InspectMemcpy(out, db, engine)
	}
}
