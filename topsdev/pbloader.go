package topsdev

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"sort"

	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/topsdev/proto/pbdef/topspb"
)

type pbLoader struct {
	pbObj *topspb.ProfileData
}

func NewPbLoader(inputFile string) (loader pbLoader, err error) {
	_, body, err := DecodeFile(inputFile)
	if err != nil {
		log.Printf("create pb loader failed for: %v", err)
		return
	}
	pb, err := ParseFromChunk(body)
	if err != nil {
		log.Printf("create pb loader failed for: %v", err)
		return
	}
	loader = pbLoader{
		pbObj: pb,
	}
	return
}

func (pb pbLoader) LoadTask() (taskMap map[int]*rtdata.RuntimeTask, taskIdOrder []int, ok bool) {
	taskMap = make(map[int]*rtdata.RuntimeTask)
	for _, task := range pb.pbObj.Dtu.Runtime.Task.TaskData {
		// fmt.Printf("%v 0x%016x %v\n", *task.TaskId, *task.ExecUuid, *task.PgMask)

		taskId := int(*task.TaskId)
		execUuid := task.GetExecUuid()
		pgMask := int(task.GetPgMask())
		if _, ok := taskMap[taskId]; ok {
			panic(errors.New("duplicate task id"))
		}
		taskMap[taskId] = &rtdata.RuntimeTask{
			TaskID:         taskId,
			ExecutableUUID: execUuid,
			PgMask:         pgMask,
		}
		taskIdOrder = append(taskIdOrder, taskId)
	}
	sort.Ints(taskIdOrder)
	ok = true
	return
}

func (pb pbLoader) LoadTimepoints() (hosttp []rtdata.HostTimeEntry, ok bool) {
	for _, tp := range pb.pbObj.Dtu.Device.SyncPoint {
		hosttp = append(hosttp, rtdata.HostTimeEntry{
			Cid:          int(tp.GetId()),
			Hosttime:     uint64(tp.GetTimestamp()),
			DpfSyncIndex: int(tp.GetDeviceCycle()),
		})
	}
	ok = true
	return
}

func (pb pbLoader) LoadExecScope(execUuid uint64) *metadata.ExecScope {
	for _, seri := range pb.pbObj.Dtu.Meta.ExecutableProfileSerialize {
		if seri.GetExecUuid() == execUuid {
			return ParseProfileSection(seri)
		}
	}
	return nil
}

func (pb pbLoader) LoadWildcards(checkExist func(str string) bool,
	notifyNew func(uint64, *metadata.ExecScope)) {

}

type PbComplex struct {
	pbLoader
	ringbufferContentIdx int
}

func NewPbComplex(name string) (
	pbCom PbComplex,
	err error,
) {
	pbl, err := NewPbLoader(name)
	if err != nil {
		return
	}
	pbCom = PbComplex{pbLoader: pbl}
	return
}

func (pb PbComplex) HasMore() bool {
	return pb.ringbufferContentIdx < 1
}

func (pb *PbComplex) LoadRingBufferContent(cid int) []byte {
	if cid < 0 || cid >= len(pb.pbObj.Dtu.Data.TimestampVec) {
		log.Printf("invalid cid: %v", cid)
		return nil
	}

	tsVec := pb.pbObj.Dtu.Data.TimestampVec[cid]
	buffer := bytes.NewBuffer(nil)
	var out [8]byte
	for _, ts := range tsVec.Timestamp {
		binary.LittleEndian.PutUint64(out[:], ts)
		buffer.Write(out[:])
	}

	pb.ringbufferContentIdx++
	return buffer.Bytes()
}
