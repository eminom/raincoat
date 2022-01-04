package rtdata

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/assert"
)

func parseInt32(t string) int {
	t = strings.TrimRight(t, ",")
	base := 10
	if strings.HasPrefix(t, "0x") || strings.HasPrefix(t, "0X") {
		t = t[2:]
		base = 16
	}
	if len(t) <= 0 {
		return 0
	}
	v, err := strconv.ParseInt(t, base, 32)
	assert.Assert(err == nil, "Must not be nil for %v", t)
	return int(v)
}

func extractEventNameMap(str string) map[int]string {
	dc := make(map[int]string)

	buffer := bytes.NewBufferString(str)
	scanner := bufio.NewScanner(buffer)
	for {
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		vs := strings.Fields(text)
		if len(vs) == 3 && vs[1] == "=" {
			val := parseInt32(vs[2])
			if _, ok := dc[val]; ok {
				panic(fmt.Errorf("duplicate key for %v: %v", text, val))
			}
			log.Printf("%v => %v added: %v", vs[2], vs[0], val)
			dc[val] = vs[0]
		}
	}
	return dc
}

/* Things I need to pay attention to
Command Packet
Debug Packet Op
Debug Packet Step
Executable
CQM Executable Launch
*/

type EventNameKeeper struct {
	nameDict map[int]string
}

func NewEventNameKeeper(str string) EventNameKeeper {
	return EventNameKeeper{extractEventNameMap(str)}
}
