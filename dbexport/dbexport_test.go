package dbexport

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestOne(t *testing.T) {
	targetName := "./foo.db"
	os.Remove(targetName)
	db, err := sql.Open("sqlite3", targetName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := createDtuOpTable + `
	delete from dtu_op;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into dtu_op(idx, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	for i := 0; i < 1; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("dtu-op(%v)", i))
		if err != nil {
			log.Fatal(err)
		}
	}
	tx.Commit()
}
