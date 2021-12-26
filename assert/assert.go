package assert

import "fmt"

func Assert(cond bool, msg string) {
	if !cond {
		panic(fmt.Errorf("%v", msg))
	}
}
