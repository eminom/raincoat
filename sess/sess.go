package sess

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/codec"
)

var (
	inputErr = errors.New("input error")
)

type SessionOpt struct {
	EngineFilter string
	Debug        bool
	DecodeFull   bool
	Sort         bool
}

type Session struct {
	items   []codec.DpfItem
	sessOpt SessionOpt
}

func NewSession(sessOpt SessionOpt) *Session {
	return &Session{}
}

func (sess *Session) AppendItem(newItem codec.DpfItem) {
	sess.items = append(sess.items, newItem)
}

// Process master text, no cache
func (sess *Session) ProcessMasterText(text string, decoder *codec.DecodeMaster) bool {
	if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
		text = text[2:]
	}
	if len(text) <= 0 {
		return true
	}
	text = strings.Trim(text, " ")
	val, err := strconv.ParseUint(text, 16, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parse hex: %v\n", err)
		return false
	}
	engineId, engineUniqIdx, ctx, ok := decoder.GetEngineInfo(uint32(val))
	if !ok {
		fmt.Fprintf(os.Stderr, "decode error for 0x%08x\n", val)
		return false
	}
	if ctx >= 16 {
		panic(fmt.Errorf("assertion error: ctx = %v, val = 0x%x", ctx, val))
	}
	engineTypeStr := decoder.EngUniqueIndexToTypeName(engineUniqIdx)
	fmt.Printf("%08x  %v %v %v\n", val, engineTypeStr, engineId, ctx)
	return true
}

func toItems(vs []string) []uint32 {
	var arr []uint32
	for _, s := range vs {
		s = strings.Trim(s, " ")
		if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
			s = s[2:]
		}
		if len(s) <= 0 {
			// fmt.Fprintf(os.Stderr, "not a valid number in hex format: '%v'\n", s)
			continue
		}
		val, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parse hex: %v\n", err)
			return nil
		}
		arr = append(arr, uint32(val))
	}
	return arr
}

func (sess *Session) ProcessFullItem(
	text string, offsetIdx int,
	decoder *codec.DecodeMaster,
) (bool, error) {
	vs := toItems(strings.Split(text, " "))
	if len(vs) != 4 {
		return true, inputErr
	}
	return sess.ProcessItems(vs, offsetIdx, decoder)
}

// Process one item, always append
func (sess *Session) ProcessItems(vs []uint32,
	offsetIdx int,
	decoder *codec.DecodeMaster,
) (bool, error) {
	item, err := decoder.NewDpfEvent(vs, offsetIdx)
	if err != nil {
		return true, err
	}
	var toAdd = true
	if len(sess.sessOpt.EngineFilter) > 0 &&
		!strings.HasPrefix(item.EngineTy, sess.sessOpt.EngineFilter) {
		toAdd = false
	}
	if toAdd {
		sess.AppendItem(item)
	}
	return true, nil
}

func (sess *Session) DecodeFromTextStream(
	inHandle *os.File,
	decoder *codec.DecodeMaster,
) {
	reader := bufio.NewReader(inHandle)
	var errWatcher = ErrorWatcher{printQuota: 10}
	for lineno := 0; ; lineno++ {
		// fmt.Print("-> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			if sess.sessOpt.Debug {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			break
		}
		text = strings.TrimSuffix(text, "\n")
		if sess.sessOpt.DecodeFull {
			shallCont, err := sess.ProcessFullItem(text, lineno, decoder)
			if nil != err {
				errWatcher.ReceiveErr(err)
			} else {
				errWatcher.TickSuccess()
			}
			if !shallCont {
				break
			}
		} else {
			if !sess.ProcessMasterText(text, decoder) {
				break
			}
		}
	}

	if sess.sessOpt.Sort {
		sort.Sort(codec.DpfItems(sess.items))
	}

	errWatcher.SumUp()
}

func (sess *Session) DecodeFromFile(filename string,
	decoder *codec.DecodeMaster,
) {
	// realpath, e2 := os.Readlink(filename)
	// if nil == e2 {
	// 	filename = realpath
	// }
	chunk, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %v\n", filename)
		os.Exit(1)
	}
	itemSize := len(chunk) / 16 * 16
	var errWatcher = ErrorWatcher{printQuota: 10}
	for i := 0; i < itemSize; i += 16 {
		offsetIdx := i >> 4
		var u32vals = [4]uint32{
			binary.LittleEndian.Uint32(chunk[i:]),
			binary.LittleEndian.Uint32(chunk[i+4:]),
			binary.LittleEndian.Uint32(chunk[i+8:]),
			binary.LittleEndian.Uint32(chunk[i+12:]),
		}
		_, err := sess.ProcessItems(u32vals[:], offsetIdx, decoder)
		if err != nil {
			errWatcher.ReceiveError(u32vals[:], offsetIdx)
		} else {
			errWatcher.TickSuccess()
		}
	}
	if sess.sessOpt.Sort {
		sort.Sort(codec.DpfItems(sess.items))
	}
	errWatcher.SumUp()
}

func (sess Session) PrintItems(printRaw bool) {
	if printRaw {
		for _, v := range sess.items {
			fmt.Printf("%-50v : %v\n", v.ToString(), v.RawRepr())
		}
	} else {
		for _, v := range sess.items {
			fmt.Println(v.ToString())
		}
	}
}
