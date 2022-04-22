package mimics

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/template"
)

const structTempl = `
// Generation of pb type {{.Name}}
type {{.Name}} struct{
  {{range .Fields}}{{.}}
  {{end}}
}`

const combinationTempl = `
// Generation for combinations
type HostInfo struct {
  {{range .}}{{.}}
  {{end}}
}`

type Fills struct {
	Name   string
	Fields []string
}

var (
	structBuilder = template.Must(template.New("structBuilder").Parse(structTempl))
	combBuilder   = template.Must(template.New("combBuilder").Parse(combinationTempl))
)

func toType(ty reflect.Type) string {
	if ty.Kind() == reflect.Pointer {
		return toType(ty.Elem())
	}
	switch ty.Kind() {
	case reflect.Int:
		return "int"
	case reflect.Int32:
		return "int32"
	case reflect.Int64:
		return "int64"
	case reflect.String:
		return "string"
	}
	return "string"
}

func isSimpleType(ty reflect.Type) bool {
	for ty.Kind() == reflect.Pointer {
		ty = ty.Elem()
	}
	switch ty.Kind() {
	case reflect.Struct:
		return false
	case reflect.Int,
		reflect.Int32,
		reflect.Int64,
		reflect.String:
		return true
	}
	return false
}

/*
 * Example:
 * `protobuf:"bytes,1,req,name=node2dev" json:"node2dev,omitempty"`
 */
func parseTag(s string) (string, bool) {
	const prefix = "protobuf:"
	const nameassign = "name="
	fs := strings.Fields(s)
	for _, f := range fs {
		if strings.HasPrefix(f, prefix) {
			str := f[len(prefix):]
			if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
				str = str[1 : len(str)-1]
				subs := strings.Split(str, ",")
				for _, sub := range subs {
					if strings.HasPrefix(sub, nameassign) {
						name := sub[len(nameassign):]
						return name, len(name) > 0
					}
				}
			}
		}
	}
	return "", false
}

type ObjDesc struct {
	Name     string
	Obj      interface{}
	IsRepeat bool
}

type ReflectSrc struct {
	Item     string
	GoStruct string
	TableGen string
}

// DFS append
func appendElement(elements *[]SubElement, fields *[]string, fd reflect.StructField) {
	if !fd.IsExported() {
		return
	}

	if isSimpleType(fd.Type) {
		protoBufName, ok := parseTag(string(fd.Tag))
		if ok {
			*elements = append(*elements, SubElement{
				Name: protoBufName,
				Type: fd.Type,
			})

			*fields = append(*fields,
				fmt.Sprintf("%v %v", fd.Name, toType(fd.Type)))
		}
		return
	}

	complexTy := fd.Type
	for complexTy.Kind() == reflect.Pointer {
		complexTy = complexTy.Elem()
	}
	if complexTy.Kind() == reflect.Struct {
		for i := 0; i < complexTy.NumField(); i++ {
			appendElement(elements, fields, complexTy.Field(i))
		}
	}
}

func GenStructDesc(target ObjDesc) ReflectSrc {
	ty := reflect.TypeOf(target.Obj)
	if ty.Kind() != reflect.Struct {
		panic("unexpected")
	}
	var fields []string
	var elements []SubElement
	for i := 0; i < ty.NumField(); i++ {
		var fd reflect.StructField = ty.Field(i)
		appendElement(&elements, &fields, fd)
	}

	var tig = TableInitGen{
		TableName: target.Name,
		Elements:  elements,
	}

	var fills = Fills{
		Name:   ty.Name(),
		Fields: fields,
	}
	out := bytes.NewBuffer(nil)
	structBuilder.Execute(out, fills)

	fmt.Fprintf(out, "\n")
	var newItem string
	if target.IsRepeat {
		newItem = fmt.Sprintf("%v []%v", ty.Name(), ty.Name())
	} else {
		newItem = fmt.Sprintf("%v %v", ty.Name(), ty.Name())
	}

	return ReflectSrc{
		Item:     newItem,
		GoStruct: out.String(),
		TableGen: tig.String(),
	}
}

func GenCombinations(fout io.Writer, seqs []string) {
	combBuilder.Execute(fout, seqs)
}
