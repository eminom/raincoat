package infoloader

type DtuOpFormatFetcher interface {
	FetchOpIdOpName(string) (string, string)
}

type compatibleFetcher struct{}

func (c *compatibleFetcher) FetchOpIdOpName(text string) (string, string) {
	vs := XSplit(text, 3)
	return vs[0], vs[1]
}

type nuevoModeFetcher struct{}

func (c *nuevoModeFetcher) FetchOpIdOpName(text string) (string, string) {
	vs := XSplit(text, 4)
	return vs[0], vs[2]
}

func newCompatibleOpFetcher() DtuOpFormatFetcher {
	return new(compatibleFetcher)
}

func newNuevoOpFetcher() DtuOpFormatFetcher {
	return new(nuevoModeFetcher)
}
