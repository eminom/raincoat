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

	sqlStmt := getDbInitSchema()
	_, err = db.Exec(getDbInitSchema())
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	// Open one transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into dtu_op(idx, name) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Open another
	tx1, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt1, err := tx1.Prepare("insert into header(table_name) values(?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt1.Close()

	// Cannot do this
	// stmt1.Exec("hello-world")
	t.Log("push some data into dtu-op")

	// Put some data into the tables
	for i := 0; i < 1; i++ {
		_, err = stmt.Exec(i, fmt.Sprintf("dtu-op(%v)", i))
		if err != nil {
			log.Fatal(err)
		}
	}

	//Finalize
	tx.Commit()
	t.Log("tx commited")
	tx1.Commit()
}
