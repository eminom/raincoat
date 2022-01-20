package sess

type JobDivider struct {
	minItemCount int
}

func (jd JobDivider) DetermineWorkThread(workerItemCount, totalItemCount int) (int, int) {
	segmentSize := (totalItemCount + workerItemCount - 1) / workerItemCount
	if segmentSize < jd.minItemCount {
		segmentSize = jd.minItemCount
	}
	if totalItemCount < segmentSize {
		segmentSize = totalItemCount
	}
	workerItemCount = (totalItemCount + segmentSize - 1) / segmentSize
	return workerItemCount, segmentSize
}

func DefaultJobDivider() JobDivider {
	return JobDivider{10000}
}
