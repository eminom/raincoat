package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/codec"
)

var (
	inputErr     = errors.New("input format error")
	decodeErr    = errors.New("decode error")
	malItemError = errors.New("malformed item error")
)

type DpfItem struct {
	RawVale [4]uint32

	Flag     int
	PacketID int
	Event    int
	Context  int
	Payload  int
	Cycle    uint64

	EngineTy    string
	EngineIndex int
}

func (d DpfItem) ToString() string {
	if d.Flag == 0 {
		return fmt.Sprintf("%-10v %-2v %-2v event=%-4v pid=%v ts=%08x",
			d.EngineTy, d.EngineIndex, d.Context, d.Event, d.PacketID, d.Cycle)
	}
	return fmt.Sprintf("%-6v %-2v event=%v payload=%v ts=%08x",
		d.EngineTy, d.EngineIndex, d.Event, d.Payload, d.Cycle)
}

func NewDpfItem(vals []uint32, decoder *codec.DecodeMaster) (DpfItem, error) {
	if len(vals) != 4 {
		panic(malItemError)
	}
	ts := uint64(vals[2]) + uint64(vals[3])<<32
	if vals[0]&1 == 0 {
		// uint32_t flag_ : 1;  // should be always 0
		// uint32_t event_ : 8;
		// uint32_t packet_id_ : 23;
		event := (vals[0] >> 1) & 0xFF
		packet_id := (vals[0] >> 9)
		engIdx, ctx, engTy, ok := decoder.GetEngineInfo(vals[1])
		if !ok {
			return DpfItem{}, decodeErr
		}

		return DpfItem{
			Flag:        0,
			PacketID:    int(packet_id),
			Event:       int(event),
			Context:     ctx,
			EngineTy:    engTy,
			EngineIndex: engIdx,
			Cycle:       ts,
		}, nil

	}
	// uint32_t flag_ : 1;  // should be always 1
	// uint32_t event_ : 7;
	// uint32_t payload_ : 24;
	event := (vals[0] >> 1) & 0x7F
	payload := (vals[0] >> 8)
	engineIdx, engTy, ok := decoder.GetEngineInfoV2(vals[1])
	if !ok {
		return DpfItem{}, decodeErr
	}
	return DpfItem{
		Flag:        1,
		Event:       int(event),
		Payload:     int(payload),
		EngineTy:    engTy,
		EngineIndex: engineIdx,
	}, nil
}

var (
	fDebug      = flag.Bool("debug", false, "for debug output")
	fArch       = flag.String("arch", "dorado", "hardware arch")
	fDecodeFull = flag.Bool("decodefull", false, "decode all line")
	fSort       = flag.Bool("sort", false, "sort by order")
	fCache      = flag.Bool("cached", false, "cache result")
	fEng        = flag.String("eng", "", "engine to filter in")
)

func init() {
	flag.Parse()

	switch *fArch {
	case "pavo":
	case "dorado":
	default:
		fmt.Fprintf(os.Stderr, "unknown arch %v\n", *fArch)
		os.Exit(0)
	}
}

type Session struct {
	items []DpfItem
}

type DpfItems []DpfItem

func (d DpfItems) Len() int {
	return len(d)
}

func (d DpfItems) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d DpfItems) Less(i, j int) bool {
	lhs, rhs := d[i], d[j]
	if lhs.PacketID != rhs.PacketID {
		return lhs.PacketID < rhs.PacketID
	}
	if lhs.Event != rhs.Event {
		return lhs.Event < rhs.Event
	}
	return lhs.Cycle < rhs.Cycle
}

func NewSession() *Session {
	return &Session{}
}

func (sess *Session) AppendItem(newItem DpfItem) {
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
	item, err := NewDpfItem(vs, decoder)
	if err != nil {
		return true, err
	}
	if *fCache {
		var toAdd = true
		if len(*fEng) > 0 && !strings.HasPrefix(item.EngineTy, *fEng) {
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

func DecodeDpfItem(sess *Session, decoder *codec.DecodeMaster) {
	reader := bufio.NewReader(os.Stdin)
	errorCounter := 0
	for {
		// fmt.Print("-> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			if *fDebug {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			break
		}
		text = strings.TrimSuffix(text, "\n")
		if *fDecodeFull {
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

	if *fCache {
		if *fSort {
			sort.Sort(DpfItems(sess.items))
		}
		for _, v := range sess.items {
			fmt.Println(v.ToString())
		}
	}

	if errorCounter > 0 {
		fmt.Fprintf(os.Stderr, "error decode %v\n", errorCounter)
	}
}

func DecodeFromFile(sess *Session, filename string, decoder *codec.DecodeMaster) {
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
	if *fCache {
		if *fSort {
			sort.Sort(DpfItems(sess.items))
		}
		for _, v := range sess.items {
			fmt.Println(v.ToString())
		}
	}
	if errCount > 0 {
		fmt.Fprintf(os.Stderr, "error for file decode: %v\n", errCount)
	}
}

func main() {

	sess := NewSession()
	decoder := codec.NewDecodeMaster(*fArch)
	if len(flag.Args()) > 0 {
		filename := flag.Args()[0]
		DecodeFromFile(sess, filename, decoder)
	} else {
		DecodeDpfItem(sess, decoder)
	}
}
