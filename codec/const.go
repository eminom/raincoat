package codec

const (
	CqmEventOpStart = 9
	CqmEventOpEnd   = 8
)

func IsCqmOpEvent(evt DpfEvent) bool {
	return evt.EngineTypeCode == EngCat_CQM &&
		(evt.Event == CqmEventOpStart || evt.Event == CqmEventOpEnd)
}
