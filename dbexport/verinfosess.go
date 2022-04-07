package dbexport

import (
	"database/sql"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

type VerInfoSession struct {
	TableSession
}

func NewVerInfoSession(db *sql.DB) *VerInfoSession {
	return &VerInfoSession{
		TableSession: NewTableSession(db,
			`insert into version(
				sdk_version, framework_version
			) values(?, ?)`),
	}
}

func (cmdS *VerInfoSession) AddVerInfo(sdk, framework string) {
	_, err := cmdS.stmt.Exec(sdk, framework)
	assert.Assert(err == nil, "Must be nil error: %v", err)
}
