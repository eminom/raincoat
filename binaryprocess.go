package main

import (
	"io"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/meta/dtuarch"
	"git.enflame.cn/hai.bai/dmaster/sess"
	"git.enflame.cn/hai.bai/dmaster/vgrule"
)

func BinaryProcess(chunk []byte,
	out io.Writer,
	decoder *codec.DecodeMaster,
	decodeGr int,
	engineOrder vgrule.EngineOrder) {

	sess := sess.NewSession(sess.SessionOpt{
		Debug:        *fDebug,
		Sort:         *fSort,
		DecodeFull:   *fDecodeFull,
		EngineFilter: *fEng,
	})
	sess.DecodeChunk(chunk, decoder, decodeGr)
	sess.PrintItems(out, *fRaw)

	// Do pg statistics automatically for Dorado
	if decoder.Arch == dtuarch.DoradoNameTrait {
		sess.CalcStat(engineOrder)
	}
}
