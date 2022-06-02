package main

import (
	"os"

	"git.enflame.cn/hai.bai/dmaster/codec"
)

func main() {
	codec.GenDictForDorado(os.Stdout)
}
