package dbexport

import "sync/atomic"

var (
	__vpid int32
)

func GetNextVpId() int {
	return int(atomic.AddInt32(&__vpid, 1))
}
