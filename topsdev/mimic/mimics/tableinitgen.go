package mimics

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
)

type SubElement struct {
	Name string
	Type reflect.Type
}

func tyToString(ty reflect.Type) (string, bool) {
	if ty.Kind() == reflect.Pointer {
		return tyToString(ty.Elem())
	}
	switch ty.Kind() {
	case reflect.Int:
	case reflect.Int32:
		return "int", true
	case reflect.Int64:
		return "int64", true
	case reflect.String:
		return "TEXT", true
	}
	return "", false
}

type TableInitGen struct {
	TableName string
	Elements  []SubElement
}

func (tg TableInitGen) String() string {
	return fmt.Sprintf("CREATE TABLE %v(", tg.TableName) +
		tg.seqs() + ")"
}

func (tg TableInitGen) seqs() string {
	buf := bytes.NewBuffer(nil)
	lz := len(tg.Elements)
	for i, v := range tg.Elements {
		tyStr, ok := tyToString(v.Type)
		if !ok {
			log.Printf("error could not stringify %v for type of %v in Table(%v)",
				v.Name, v.Type,
				tg.TableName)
			continue
		}
		fmt.Fprintf(buf, "%v %v", v.Name, tyStr)
		if i != lz-1 {
			fmt.Fprintf(buf, ",")
		}
	}
	return buf.String()
}
