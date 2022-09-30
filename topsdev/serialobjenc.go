package topsdev

import (
	"bytes"
	"encoding/binary"

	"git.enflame.cn/hai.bai/dmaster/topsdev/proto/pbdef/topspb"
	"google.golang.org/protobuf/proto"
)

type SerialObjEnc struct {
	buffer *bytes.Buffer
}

type ProfileOpt struct {
	ProcessId       int
	ProfileSections []ProfileSecElement
}

func NewSerailObjEnc() SerialObjEnc {
	return SerialObjEnc{
		buffer: bytes.NewBuffer(nil),
	}
}

func getInt32(value int) *int32 {
	var v = int32(value)
	return &v
}

func (soe SerialObjEnc) EncodeBody(rawdata []byte, profOpt ProfileOpt) []byte {
	if profOpt.ProfileSections == nil {
		panic("profsection must not be nil for encoding")
	}

	packCount := 1
	rawpb := soe.makeProfileData(profOpt, packCount)
	soe.EncodeTypeCode(ProfileDataTypeCode)
	soe.EncodeLength(len(rawpb))
	soe.buffer.Write(rawpb)

	// 1 cluster pack
	clusterbuff := soe.makeClusterPack(rawdata)
	soe.EncodeTypeCode(ClusterPackTypeCode)
	soe.EncodeLength(len(clusterbuff))
	soe.buffer.Write(clusterbuff)

	return soe.buffer.Bytes()
}

func (SerialObjEnc) makeClusterPack(rawdata []byte) []byte {

	pack := &topspb.ClusterDataPack{}
	var zeroVal int32 = 0
	pack.ClusterId = &zeroVal
	pack.DeviceId = &zeroVal
	pack.NodeId = &zeroVal

	// encode all
	lz := len(rawdata)
	// align to 64 bits integer
	count := lz / 8
	timestamps := make([]uint64, count)
	for i := 0; i < count; i++ {
		timestamps[i] = binary.LittleEndian.Uint64(rawdata[i*8:])
	}
	pack.Timestamp = timestamps
	chunk, err := proto.Marshal(pack)
	if err != nil {
		panic(err)
	}
	return chunk
}

func (SerialObjEnc) makeProfileData(profOpt ProfileOpt, packCount int) []byte {
	cmds := "fake"
	sdkVersion := "tops-sdk"
	frameworkVersion := "fake-fmk"
	profileDataName := "I20data"
	profileDataType := "prof-datatype"
	profileDataVer := "prof-dataver"
	osName := "linux"
	osVersion := "4.15"
	product := "I20"
	platform := "Linux"

	var pb topspb.ProfileData
	pb.Dtu = &topspb.DTUProfile{
		Device: &topspb.DTUDeviceInfo{
			Node2Dev: &topspb.IdMapData{
				Id1: new(int32),
				Id2: new(int32),
			},
		},
		Meta: &topspb.DTUProfileMeta{},
	}
	pb.Info = &topspb.ProfileInfo{
		PlatformInfo: []*topspb.PlatformInfo{{
			Product:   &product,
			Platform:  &platform,
			OsName:    &osName,
			OsVersion: &osVersion,
		}},
		VersionInfo: &topspb.VersionInfo{
			SdkVersion:       &sdkVersion,
			FrameworkVersion: &frameworkVersion,
			ProfileVersion: &topspb.ProfileVersion{
				ProfileDataName:    &profileDataName,
				ProfileDataType:    &profileDataType,
				ProfileDataVersion: &profileDataVer,
			},
		},
		CommandInfo: &topspb.CommandInfo{
			StartTimestamp: new(int64),
			EndTimestamp:   new(int64),
			Command:        &cmds,
		},
		ConfigInfo: &topspb.ConfigInfo{},
		ProcessId:  getInt32(profOpt.ProcessId),
	}

	// profileVersion := "profile-ver"
	// Fake the executable
	pb.GetDtu().GetMeta().ExecutableProfileSerialize =
		genProfileMeta(profOpt.ProfileSections)

	// set cluster pack count to 1
	// TODO: split by size
	var one int32 = int32(packCount)
	// var topspb.Clu
	pb.Dtu.Data = &topspb.DTUProfileData{
		DataPackDesc: &topspb.ClusterDataPackDescriptor{
			PackCount: &one,
		},
	}
	chunk, err := proto.Marshal(&pb)
	if err != nil {
		panic(err)
	}
	return chunk
}

func genProfileMeta(profSec []ProfileSecElement) []*topspb.SerializeExecutableData {
	var id int32 = 1
	name := "anonymous"
	var profSer []*topspb.SerializeExecutableData
	for _, prof := range profSec {
		theId := id
		id++
		profSer = append(profSer, &topspb.SerializeExecutableData{
			Id:       &theId,
			Name:     &name,
			Data:     prof.profSecRaw,
			ExecUuid: &prof.execUuid,
		})
	}
	return profSer
}

func (soe SerialObjEnc) EncodeTypeCode(typeCode uint64) {
	typeCodeBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(typeCodeBuffer, typeCode)
	soe.buffer.Write(typeCodeBuffer)
}

func (soe SerialObjEnc) Bytes() []byte {
	return soe.buffer.Bytes()
}

func (soe SerialObjEnc) EncodeLength(length int) {
	lengthBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(lengthBuf, uint64(length))
	soe.buffer.Write(lengthBuf)
}
