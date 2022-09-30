package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/topsdev"
)

func init() {
	flag.Parse()
}

func genOutName(org string) string {
	const suffix = ".topspti.data"
	if strings.HasSuffix(org, ".bin") {
		return org[:len(org)-4] + suffix
	}
	return org + suffix
}

func main() {
	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "not enough paramter to process")
	}

	rawdpf := args[0]
	execs := args[1:]

	var profs []topsdev.ProfileSecElement
	for _, exec := range execs {
		profs = append(profs, topsdev.LoadProfileSection(exec)...)
	}

	filerawchunk, err := os.ReadFile(rawdpf)
	if err != nil {
		panic(err)
	}

	buffer := bytes.NewBuffer(nil)

	// header
	enc := topsdev.CreateProfHeaderEnc()
	buffer.Write(enc.EncodeBuffer())

	// body
	soe := topsdev.NewSerailObjEnc()
	soe.EncodeBody(filerawchunk, profs)
	buffer.Write(soe.Bytes())

	os.WriteFile(genOutName(rawdpf), buffer.Bytes(), 0444)

}
