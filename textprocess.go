package main

import (
	"os"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/sess"
)

func TextProcess(decoder *codec.DecodeMaster) {
	sess := sess.NewSession(sess.SessionOpt{
		Debug:        *fDebug,
		Sort:         *fSort,
		DecodeFull:   *fDecodeFull,
		EngineFilter: *fEng,
	})
	sess.DecodeFromTextStream(os.Stdin, decoder)
	if *fDump {
		sess.PrintItems(*fRaw)
	}
}
