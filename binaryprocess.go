package main

import (
	"io"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/sess"
)

func BinaryProcess(chunk []byte,
	out io.Writer,
	decoder *codec.DecodeMaster,
	decodeGr int) {

	sess := sess.NewSession(sess.SessionOpt{
		Debug:        *fDebug,
		Sort:         *fSort,
		DecodeFull:   *fDecodeFull,
		EngineFilter: *fEng,
	})
	sess.DecodeChunk(chunk, decoder, decodeGr)
	sess.PrintItems(out, *fRaw)
}
