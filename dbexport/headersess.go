package dbexport

import (
	"database/sql"
	"fmt"
)

type HeaderSession struct {
	TableSession
}

func NewHeaderSess(db *sql.DB) *HeaderSession {
	return &HeaderSession{
		TableSession: NewTableSession(db,
			`insert into header(table_name, version, category,
			count, time_unit) values(?, ?, ?, ?, ?)`),
	}
}

func (hs *HeaderSession) AddHeader(tableName string, version string, category string,
	count int, timeUnit string) {
	hs.stmt.Exec(tableName,
		version, category, fmt.Sprintf("%v", count),
		timeUnit)
}
