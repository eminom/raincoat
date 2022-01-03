package topsdev

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"git.enflame.cn/hai.bai/dmaster/topsdev/proto/pbdef/topspb"
	"github.com/golang/protobuf/proto"
)

func ParseFromChunk(body []byte) (*topspb.ProfileData, error) {
	var data topspb.ProfileData
	err := proto.Unmarshal(body, &data)
	if err == nil {
		return &data, nil
	}
	return nil, err
}

func dumpTask(pb *topspb.ProfileData) {
	fmt.Printf("# Dump of task info\n")
	for _, task := range pb.Dtu.Runtime.Task.TaskData {
		fmt.Printf("%v 0x%016x %v\n", *task.TaskId, *task.ExecUuid, *task.PgMask)
	}
	fmt.Println()
}

func dumpTimepoints(pb *topspb.ProfileData) {
	fmt.Printf("# Dump of time sync info\n")
	for _, timepoint := range pb.Dtu.Device.SyncPoint {
		fmt.Printf("%v %v %v\n", *timepoint.Id,
			*timepoint.Timestamp,
			*timepoint.DeviceCycle)
	}
	fmt.Println()
}

func dumpDpfringbuffer(pb *topspb.ProfileData) {
	fout, err := os.Create("pbdump.data")
	if err != nil {
		log.Printf("Could not create an output to store data")
		return
	}
	defer fout.Close()
	var out [8]byte
	for cid, tsVec := range pb.Dtu.Data.TimestampVec {
		fmt.Printf("cid(%v) count: %v\n", cid, len(tsVec.Timestamp))
		for _, ts := range tsVec.Timestamp {
			binary.LittleEndian.PutUint64(out[:], ts)
			fout.Write(out[:])
		}
	}
}