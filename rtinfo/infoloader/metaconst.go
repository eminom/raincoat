package infoloader

var (
	pkt2opFileSuffixes = []string{"_pkt2op.dumptxt", "_pkt2op.pbdumptxt"}
)

type OpInfoSuffixConf struct {
	suffixName     string
	fetcherCreator func() DtuOpMapLoader
}

var (
	opFileSuffixes = []OpInfoSuffixConf{
		{
			"_dtuop.dumptxt",
			newCompatibleOpLoader,
		},
		{
			"_opmeta.pbdumptxt",
			newNuevoOpLoader,
		},
	}
)

func GetOpFileSuffixes() []OpInfoSuffixConf {
	return opFileSuffixes
}

type DmaInfoSuffixConf struct {
	suffixName     string
	fetcherCreator func() DmaOpFormatFetcher
}

var (
	dmaInfoFileSuffixes = []DmaInfoSuffixConf{
		{
			"_memcpy_meta.dumptxt",
			NewCompatibleDmaInfoLoader,
		},
		{
			"_memcpy.pbdumptxt",
			NewPbDmaInfoLoader,
		},
	}
)

func GetDmaInfoFileSuffix() []DmaInfoSuffixConf {
	return dmaInfoFileSuffixes
}
