package main

import (
	"fmt"
	"html/template"
	"os"
	"strconv"

	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

const pythonDebugOpTmpl = `
# Automatically generated

def LoadOps():
  opVec = [ \
  {{range $_, $Item :=.}} [ {{ExtractOpId $Item}}, "{{ExtractOpName $Item}}", {{$Item.Start.Cycle}}, {{$Item.End.Cycle}}, ],
  {{end}}
  ]
  return opVec
`

func DumpOpsToPythonDebugCode(dtuOps []rtdata.OpActivity) {
	fout, err := os.Create("oplist1.py")
	if err != nil {
		panic(err)
	}
	defer fout.Close()

	srcTmpl := template.Must(
		template.New("python-debug-op").Funcs(template.FuncMap{
			"ExtractOpName": func(dtuOp rtdata.OpActivity) string {
				return dtuOp.GetOp().OpName
			},
			"ExtractOpId": func(dtuOp rtdata.OpActivity) string {
				return strconv.Itoa(dtuOp.GetOp().OpId)
			},
			"html": func(str string) string {
				return str
			},
		}).Parse(pythonDebugOpTmpl))
	err = srcTmpl.Execute(fout, dtuOps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error dump dtu op to python debug code: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "# done pythondebug code\n")
	}
}
