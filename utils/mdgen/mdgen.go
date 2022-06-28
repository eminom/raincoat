package main

import (
	"flag"
	"os"

	"git.enflame.cn/hai.bai/dmaster/codec"
)

var (
	fKmd = flag.Bool("kmd", false, "gen affinity for kmd")
)

func init() {
	flag.Parse()
}

func main() {

	if *fKmd {
		codec.GenAffinityMapForDorado(os.Stdout)
		return
	}

	codec.GenDictForDorado(os.Stdout)
}
