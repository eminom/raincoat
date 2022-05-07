package codec

import (
	"testing"
)

func TestCodecBasic(t *testing.T) {
	start := int(EngCat_SIP)
	end := int(EngCat_UNKNOWN)
	for i := start; i <= end; i++ {
		v := EngineTypeCode(i)
		t.Logf("%v: %v", v.String(), v)
		if ToEngineTypeCode(v.String()) != v {
			t.Fatalf("error codec for %v\n", v)
			t.Fail()
		}
	}
	t.Logf("all %v codes are verified", (end - start + 1))
}
