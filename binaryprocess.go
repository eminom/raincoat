package main

import (
	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/sess"
)

func BinaryProcess(chunk []byte, decoder *codec.DecodeMaster, decodeGr int) {

	sess := sess.NewSession(sess.SessionOpt{
		Debug:        *fDebug,
		Sort:         *fSort,
		DecodeFull:   *fDecodeFull,
		EngineFilter: *fEng,
	})
	sess.DecodeChunk(chunk, decoder, decodeGr)
	sess.PrintItems(*fRaw)
}
