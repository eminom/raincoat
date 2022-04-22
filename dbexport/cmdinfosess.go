package dbexport

import (
	"database/sql"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

type CmdInfoSession struct {
	TableSession
}

func NewCmdInfoSession(db *sql.DB) *CmdInfoSession {
	return &CmdInfoSession{
		TableSession: NewTableSession(db,
			`insert into command(
				command, start_timestamp, end_timestamp
			) values(?, ?, ?)`),
	}
}

func (cmdS *CmdInfoSession) AddCmdInfo(cmd string, start, end int64) {
	_, err := cmdS.stmt.Exec(cmd, start, end)
	assert.Assert(err == nil, "Must be nil error: %v", err)
}
