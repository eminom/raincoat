package topsdev

/*
#include <stdint.h>

// Insetion from developer manually
typedef uint64_t u64;

// Copy from source definitions
struct SectionHeader {
  u64 sh_type;
  u64 sh_offset;
  u64 sh_size;
};

// Manually
typedef struct SectionHeader SectionHeader;

struct ExecutableHeader {
  u64 tag;
  u64 crc64;
  u64 version;
  u64 device;
  u64 shnum;
  u64 reserve[16];
  // reserve[0]: As a flag of prefetch of sip code
  // reserve[1]: As executable uuid
  SectionHeader shlist[0];
};

typedef struct ExecutableHeader ExecutableHeader;


size_t SizeOfExecHeader() {
	return sizeof(ExecutableHeader);
}

size_t SizeOfShSectionHeader() {
	return sizeof(SectionHeader);
}

*/
import "C"
import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"unsafe"

	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
)

// Remove :u64
const (
	SHT_NULL_ = 0 + iota
	SHT_HEADER
	SHT_RESOURCE
	SHT_INPUT
	SHT_OUT
	SHT_CONSTANT
	SHT_SIPCODE
	SHT_SIP_FUNC_PARAM
	SHT_SIP_FUNC_PARAM_UPDATE
	SHT_STREAM
	SHT_PACKET
	SHT_PACKET_UPDATE
	SHT_CLUSTER
	SHT_JITBINARY
	SHT_CPU_FUNC_PARAM
	SHT_CPU_FUNC_DATA
	SHT_CPU_FUNCTION
	SHT_PROFILE //17
	SHT_HOST_CONST
	SHT_TASK
	SHT_SIP_CALLBACK //20
	SHT_TARGET_RESOURCE
	SHT_TENSOR_TABLE
	SHT_KERN_ASSERT_INFO //23
	SHT_RAND_STATE
	SHT_USER4
	SHT_USER5
	SHT_USER6
	SHT_USER7
	SHT_USER8
	SHT_USER9
)

func SizeOfExecHeader() int {
	return int(C.SizeOfExecHeader())
}

func SizeOfShSectionHeader() int {
	return int(C.SizeOfShSectionHeader())
}

type ExecRawData struct {
	ExecUuid  uint64
	DataChunk []byte
	DataType  int
}

func LoadSectionsFromExec(filename string) []ExecRawData {
	chunk, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("error open %v: %v", filename, err)
	}

	var rawVec []ExecRawData
	uVal := reflect.ValueOf(chunk).Pointer()
	sec := *(*C.ExecutableHeader)(unsafe.Pointer(uVal))

	sectionStart := chunk[SizeOfExecHeader():]
	for i := 0; i < int(sec.shnum); i++ {
		piece := sectionStart[i*SizeOfShSectionHeader() : (i+1)*SizeOfShSectionHeader()]
		uVal1 := reflect.ValueOf(piece).Pointer()
		shSec := *(*C.SectionHeader)(unsafe.Pointer(uVal1))
		if shSec.sh_type == SHT_PROFILE {
			// log.Printf("%v(SHT_PROFILE) detected", SHT_PROFILE)
			offset := int(shSec.sh_offset)
			chunkSize := int(shSec.sh_size)
			sli := chunk[offset : offset+chunkSize]
			rawVec = append(rawVec, ExecRawData{
				DataChunk: bytes.Repeat(sli, 1),
				ExecUuid:  uint64(sec.crc64),
				DataType:  SHT_PROFILE,
			})
		} else if shSec.sh_type == SHT_KERN_ASSERT_INFO {
			// log.Printf("%v(SHT_KERN_ASSERT_INFO) detected", SHT_KERN_ASSERT_INFO)
			offset := int(shSec.sh_offset)
			chunkSize := int(shSec.sh_size)
			sli := chunk[offset : offset+chunkSize]
			rawVec = append(rawVec, ExecRawData{
				DataChunk: bytes.Repeat(sli, 1),
				ExecUuid:  uint64(sec.crc64),
				DataType:  SHT_KERN_ASSERT_INFO,
			})
		} else {
			fmt.Fprintf(io.Discard, "sec type %v\n", shSec)
		}
	}
	return rawVec
}

func DumpSectionsFromExecutable(filename string, checkFormatOnly bool) {
	chunkVec := LoadSectionsFromExec(filename)
	for _, execRaw := range chunkVec {
		switch execRaw.DataType {
		case SHT_PROFILE:

			execScope, fc := ParseProfileSectionFromData(execRaw.DataChunk,
				execRaw.ExecUuid,
				io.Discard,
			)

			if checkFormatOnly {
				fmt.Printf("ProfSec for %v(0x%016x) is type-%v\n",
					filename, execRaw.ExecUuid, fc)
				break
			}
			if execScope == nil {
				break
			}
			execScope.DumpDtuOpToFile()
			execScope.DumpDmaToFile()
			execScope.DumpPktOpMapToFile()
		case SHT_KERN_ASSERT_INFO:
			NewAssertRtDict(execRaw).DumpAssertInfo()
		}
	}
}

func LoadExecScopeFromExec(filename string) []*metadata.ExecScope {
	var colls []*metadata.ExecScope
	chunkVec := LoadSectionsFromExec(filename)
	for _, execRaw := range chunkVec {
		switch execRaw.DataType {
		case SHT_PROFILE:

			execScope, _ := ParseProfileSectionFromData(execRaw.DataChunk,
				execRaw.ExecUuid,
				io.Discard,
			)
			if execScope != nil {
				colls = append(colls, execScope)
			}
		}
	}
	return colls
}

type ProfileSecElement struct {
	execUuid   uint64
	profSecRaw []byte
}

func LoadProfileSection(filename string) []ProfileSecElement {
	var profs []ProfileSecElement
	chunkVec := LoadSectionsFromExec(filename)
	for _, execRaw := range chunkVec {
		switch execRaw.DataType {
		case SHT_PROFILE:

			profs = append(profs, ProfileSecElement{
				execUuid:   execRaw.ExecUuid,
				profSecRaw: execRaw.DataChunk,
			})
		}
	}
	return profs
}
