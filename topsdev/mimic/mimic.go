package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"git.enflame.cn/hai.bai/dmaster/topsdev/mimic/mimics"
	"git.enflame.cn/hai.bai/dmaster/topsdev/proto/pbdef/topspb"
)

type ProtoDefs struct {
	GoStruct  string
	TableInit string
}

func genX(fout io.Writer) {
	var seqs []string
	for _, target := range []mimics.ObjDesc{
		{"command", topspb.CommandInfo{}, false},
		{"platform", topspb.PlatformInfo{}, true},
		{"version", topspb.VersionInfo{}, false}} {
		entry := mimics.GenStructDesc(target)
		seqs = append(seqs, entry.Item)
		fmt.Fprintf(fout, "%v\n", entry.GoStruct)
		log.Printf("Create init:\n%v\n", entry.TableGen)
	}
	fmt.Fprintf(fout, "\n")
	mimics.GenCombinations(fout, seqs)
	fmt.Fprintf(fout, "\n\n")
}

func main() {
	fout := os.Stdout
	genX(fout)
}
