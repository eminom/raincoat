package dbexport

import (
	"database/sql"
	"log"
)

type TableSession struct {
	stmt      *sql.Stmt
	tx        *sql.Tx
	cmdString string
}

func NewTableSession(db *sql.DB, cmdString string) TableSession {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare(cmdString)
	if err != nil {
		// log.Fatal(err)
		log.Printf("error for %v", cmdString)
		panic(err)
	}
	return TableSession{
		stmt:      stmt,
		tx:        tx,
		cmdString: cmdString,
	}
}

func (tabSess *TableSession) Close() {
	tabSess.tx.Commit()
	tabSess.stmt.Close()
}
