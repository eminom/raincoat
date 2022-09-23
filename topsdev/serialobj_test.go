package topsdev

import (
	"testing"
)

func TestSerialObjFormat(t *testing.T) {
	SerialObjEnc{}.makeProfileData([]byte{}, 1)
	t.Logf("done")
}
