package assert

import (
	"fmt"
	"testing"
)

func TestAssert(t *testing.T) {
	(func() {
		defer func() {
			if err := recover(); fmt.Sprintf("%v", err) != "error for test" {
				t.Fail()
			}
		}()
		Assert(false, "error for test")
	})()
}
