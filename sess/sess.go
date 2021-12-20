package sess

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
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
	Cached       bool
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

func (sess *Session) ProcessMasterText(text string, decoder *codec.DecodeMaster) bool {
	if strings.HasPrefix(text, "0x") || strings.HasPrefix(text, "0X") {
		text = text[2:]
	}
	if len(text) <= 0 {
		fmt.Printf("\n")
		return true
	}
	text = strings.Trim(text, " ")
	val, err := strconv.ParseUint(text, 16, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parse hex: %v\n", err)
		return false
	}
	engineId, engineType, ctx, ok := decoder.GetEngineInfo(uint32(val))
	if !ok {
		fmt.Printf("decode error for 0x%08x\n", val)
		return false
	}
	fmt.Printf("%08x  %v %v %v", val, engineType, engineId, ctx)
	fmt.Printf("\n")
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

func (sess *Session) ProcessFullItem(text string, decoder *codec.DecodeMaster) (bool, error) {
	vs := toItems(strings.Split(text, " "))
	if len(vs) != 4 {
		return true, inputErr
	}
	return sess.ProcessItems(vs, decoder)
}

func (sess *Session) ProcessItems(vs []uint32, decoder *codec.DecodeMaster) (bool, error) {
	item, err := decoder.NewDpfItem(vs)
	if err != nil {
		return true, err
	}
	if sess.sessOpt.Cached {
		var toAdd = true
		if len(sess.sessOpt.EngineFilter) > 0 &&
			!strings.HasPrefix(item.EngineTy, sess.sessOpt.EngineFilter) {
			toAdd = false
		}
		if toAdd {
			sess.AppendItem(item)
		}
	} else {
		fmt.Println(item.ToString())
	}
	return true, nil
}

func (sess *Session) DecodeDpfItem(decoder *codec.DecodeMaster) {
	reader := bufio.NewReader(os.Stdin)
	errorCounter := 0
	for {
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
			shallCont, err := sess.ProcessFullItem(text, decoder)
			if nil != err {
				errorCounter++
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

	if sess.sessOpt.Cached {
		if sess.sessOpt.Sort {
			sort.Sort(codec.DpfItems(sess.items))
		}
		for _, v := range sess.items {
			fmt.Println(v.ToString())
		}
	}

	if errorCounter > 0 {
		fmt.Fprintf(os.Stderr, "error decode %v\n", errorCounter)
	}
}

func (sess *Session) DecodeFromFile(filename string, decoder *codec.DecodeMaster) {
	chunk, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading %v\n", filename)
		os.Exit(1)
	}
	itemSize := len(chunk) / 16 * 16
	errCount := 0
	for i := 0; i < itemSize; i += 16 {
		var u32vals = [4]uint32{
			binary.LittleEndian.Uint32(chunk[i:]),
			binary.LittleEndian.Uint32(chunk[i+4:]),
			binary.LittleEndian.Uint32(chunk[i+8:]),
			binary.LittleEndian.Uint32(chunk[i+12:]),
		}
		_, err := sess.ProcessItems(u32vals[:], decoder)
		if err != nil {
			errCount++
		}
	}
	if sess.sessOpt.Cached {
		if sess.sessOpt.Sort {
			sort.Sort(codec.DpfItems(sess.items))
		}
		for _, v := range sess.items {
			fmt.Println(v.ToString())
		}
	}
	if errCount > 0 {
		fmt.Fprintf(os.Stderr, "error for file decode: %v\n", errCount)
	}
}
