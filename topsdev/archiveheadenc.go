package topsdev

import "bytes"

type ProfileHeaderEnc struct {
	ProfHeader
}

func CreateProfHeaderEnc() ProfileHeaderEnc {
	hd := ProfileHeaderEnc{
		ProfHeader: ProfHeader{
			Magic:  genProfHeaderMagic(),
			Tag:    genProfHeadTag(),
			BNum:   genProfHeadBNum(),
			Sha256: genProfHeadSha256(),
			MD5:    genProfHeadMd5(),
			Reserv: genProfHeadReserved(),
		},
	}
	return hd
}

func (ph ProfileHeaderEnc) EncodeBuffer() []byte {
	buffer := bytes.NewBuffer(nil)
	buffer.Write(ph.Magic)
	buffer.Write(ph.Tag)
	buffer.Write(ph.BNum)
	buffer.Write(ph.Sha256)
	buffer.Write(ph.MD5)
	buffer.Write(ph.Reserv)
	return buffer.Bytes()
}
