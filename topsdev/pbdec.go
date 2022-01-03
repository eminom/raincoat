package topsdev

import (
	"fmt"

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
	for _, task := range pb.Dtu.Runtime.Task.TaskData {
		fmt.Printf("%v 0x%016x %v\n", *task.TaskId, *task.ExecUuid, *task.PgMask)
	}
}
