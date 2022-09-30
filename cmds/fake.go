package main

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
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

var (
	pidPat = regexp.MustCompile(`pid_(\d+)`)
)

func extractProcessIdFromName(name string) int {
	sl := pidPat.FindStringSubmatchIndex(name)
	if len(sl) >= 4 {
		if v, err := strconv.Atoi(name[sl[2]:sl[3]]); err == nil {
			return v
		}
	}
	return rand.Intn(100000)
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
	profOpt := topsdev.ProfileOpt{
		ProfileSections: profs,
		ProcessId:       extractProcessIdFromName(rawdpf),
	}
	soe := topsdev.NewSerailObjEnc()
	soe.EncodeBody(filerawchunk, profOpt)
	buffer.Write(soe.Bytes())

	os.WriteFile(genOutName(rawdpf), buffer.Bytes(), 0444)

}
