package sess

const (
	MinPackItemCount = 100
)

func DetermineWorkThread(workerItemCount, totalItemCount int) (int, int) {
	segmentSize := (totalItemCount + workerItemCount - 1) / workerItemCount
	if segmentSize < MinPackItemCount {
		segmentSize = MinPackItemCount
	}
	if totalItemCount < segmentSize {
		segmentSize = totalItemCount
	}
	workerItemCount = (totalItemCount + segmentSize - 1) / segmentSize
	return workerItemCount, segmentSize
}
