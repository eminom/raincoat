package infoloader

var (
	pkt2opFileSuffixes = []string{"_pkt2op.dumptxt", "_pkt2op.pbdumptxt"}
)

type SuffixConf struct {
	suffixName     string
	fetcherCreator func() DtuOpMapLoader
}

var (
	opFileSuffixes = []SuffixConf{
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
