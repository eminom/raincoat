package codec

import (
	"log"
	"testing"
)

func TestCodecBasic(t *testing.T) {
	start := int(EngCat_SIP)
	end := int(EngCat_UNKNOWN)
	for i := start; i <= end; i++ {
		v := EngineTypeCode(i)
		log.Printf("%v: %v", v.ToString(), v)
		if ToEngineTypeCode(v.ToString()) != v {
			log.Fatalf("error codec for %v\n", v)
		}
	}
	log.Printf("all %v codes are verified", (end - start + 1))
}
