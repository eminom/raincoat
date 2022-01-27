package main

import (
	"os"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/sess"
)

func BinaryProcess(inputfile string, decoder *codec.DecodeMaster) {

	chunk, err := os.ReadFile(inputfile)
	if err != nil {
		panic(err)
	}

	sess := sess.NewSession(sess.SessionOpt{
		Debug:        *fDebug,
		Sort:         *fSort,
		DecodeFull:   *fDecodeFull,
		EngineFilter: *fEng,
	})
	sess.DecodeChunk(chunk, decoder, false)
	sess.PrintItems(*fRaw)
}
