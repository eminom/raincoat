package main

import (
	"flag"
	"os"

	"git.enflame.cn/hai.bai/dmaster/codec"
)

var (
	fEng = flag.Bool("eng", false, "engine show")
	fKmd = flag.Bool("kmd", false, "gen affinity for kmd")
)

func init() {
	flag.Parse()
}

func main() {

	if *fEng {
		codec.GenEngineTypeMapForDorado(os.Stdout)
		return
	}

	if *fKmd {
		codec.GenAffinityMapForDorado(os.Stdout)
		return
	}

	codec.GenDictForDorado(os.Stdout)
}
