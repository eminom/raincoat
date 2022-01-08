package linklist

import (
	"bytes"
	"fmt"
	"log"
	"testing"
)

func toString(lnk *Lnk) string {
	buf := bytes.NewBuffer(nil)
	lnk.ConstForEach(func(v interface{}) {
		fmt.Fprintf(buf, "%v,", v.(int))
	})
	log.Printf("this <%v>", buf.String())
	return buf.String()
}

func genMatcher(v int) func(interface{}) bool {
	return func(a interface{}) bool {
		return a.(int) == v
	}
}

func TestLinkList(t *testing.T) {
	lnk := NewLnkHead()

	lnk.AppendAtFront(1)
	lnk.AppendAtFront(2)
	lnk.AppendAtFront(3)
	if toString(lnk) != "3,2,1," {
		t.Fail()
	}
	lnk.Extract(genMatcher(2))
	if toString(lnk) != "3,1," {
		t.Fail()
	}
	lnk.AppendAtTail(4)
	lnk.AppendAtTail(5)
	lnk.AppendAtTail(4)
	if toString(lnk) != "3,1,4,5,4," {
		t.Fail()
	}

	lnk.Extract(genMatcher(1))
	if toString(lnk) != "3,4,5,4," {
		t.Fail()
	}

	for i := 0; i < 10; i++ {
		lnk.Extract(genMatcher(4))
	}
	if toString(lnk) != "3,5," {
		t.Fail()
	}

	for i := 0; i < 100; i++ {
		lnk.Extract(func(interface{}) bool {
			return true
		})
	}
	if toString(lnk) != "" {
		t.Fail()
	}

	for i := 11; i < 16; i++ {
		if i&1 != 0 {
			lnk.AppendAtTail(i)
		} else {
			lnk.AppendAtFront(i)
		}
	}
	if toString(lnk) != "14,12,11,13,15," {
		t.Fail()
	}
}
