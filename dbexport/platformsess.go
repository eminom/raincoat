package dbexport

import (
	"database/sql"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

type PlatformInfoSession struct {
	TableSession
}

func NewPlatformInfoSession(db *sql.DB) *PlatformInfoSession {
	return &PlatformInfoSession{
		TableSession: NewTableSession(db,
			`insert into platform(
				product
			) values(?)`),
	}
}

func (platS *PlatformInfoSession) AddPlatform(product string) {
	_, err := platS.stmt.Exec(product)
	assert.Assert(err == nil, "Must be nil error: %v", err)
}
