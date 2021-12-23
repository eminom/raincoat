package sess

import (
	"fmt"
	"os"
)

type ErrorWatcher struct {
	errCount   int
	okCount    int
	printQuota int
}

func (e *ErrorWatcher) TickSuccess() {
	e.okCount++
}

func (e *ErrorWatcher) ReceiveError(vals []uint32, lineno int) {
	e.errCount++
	e.printQuota--
	if e.printQuota >= 0 {
		fmt.Fprintf(os.Stderr, "%d: ", lineno)
		for _, v := range vals {
			fmt.Fprintf(os.Stderr, "%08x,", v)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}
}

func (e *ErrorWatcher) ReceiveErr(err error) {
	e.errCount++
	e.printQuota--
	if e.printQuota >= 0 {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func (e ErrorWatcher) SumUp() {
	if e.errCount > 0 {
		fmt.Fprintf(os.Stderr, "ok count: %v\n", e.okCount)
		fmt.Fprintf(os.Stderr, "error for file decode: %v\n", e.errCount)
	}
}
