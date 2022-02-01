package topsdev

/*
#include <stdint.h>
// Make sure all these structures' sizes are multiple of 8 bytes
typedef struct ProfileSection {
  uint64_t size;
  uint64_t flag;
  uint64_t reserved_0;
  uint64_t reserved_1;
  uint64_t reserved_2;
  uint64_t map_num;
  uint64_t thunk_num;
  uint64_t hlo_module_num;
  uint64_t memcpy_num;
  uint64_t string_id_num;
  uint64_t str_len;
  uint8_t data[0];
} ProfileSection;


typedef struct Pkt2OpSec {
  int pkt;
  int op;
} Pkt2OpSec;


// typedef enum ExtraMetaType : uint8_t {
//   EXTRA_META_TYPE_NONE,
//   EXTRA_META_TYPE_LEO,
//   EXTRA_META_TYPE_MAX
// } ExtraMetaType;

typedef uint8_t ExtraMetaType;

typedef struct ExtraMetaLeo {
  int32_t queue_id;
  int32_t slave_packet_id;
} ExtraMetaLeo;

typedef union ExtraMeta { ExtraMetaLeo leo; } ExtraMeta;

typedef struct PbMemcpySec {
  int32_t pkt;

  int32_t op_id;
  int32_t engine_type;
  int32_t engine_id;  // engine id (CDMAX/SIP id)
  int32_t cluster_id;
  int32_t context_id;

  ExtraMetaType extra_meta_type;
  ExtraMeta extra_meta;

  int64_t src_addr;
  int64_t dst_addr;
  int32_t src_size;
  int32_t direction;
  int32_t tiling_mode;

  uint32_t args_count;
  uint32_t args;
} PbMemcpySec;


typedef struct OpMetaSec {
  uint32_t id;
  uint32_t name;
  uint32_t kind;
  uint32_t fusion_kind;
  uint32_t meta;
  int module_id;
  uint32_t module_name;
  uint32_t args_count;
  uint32_t args;
} OpMetaSec;


typedef struct ModuleSec {
  int id;
  uint32_t name;
  uint32_t size;
  uint32_t data;
} ModuleSec;

typedef struct StringIdSec {
  uint32_t str;
  int32_t string_id;
} StringIdSec;

typedef struct StringSec {
  uint64_t size;
  char str[0];
} StringSec;

size_t GetProfileSectionSize() {
	return sizeof(ProfileSection);
}
size_t GetStringSecSize() {
	return sizeof(StringSec);
}
size_t GetPkt2OpSecSize() {
	return sizeof(Pkt2OpSec);
}
size_t GetPbMemcpySecSize() {
	return sizeof(PbMemcpySec);
}
size_t GetOpMetaSecSize() {
	return sizeof(OpMetaSec);
}
size_t GetModuleSecSize() {
	return sizeof(ModuleSec);
}
size_t GetStringIdSecSize() {
	return sizeof(StringIdSec);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"unsafe"

	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/topsdev/proto/pbdef/topspb"
)

var (
	errProfileSectionSizeMismatched = errors.New("ProfileSection-size-mismatched")
)

func GetProfileSectionSize() uintptr {
	return uintptr(C.GetProfileSectionSize())
}

func GetOpMetaSecSize() uintptr {
	return uintptr(C.GetOpMetaSecSize())
}

func GetPkt2OpSecSize() uintptr {
	return uintptr(C.GetPkt2OpSecSize())
}

func GetStringSecSize() uintptr {
	return uintptr(C.GetStringSecSize())
}

func GetModuleSecSize() uintptr {
	return uintptr(C.GetModuleSecSize())
}

func GetPbMemcpySecSize() uintptr {
	return uintptr(C.GetPbMemcpySecSize())
}

func GetStringIdSecSize() uintptr {
	return uintptr(C.GetStringIdSecSize())
}

func dumpProfSec(sec C.ProfileSection) {
	fmt.Printf("flag: %v\n", sec.flag)
	fmt.Printf("reserved0: %x\n", sec.reserved_0)
	fmt.Printf("reserved0: %x\n", sec.reserved_1)
	fmt.Printf("reserved0: %x\n", sec.reserved_2)
	fmt.Printf("packet-id to op-id count: %v\n", sec.map_num)
	fmt.Printf("op-items: %v\n", sec.thunk_num)
	fmt.Printf("memcpy count: %v\n", sec.memcpy_num)
	fmt.Printf("string length: %v\n", sec.str_len)
}

type ProfileSecPipBoy struct {
	rawData    []byte
	sec        C.ProfileSection
	stringPool []byte
}

func NewProfileSecPipBoy(rawData []byte) ProfileSecPipBoy {
	uVal := reflect.ValueOf(rawData).Pointer()
	sec := *(*C.ProfileSection)(unsafe.Pointer(uVal))
	dumpProfSec(sec)
	rv := ProfileSecPipBoy{
		sec:     sec,
		rawData: rawData, //bytes.Repeat(rawData, 1),
	}
	rv.initStringPool()
	return rv
}

func (ps *ProfileSecPipBoy) initStringPool() {
	if ps.sec.str_len > 0 {
		stringPool := ps.rawData[ps.HeaderSize()+
			ps.pkt2OpSecSize()+
			ps.opMetaSecSize()+
			ps.moduleSecSize()+
			ps.memcpySecSize()+
			ps.stringIdSecSize()+
			int(GetStringSecSize()):]
		ps.stringPool = stringPool
	}
}

func (ps ProfileSecPipBoy) Pkt2OpCount() int {
	return int(ps.sec.map_num)
}

func (ps ProfileSecPipBoy) OpMetaCount() int {
	return int(ps.sec.thunk_num)
}

func (ps ProfileSecPipBoy) ModuleCount() int {
	return int(ps.sec.hlo_module_num)
}

func (ps ProfileSecPipBoy) MemcpyCount() int {
	return int(ps.sec.memcpy_num)
}

func (ps ProfileSecPipBoy) pkt2OpSecSize() int {
	return int(ps.sec.map_num) * int(GetPkt2OpSecSize())
}

func (ps ProfileSecPipBoy) opMetaSecSize() int {
	return int(ps.sec.thunk_num) * int(GetOpMetaSecSize())
}

func (ps ProfileSecPipBoy) moduleSecSize() int {
	return int(ps.sec.hlo_module_num) * int(GetModuleSecSize())
}

func (ps ProfileSecPipBoy) memcpySecSize() int {
	return int(ps.sec.memcpy_num) * int(GetPbMemcpySecSize())
}

func (ps ProfileSecPipBoy) stringIdSecSize() int {
	return int(ps.sec.string_id_num) * int(GetStringIdSecSize())
}

func (ps ProfileSecPipBoy) HeaderSize() int {
	return int(GetProfileSectionSize())
}

func (ps ProfileSecPipBoy) ExtractStringAt(offset int) string {
	i := offset
	for i < len(ps.stringPool) && ps.stringPool[i] != 0 {
		i++
	}
	return string(ps.stringPool[offset:i])
}

func (ps ProfileSecPipBoy) ExtractArgsAt(offset int, args int) map[string]string {
	dc := make(map[string]string)
	start := offset

	extractOne := func() string {
		for start < len(ps.stringPool) && ps.stringPool[start] == 0 {
			start++
		}
		i := start
		for i < len(ps.stringPool) && ps.stringPool[i] != 0 {
			i++
		}
		oneStr := string(ps.stringPool[start:i])
		start = i
		return oneStr
	}
	for j := 0; j < args; j++ {
		key := extractOne()
		ty := extractOne()
		_ = ty
		val := extractOne()
		dc[key] = val
	}
	return dc

}

func (ps ProfileSecPipBoy) verifySize() bool {
	totSize := ps.HeaderSize() + ps.pkt2OpSecSize() +
		ps.opMetaSecSize() + ps.moduleSecSize() + ps.memcpySecSize()
	if ps.sec.str_len > 0 {
		uVal := reflect.ValueOf(ps.rawData[totSize:]).Pointer()
		strSec := *(*C.StringSec)(unsafe.Pointer(uVal))
		totSize += int(strSec.size)
	}
	if int(ps.sec.size) != totSize {
		log.Printf("whole size: %v", ps.sec.size)
		log.Printf("the add-up: %v", totSize)
		panic(errors.New("error whole section size mismatched"))
	}
	return true
}

func ParseProfileSection(
	pb *topspb.SerializeExecutableData,
	debugStdout io.Writer,
) *metadata.ExecScope {
	data := pb.GetData()

	newPb := NewProfileSecPipBoy(data)
	newPb.verifySize()

	curDataOffset := 0

	getDataChunk := func(walkOver int) []byte {
		curDataOffset += walkOver
		return data[curDataOffset:]
	}

	dataStart := getDataChunk(newPb.HeaderSize())

	// Please follow the business order listed below
	// Pkt2Op
	// OpMeta
	// Module
	// MemcpyMeta
	// StringIdSec  (Feb.2022)
	// StringSec
	pkt2OpDict := make(map[int]int)

	addPkt2Op := func(pktId, opId int) {
		if _, ok := pkt2OpDict[pktId]; ok {
			panic(errors.New("duplicate packet id entry"))
		}
		pkt2OpDict[pktId] = opId
	}

	for i := 0; i < newPb.Pkt2OpCount(); i++ {
		uVal := reflect.ValueOf(dataStart[i*int(GetPkt2OpSecSize()):]).Pointer()
		pkt2opSec := *(*C.Pkt2OpSec)(unsafe.Pointer(uVal))
		addPkt2Op(int(pkt2opSec.pkt), int(pkt2opSec.op))
	}
	fmt.Fprintf(debugStdout, "# %v pkt to op entries added", len(pkt2OpDict))
	fmt.Fprintf(debugStdout, "\n")

	// Now we skip to op-meta start
	dataStart = getDataChunk(newPb.pkt2OpSecSize())
	opInformationMap := make(map[int]metadata.DtuOp)
	addOpInfo := func(opId int, name string) {
		if _, ok := opInformationMap[opId]; ok {
			panic(errors.New("duplicate op-id"))
		}
		fmt.Fprintf(debugStdout, "op: <%v>", name)
		fmt.Fprintf(debugStdout, "\n")
		opInformationMap[opId] = metadata.DtuOp{
			OpId:   opId,
			OpName: name,
		}
	}
	for i := 0; i < newPb.OpMetaCount(); i++ {
		uVal := reflect.ValueOf(dataStart[i*int(GetOpMetaSecSize()):]).Pointer()
		opMetaSec := *(*C.OpMetaSec)(unsafe.Pointer(uVal))
		addOpInfo(int(opMetaSec.id), newPb.ExtractStringAt(int(opMetaSec.name)))
	}
	fmt.Fprintf(debugStdout, "%v op info entries added", len(opInformationMap))
	fmt.Fprintf(debugStdout, "\n")

	// skip meta chunk, now we reach module section
	getDataChunk(newPb.opMetaSecSize())
	// skip module section, now we reach dma section
	dataStart = getDataChunk(newPb.moduleSecSize())
	dmaInfoMap := make(map[int]metadata.DmaOp)
	addDmaInfo := func(packetId int, dc map[string]string) {
		if _, ok := dmaInfoMap[packetId]; ok {
			panic(fmt.Errorf("duplicate dma packet id(in meta): %v", packetId))
		}
		fmt.Fprintf(debugStdout, "dma: <%v>", dc["dma_op"])
		fmt.Fprintf(debugStdout, "\n")
		dmaInfoMap[packetId] = metadata.DmaOp{
			PktId:       packetId,
			DmaOpString: dc["dma_op"],
			Input:       dc["input"],
			Output:      dc["output"],
			Attrs:       dc,
		}
	}
	for i := 0; i < newPb.MemcpyCount(); i++ {
		slice := dataStart[i*int(GetPbMemcpySecSize()):]
		uVal := reflect.ValueOf(slice).Pointer()
		memcpyMetaSec := *(*C.PbMemcpySec)(unsafe.Pointer(uVal))
		dc := newPb.ExtractArgsAt(int(memcpyMetaSec.args), int(memcpyMetaSec.args_count))
		addDmaInfo(int(memcpyMetaSec.pkt), dc)
	}

	return metadata.NewExecScope(pb.GetExecUuid(),
		pkt2OpDict,
		opInformationMap,
		metadata.DmaInfoMap{
			Info: dmaInfoMap,
		},
	)
}

func doAssertOnProfileSection() {
	var sec0 C.ProfileSection
	if GetProfileSectionSize() != unsafe.Sizeof(sec0) {
		panic(errProfileSectionSizeMismatched)
	}
	var secOpMeta C.OpMetaSec
	if GetOpMetaSecSize() != unsafe.Sizeof(secOpMeta) {
		panic(errors.New("opmetasize-mismtached"))
	}
	var secPkt2Op C.Pkt2OpSec
	if GetPkt2OpSecSize() != unsafe.Sizeof(secPkt2Op) {
		panic(errors.New("pkt2op sec size mismatched"))
	}

	var secStringSec C.StringSec
	if GetStringSecSize() != unsafe.Sizeof(secStringSec) {
		panic(errors.New("string-sec size mismatched"))
	}
	var secModuleSec C.ModuleSec
	if GetModuleSecSize() != unsafe.Sizeof(secModuleSec) {
		panic(errors.New("module sec size mismatched"))
	}
	var secPbMemcpySec C.PbMemcpySec
	if GetPbMemcpySecSize() != unsafe.Sizeof(secPbMemcpySec) {
		panic(errors.New("pb-memcpysec size mismatched"))
	}
	var secPbStringIdSec C.StringIdSec
	if GetStringIdSecSize() != unsafe.Sizeof(secPbStringIdSec) {
		panic(errors.New("stringid-sec size mismatched"))
	}
}
