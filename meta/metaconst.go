package meta

var (
	pkt2opFileSuffixes = []string{"_pkt2op.dumptxt", "_pkt2op.pbdumptxt"}
)

type SuffixConf struct {
	suffixName     string
	fetcherCreator func() FormatFetcher
}

var (
	opFileSuffixes = []SuffixConf{
		{
			"_dtuop.dumptxt",
			newCompatibleOpFetcher,
		},
		{
			"_opmeta.pbdumptxt",
			newNuevoOpFetcher,
		},
	}
)
