package mimics

import (
	"reflect"
	"testing"
)

func TestTableInitGen(t *testing.T) {
	tg := TableInitGen{
		TableName: "Platform",
		Elements: []SubElement{
			{"User", reflect.TypeOf("")},
		},
	}
	t.Logf("%v\n", tg.String())
}
