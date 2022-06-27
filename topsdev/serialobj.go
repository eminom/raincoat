package topsdev

import (
	"encoding/binary"
	"errors"
	"log"

	"git.enflame.cn/hai.bai/dmaster/topsdev/proto/pbdef/topspb"
	"github.com/golang/protobuf/proto"
)

type SerialObj struct {
	data []byte
}

const (
	ProfileDataTypeCode = 101
	ClusterPackTypeCode = 102
)

var (
	errUnexpectedTypeCode = errors.New("unexpected type-code")
	errUnableToDecodePb   = errors.New("unable to decode pb")
)

func (s *SerialObj) decodeUint64() uint64 {
	val := binary.BigEndian.Uint64(s.data)
	s.data = s.data[8:]
	return val
}

func (s *SerialObj) DecodeLength() int {
	return int(s.decodeUint64())
}

func (s *SerialObj) DecodeTypeCode() int {
	return int(s.decodeUint64())
}

func (s *SerialObj) decodeProfileData() *topspb.ProfileData {
	ty := s.DecodeTypeCode()
	if ty != ProfileDataTypeCode {
		panic(errUnexpectedTypeCode)
	}
	sz := s.DecodeLength()
	chunk := s.data[:sz]
	s.data = s.data[sz:]

	var pb topspb.ProfileData
	err := proto.Unmarshal(chunk, &pb)
	if err != nil {
		log.Printf("error decode ProfileData")
		return nil
	}
	return &pb
}

func (s *SerialObj) DecodeProfileData() (*topspb.ProfileData, error) {
	pb := s.decodeProfileData()

	var timeVec *[]*topspb.DTUActivityTimestamp = &pb.GetDtu().GetData().TimestampVec
	// TODO: Hard-coded
	*timeVec = []*topspb.DTUActivityTimestamp{}
	for i := 0; i < 4; i++ {
		var cluster topspb.DTUActivityTimestamp
		(*timeVec) = append(*timeVec, &cluster)
	}

	packCount := int(pb.Dtu.Data.DataPackDesc.GetPackCount())
	log.Printf("cluster pack count is %v", packCount)
	for i := 0; i < packCount; i++ {
		ty := s.DecodeTypeCode()
		if ty != ClusterPackTypeCode {
			return nil, errUnexpectedTypeCode
		}
		sz := s.DecodeLength()
		chunk := s.data[:sz]
		s.data = s.data[sz:]
		var pack topspb.ClusterDataPack
		err := proto.Unmarshal(chunk, &pack)
		if err != nil {
			log.Printf("error decode Cluster pack")
			return nil, errUnableToDecodePb
		}
		cid := pack.GetClusterId()
		(*timeVec)[cid].Timestamp = append((*timeVec)[cid].Timestamp, pack.Timestamp...)
	}
	return pb, nil
}
