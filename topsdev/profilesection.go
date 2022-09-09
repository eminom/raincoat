package topsdev

/*
#include <stdint.h>
// Make sure all these structures' sizes are multiple of 8 bytes

typedef struct ProfileSectionSub {
  uint64_t section_type;
  uint64_t size;
  uint64_t element_size;
  uint64_t count;
  uint64_t offset;
} ProfileSectionSub;

typedef struct ProfileSection {
  uint64_t size;
  uint64_t profile_section_header_size;
  uint64_t sub_section_count;
  uint64_t flag;
  uint64_t reserved_0;
  uint64_t reserved_1;
  uint64_t reserved_2;
  // New fields are added after reserved_2
  // prof_sub_section = ProfileSection base + profile_section_header_size
  ProfileSectionSub prof_sub_section[0];
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
  int type_id;
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

typedef struct SubOpMetaSec {
  int32_t master_op_id;
  int32_t slave_op_id;
  int32_t tid;
  uint32_t name;
} SubOpMetaSec;

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
size_t GetSubOpMetaSecSize() {
	return sizeof(SubOpMetaSec);
}
size_t GetProfileSubSectionSize() {
	return sizeof(ProfileSectionSub);
}
*/
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"unsafe"

	"git.enflame.cn/hai.bai/dmaster/assert"
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/topsdev/proto/pbdef/topspb"
)

/*

#define PROFSEC_TYPE_PKT2OPMAP 0x10
#define PROFSEC_TYPE_OPTHUNK 0x11
#define PROFSEC_TYPE_MODULETHUNK 0x12
#define PROFSEC_TYPE_MEMCPY_THUNK 0x13
#define PROFSEC_TYPE_STRINGID_THUNK 0x14
#define PROFSEC_TYPE_SUBOPMETATHUNK 0x15
#define PROFSEC_TYPE_STRINGPOOL 0x51
*/

const (
	PROFSEC_TYPE_PKT2OPMAP      = 0x10
	PROFSEC_TYPE_OPTHUNK        = 0x11
	PROFSEC_TYPE_MODULETHUNK    = 0x12
	PROFSEC_TYPE_MEMCPY_THUNK   = 0x13
	PROFSEC_TYPE_STRINGID_THUNK = 0x14
	PROFSEC_TYPE_SUBOPMETATHUNK = 0x15
	PROFSEC_TYPE_STRINGPOOL     = 0x51
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

func GetSubOpMetaSecSize() uintptr {
	return uintptr(C.GetSubOpMetaSecSize())
}

func dumpProfSec(sec C.ProfileSection, execUuidPrompt uint64) {
	fmt.Printf("exec(0x%016x)\n", execUuidPrompt)
	indent := strings.Repeat(" ", 2)
	fmt.Printf(indent+"flag: %v\n", sec.flag)
	fmt.Printf(indent+"header_size: %v\n", sec.profile_section_header_size)
	fmt.Printf(indent+"reserved0: %x\n", sec.reserved_0)
	fmt.Printf(indent+"reserved0: %x\n", sec.reserved_1)
	fmt.Printf(indent+"reserved0: %x\n", sec.reserved_2)
	fmt.Printf(indent+"sub_sections: %v\n", sec.sub_section_count)
	fmt.Printf("\n")
}

type SubInfo struct {
	count       int
	elementSize int
	rawChunk    []byte
}

type ProfileSecPipBoy struct {
	sec C.ProfileSection

	stringPool []byte

	// Cached
	pkt2opRec SubInfo
	opMetaRec SubInfo
	memcpyRec SubInfo
	stringRec SubInfo
	subOpRec  SubInfo
	moduleRec SubInfo
}

type RawDataSet struct {
	subInfo C.ProfileSectionSub
	rawCopy []byte
}

func DetectFormatCode(rawData []byte, execUuidPrompt uint64) int {
	uVal := reflect.ValueOf(rawData).Pointer()
	sec := *(*C.ProfileSection)(unsafe.Pointer(uVal))
	dumpProfSec(sec, execUuidPrompt)
	return int(sec.flag)
}

func NewProfileSecPipBoy(rawData []byte) ProfileSecPipBoy {
	uVal := reflect.ValueOf(rawData).Pointer()
	sec := *(*C.ProfileSection)(unsafe.Pointer(uVal))

	profileSectionSize := int(sec.profile_section_header_size)
	perSubSecSize := int(C.GetProfileSubSectionSize())
	subSectionOffset := profileSectionSize

	var subSecDc = make(map[int]RawDataSet)
	var totDataSize = 0
	for i := 0; i < int(sec.sub_section_count); i++ {
		subChunk :=
			rawData[subSectionOffset+i*perSubSecSize : subSectionOffset+(i+1)*perSubSecSize]
		uValue := reflect.ValueOf(subChunk).Pointer()
		subSec := *(*C.ProfileSectionSub)(unsafe.Pointer(uValue))

		typeCode := int(subSec.section_type)
		if _, ok := subSecDc[typeCode]; ok {
			panic(fmt.Errorf("duplicate sub section type %v", typeCode))
		}
		dataOffset := int(subSec.offset)
		dataSize := int(subSec.size)
		subSecDc[typeCode] = RawDataSet{
			subSec,
			bytes.Repeat(rawData[dataOffset:dataOffset+dataSize], 1),
		}
		totDataSize += dataSize
	}

	//
	// dataOffset := int(sec.profile_section_header_size) +
	// 	int(sec.sub_section_count)*perSubSecSize
	rv := ProfileSecPipBoy{
		sec: sec,
	}
	assert.Assert(len(rawData) == int(sec.size), "must be length verified")

	retrieveFor := func(typeCode int, required bool) SubInfo {
		if sub, ok := subSecDc[typeCode]; ok {
			return SubInfo{
				int(sub.subInfo.count),
				int(sub.subInfo.element_size),
				sub.rawCopy,
			}
		}
		if required {
			panic(fmt.Errorf("no section of type(%v)", typeCode))
		}
		return SubInfo{0, 0, nil}
	}
	rv.pkt2opRec = retrieveFor(PROFSEC_TYPE_PKT2OPMAP, true)
	rv.opMetaRec = retrieveFor(PROFSEC_TYPE_OPTHUNK, true)
	rv.memcpyRec = retrieveFor(PROFSEC_TYPE_MEMCPY_THUNK, true)
	rv.stringRec = retrieveFor(PROFSEC_TYPE_STRINGPOOL, true)
	rv.subOpRec = retrieveFor(PROFSEC_TYPE_SUBOPMETATHUNK, false)
	rv.moduleRec = retrieveFor(PROFSEC_TYPE_MODULETHUNK, false)

	assert.Assert(len(rawData) == profileSectionSize+
		int(sec.sub_section_count)*perSubSecSize+
		totDataSize,
		"must be length verified, with all sub sections: ")
	//
	rv.initStringPool()
	return rv
}

func (ps *ProfileSecPipBoy) initStringPool() {
	stringSecChunk := ps.stringRec.rawChunk
	uVal := reflect.ValueOf(stringSecChunk).Pointer()
	stringSec := *(*C.StringSec)(unsafe.Pointer(uVal))
	stringRaw := stringSecChunk[GetStringSecSize():]
	assert.Assert(int(stringSec.size) == len(stringSecChunk),
		"%v == %v, string chunk size", int(stringSec.size), len(stringSecChunk))
	ps.stringPool = stringRaw
}

func (ps ProfileSecPipBoy) GetFormatCode() int {
	return int(ps.sec.flag)
}

func (ps ProfileSecPipBoy) Pkt2OpCount() int {
	return ps.pkt2opRec.count
}

func (ps ProfileSecPipBoy) OpMetaCount() int {
	return ps.opMetaRec.count
}

func (ps ProfileSecPipBoy) MemcpyCount() int {
	return ps.memcpyRec.count
}

func (ps ProfileSecPipBoy) SubOpCount() int {
	return ps.subOpRec.count
}

func (ps ProfileSecPipBoy) ModuleThunkCount() int {
	return ps.moduleRec.count
}

func (ps ProfileSecPipBoy) HeaderSize() int {
	return int(ps.sec.profile_section_header_size)
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
		i := start
		for i < len(ps.stringPool) && ps.stringPool[i] != 0 {
			i++
		}
		oneStr := string(ps.stringPool[start:i])
		start = i + 1 // skip the last zero
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

func ParseProfileSection(
	pb *topspb.SerializeExecutableData,
	debugStdout io.Writer,
) *metadata.ExecScope {
	data := pb.GetData()
	rv, _ := ParseProfileSectionFromData(data, pb.GetExecUuid(), debugStdout)
	return rv
}

func ParseProfileSectionFromData(
	data []byte,
	execUuid uint64,
	debugStdout io.Writer,
) (*metadata.ExecScope, int) {
	fc := DetectFormatCode(data, execUuid)
	if fc != 1 {
		fmt.Fprintf(os.Stderr,
			"exec 0x%016x: format(%v) is not supported\n", execUuid, fc)
		return nil, fc
	}

	newPb := NewProfileSecPipBoy(data)

	var rawChunk []byte
	var elementSize int
	// Please follow the business order listed below
	// Pkt2Op
	// OpMeta
	// Module
	// MemcpyMeta
	// StringIdSec  (Feb.2022)
	// StringSec
	pkt2OpDict := make(map[int]int)
	addPkt2Op := func(pktId, opId int) {
		if orgOpId, ok := pkt2OpDict[pktId]; ok {
			// panic(errors.New("duplicate packet id entry"))
			fmt.Printf("error duplicate pid %v\n", pktId)
			if orgOpId != opId {
				panic(errors.New("mismatched pid to opid"))
			}
		}
		pkt2OpDict[pktId] = opId
	}
	rawChunk = newPb.pkt2opRec.rawChunk
	elementSize = newPb.pkt2opRec.elementSize
	for i := 0; i < newPb.Pkt2OpCount(); i++ {
		uVal := reflect.ValueOf(rawChunk[i*elementSize:]).Pointer()
		pkt2opSec := *(*C.Pkt2OpSec)(unsafe.Pointer(uVal))
		addPkt2Op(int(pkt2opSec.pkt), int(pkt2opSec.op))
	}
	fmt.Fprintf(debugStdout, "# %v pkt to op entries added", len(pkt2OpDict))
	fmt.Fprintf(debugStdout, "\n")

	// Now we skip to op-meta start
	opInformationMap := make(map[int]metadata.DtuOp)
	addOpInfo := func(opId int, name string,
		moduleId int,
		moduleName string,
		kind string,
		fusionKind string,
		meta string) {
		if _, ok := opInformationMap[opId]; ok {
			panic(errors.New("duplicate op-id"))
		}
		fmt.Fprintf(debugStdout, "op: <%v>", name)
		fmt.Fprintf(debugStdout, "\n")
		opInformationMap[opId] = metadata.DtuOp{
			OpId:       opId,
			OpName:     name,
			ModuleId:   moduleId,
			ModuleName: moduleName,
			Kind:       kind,
			FusionKind: fusionKind,
			MetaString: meta,
		}
	}
	rawChunk = newPb.opMetaRec.rawChunk
	elementSize = newPb.opMetaRec.elementSize
	for i := 0; i < newPb.OpMetaCount(); i++ {
		uVal := reflect.ValueOf(rawChunk[i*elementSize:]).Pointer()
		opMetaSec := *(*C.OpMetaSec)(unsafe.Pointer(uVal))
		addOpInfo(int(opMetaSec.id),
			newPb.ExtractStringAt(int(opMetaSec.name)),
			int(opMetaSec.module_id),
			newPb.ExtractStringAt(int(opMetaSec.module_name)),
			newPb.ExtractStringAt(int(opMetaSec.kind)),
			newPb.ExtractStringAt(int(opMetaSec.fusion_kind)),
			newPb.ExtractStringAt(int(opMetaSec.meta)),
		)
	}
	fmt.Fprintf(debugStdout, "%v op info entries added", len(opInformationMap))
	fmt.Fprintf(debugStdout, "\n")

	// Memcpy
	dmaInfoMap := make(map[int]metadata.DmaOp)
	addDmaInfo := func(memcpyChunk C.PbMemcpySec) {
		packetId := int(memcpyChunk.pkt)
		engineType := newPb.ExtractStringAt(int(memcpyChunk.engine_type))
		engineIndx := int(memcpyChunk.engine_id)
		// tiling mode is now -1 from sec element
		// log.Printf("DmaEngineType code is %v, engine-idx(%v)\n",
		// 	engineType, engineIndx)

		dc := newPb.ExtractArgsAt(int(memcpyChunk.args), int(memcpyChunk.args_count))
		if _, ok := dmaInfoMap[packetId]; ok {
			panic(fmt.Errorf("duplicate dma packet id(in meta): %v", packetId))
		}
		fmt.Fprintf(debugStdout, "dma: <%v>", dc["dma_op"])
		fmt.Fprintf(debugStdout, "\n")
		dmaInfoMap[packetId] = metadata.DmaOp{
			PktId:       packetId,
			EngineTy:    engineType,
			EngineIndex: engineIndx,
			DmaOpString: dc["dma_op"],
			Input:       dc["input"],
			Output:      dc["output"],
			Attrs:       dc,
		}
	}
	rawChunk = newPb.memcpyRec.rawChunk
	elementSize = newPb.memcpyRec.elementSize
	for i := 0; i < newPb.MemcpyCount(); i++ {
		slice := rawChunk[i*elementSize:]
		uVal := reflect.ValueOf(slice).Pointer()
		memcpyMetaSec := *(*C.PbMemcpySec)(unsafe.Pointer(uVal))
		addDmaInfo(memcpyMetaSec)
	}

	// SubOp
	subOpInfoMap := make(map[int][]metadata.SubOpMeta)
	rawChunk = newPb.subOpRec.rawChunk
	elementSize = newPb.subOpRec.elementSize
	for i := 0; i < newPb.SubOpCount(); i++ {
		slice := rawChunk[i*elementSize:]
		uVal := reflect.ValueOf(slice).Pointer()
		subOpMetaSec := *(*C.SubOpMetaSec)(unsafe.Pointer(uVal))
		masterOpId := int(subOpMetaSec.master_op_id)
		slaveOpId := int(subOpMetaSec.slave_op_id)
		tid := int(subOpMetaSec.tid)
		name := newPb.ExtractStringAt(int(subOpMetaSec.name))
		subOpInfoMap[masterOpId] = append(subOpInfoMap[masterOpId],
			metadata.SubOpMeta{
				MasterOpId: masterOpId,
				SlaveOpId:  slaveOpId,
				Tid:        tid,
				SubOpName:  name,
			})
	}
	for _, subVec := range subOpInfoMap {
		sort.Sort(metadata.SubOpMetaVec(subVec))
	}

	// Module
	moduleThunkMap := make(map[string]metadata.ModThunk)
	rawChunk = newPb.moduleRec.rawChunk
	elementSize = newPb.moduleRec.elementSize
	for i := 0; i < newPb.ModuleThunkCount(); i++ {
		slice := rawChunk[i*elementSize:]
		uVal := reflect.ValueOf(slice).Pointer()
		moduleSec := *(*C.ModuleSec)(unsafe.Pointer(uVal))
		id := int(moduleSec.id)
		name := newPb.ExtractStringAt(int(moduleSec.name))
		content := newPb.ExtractStringAt(int(moduleSec.data))
		modTh := metadata.NewModThunk(id, name, content)
		key := modTh.ToNameString()
		if _, ok := moduleThunkMap[key]; ok {
			fmt.Fprintf(os.Stderr, "%v already exists\n", key)
			continue
		}
		moduleThunkMap[key] = modTh
	}

	return metadata.NewExecScope(execUuid,
		pkt2OpDict,
		opInformationMap,
		metadata.DmaInfoMap{
			Info: dmaInfoMap,
		},
		subOpInfoMap,
		moduleThunkMap,
	), fc
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
	var secSubOpMetaSec C.SubOpMetaSec
	if GetSubOpMetaSecSize() != unsafe.Sizeof(secSubOpMetaSec) {
		panic(errors.New("subop-sec size mismatched"))
	}
}
