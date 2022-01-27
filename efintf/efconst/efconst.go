package efconst

const (
	SolidTaskID = 102030405060
)

func IsWildcardExecuuid(execUuid uint64) bool {
	return execUuid == 0
}

func IsAllZeroPgMask(pgMask int) bool {
	return pgMask == 0
}
