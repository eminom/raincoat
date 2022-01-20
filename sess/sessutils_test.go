package sess

import "testing"

func testSuite(t *testing.T, jd JobDivider,
	workCount, itemCount int, expectedWorkCount, expectedSegSize int) {
	work, seg := jd.DetermineWorkThread(workCount, itemCount)
	if work != expectedWorkCount {
		t.Logf("expecting %v, got %v", expectedWorkCount, work)
		t.Fail()
	}
	if seg != expectedSegSize {
		t.Logf("expecting seg %v, got %v", expectedSegSize, seg)
		t.Fail()
	}
}

func TestWorkerCount(t *testing.T) {
	jd := JobDivider{100}
	testSuite(t, jd, 7, 12, 1, 12)
	testSuite(t, jd, 2, 100, 1, 100)
	testSuite(t, jd, 2, 99, 1, 99)
	testSuite(t, jd, 1, 101, 1, 101)
	testSuite(t, jd, 1, 101101101, 1, 101101101)
	testSuite(t, jd, 5, 500, 5, 100)
	testSuite(t, jd, 5, 501, 5, 101)
	testSuite(t, jd, 7, 10000, 7, 1429)
}
