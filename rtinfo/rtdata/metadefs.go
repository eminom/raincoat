package rtdata

import (
	"bufio"
	"bytes"
	"fmt"
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

func decorateName(name string) string {
	vs := strings.Split(name, "_")
	if len(vs) > 0 && vs[len(vs)-1] == "START" {
		vs = vs[:len(vs)-1]
	}
	out := bytes.NewBuffer(nil)
	for i := 0; i < len(vs); i++ {
		p := vs[i]
		if len(p) <= 0 {
			continue
		}
		if out.Len() > 0 {
			out.WriteByte(' ')
		}
		out.WriteByte(p[0])
		for j := 1; j < len(p); j++ {
			q := p[j]
			if q >= 'A' && q <= 'Z' {
				q += ('a' - 'A')
			}
			out.WriteByte(q)
		}
	}
	return out.String()
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
			orgValueStr := vs[2]
			val := parseInt32(orgValueStr)
			if _, ok := dc[val]; ok {
				panic(fmt.Errorf("duplicate key for %v: %v", text, val))
			}
			nameStr := decorateName(vs[0])
			// log.Printf("%v => %v added: %v", orgValueStr, nameStr, val)
			dc[val] = nameStr
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
