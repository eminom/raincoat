package topsdev

/*
#include <stdint.h>

typedef struct KernelAssertInfoSec {
	uint64_t size;
	uint8_t data[0];
}KernelAssertInfoSec;
*/
import "C"
import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"reflect"
	"unsafe"
)

type AssertInfo struct {
	CauseId int
	SrcFile string
	LineNo  int
	UserMsg string
}

type AssertRtDict struct {
	Info     map[int][]AssertInfo
	ExecUuid uint64
}

func NewAssertRtDict(rawData ExecRawData) AssertRtDict {

	if rawData.DataType != SHT_KERN_ASSERT_INFO {
		panic("unexpected raw chunk")
	}
	chunk := rawData.DataChunk

	slice := chunk
	uVal := reflect.ValueOf(slice).Pointer()
	assertHead := *(*C.KernelAssertInfoSec)(unsafe.Pointer(uVal))
	totSize := int(assertHead.size)
	fmt.Printf("assertion total size is %v\n", totSize)
	fmt.Printf("chunk size is %v\n", len(chunk))

	// skip size
	chunk = chunk[8:]

	order := binary.LittleEndian

	offset := 0
	decodeUint16 := func() int {
		rv := int(order.Uint16(chunk[offset:]))
		offset += 2
		return rv
	}
	decodeUint32 := func() int {
		rv := int(order.Uint32(chunk[offset:]))
		offset += 4
		return rv
	}
	decodeString := func() string {
		length := decodeUint16()
		rv := string(chunk[offset : offset+length])
		offset += length + 1 // skip 0
		return rv
	}

	dc := make(map[int][]AssertInfo)
	clauseCnt := decodeUint32()
	for i := 0; i < clauseCnt; i++ {
		causeId := decodeUint32()
		srcFile := decodeString()
		lineNo := decodeUint32()
		userMsg := decodeString()

		if _, ok := dc[causeId]; ok {
			log.Printf("duplicate cause id: %v\n", causeId)
		}
		dc[causeId] = append(dc[causeId], AssertInfo{
			CauseId: causeId,
			SrcFile: srcFile,
			LineNo:  lineNo,
			UserMsg: userMsg,
		})
	}
	return AssertRtDict{
		Info:     dc,
		ExecUuid: rawData.ExecUuid,
	}
}

func (ard AssertRtDict) DumpAssertInfo() {
	outFile := fmt.Sprintf("0x%016x_assert.pbdumptxt", ard.ExecUuid)

	out, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	for causeId, itemv := range ard.Info {
		for _, item := range itemv {
			fmt.Fprintf(out, "%v \"%v,%v,%v\"\n", causeId,
				item.UserMsg, item.SrcFile, item.LineNo)
		}
	}
}
