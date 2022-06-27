package topsdev

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/codec"
	"git.enflame.cn/hai.bai/dmaster/efintf/affinity"
	"git.enflame.cn/hai.bai/dmaster/meta/dtuarch"
	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/infoloader"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
	"git.enflame.cn/hai.bai/dmaster/topsdev/mimic/mimicdefs"
	"git.enflame.cn/hai.bai/dmaster/topsdev/proto/pbdef/topspb"
)

type pbLoader struct {
	pbObj     *topspb.ProfileData
	inputName string
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
		pbObj:     pb,
		inputName: inputFile,
	}
	return
}

func (pb pbLoader) GetInputName() string {
	return pb.inputName
}

func (pb pbLoader) LoadTask(oneSolid bool) (taskMap map[int]*rtdata.RuntimeTask, taskIdOrder []int, ok bool) {
	taskMap = make(map[int]*rtdata.RuntimeTask)
	for _, task := range pb.pbObj.Dtu.Runtime.Task.TaskData {
		// fmt.Printf("%v 0x%016x %v\n", *task.TaskId, *task.ExecUuid, *task.PgMask)

		taskId := int(*task.TaskId)
		execUuid := task.GetExecUuid()
		pgMask := int(task.GetPgMask())
		if _, ok := taskMap[taskId]; ok {
			// panic(errors.New("duplicate task id"))
			log.Printf("duplicated task id %v\n", taskId)
			continue // do not add again
		}
		taskMap[taskId] = &rtdata.RuntimeTask{
			RuntimeTaskBase: rtdata.RuntimeTaskBase{TaskID: taskId,
				ExecutableUUID: execUuid,
				PgMask:         pgMask,
			},
		}
		taskIdOrder = append(taskIdOrder, taskId)
	}

	if len(taskIdOrder) == 0 || oneSolid {
		// Exception is made
		return infoloader.OneSolidTaskLoader{}.LoadTask(oneSolid)
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

type DummyStdout struct{}

func (DummyStdout) Write(buf []byte) (int, error) {
	return len(buf), nil
}

func (pb pbLoader) LoadExecScope(execUuid uint64) *metadata.ExecScope {
	for _, seri := range pb.pbObj.Dtu.Meta.GetExecutableProfileSerialize() {
		if seri.GetExecUuid() == execUuid {
			return ParseProfileSection(seri, DummyStdout{})
		}
	}
	return nil
}

func (pb pbLoader) DumpMeta() {
	// Meta .Dtu.Meta may be nil
	// if access to ExecutableProfileSerialize directly
	// it may crash
	for _, seri := range pb.pbObj.Dtu.Meta.GetExecutableProfileSerialize() {
		execMeta := ParseProfileSection(seri, DummyStdout{})
		execMeta.DumpDtuOpToFile()
		execMeta.DumpDmaToFile()
		execMeta.DumpPktOpMapToFile()
		execMeta.DumpSubOpToFile()
	}
}

// utils for KVData
func extractStringValueByKey(args []*topspb.KVData, key string) string {
	for _, kv := range args {
		if kv.GetK() == key {
			return kv.GetStringV()
		}
	}
	return ""
}

func (pb pbLoader) DumpCpuOpTrace(inputNameHint string) {
	fout, err := os.Create(fmt.Sprintf("%v_cpuop.pbdumptxt", inputNameHint))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error open cpu op trace file for record: %v\n", err)
		return
	}
	defer fout.Close()

	for _, cpuOp := range pb.pbObj.Cpu.Events {
		// expecting Backend
		cpuOpCatName := cpuOp.GetName()
		fmt.Fprintf(os.Stderr, "%v\n", cpuOpCatName)
		for _, ev := range (*cpuOp).Event {
			startTs := ev.GetStartTimestamp()
			endTs := ev.GetEndTimestamp()
			opName := extractStringValueByKey(ev.GetArgs(), "op_name")
			fmt.Fprintf(fout, "%v %v %v\n", opName, startTs, endTs)
		}
	}
}

func (pb pbLoader) DumpRuntimeInformation(inputNameHint string) {
	pb.dumpTimepoints(inputNameHint)
	pb.dumpRuntimeTasks(inputNameHint)
	pb.dumpPgAffinityInfo(inputNameHint)
}

func (pb pbLoader) DumpMisc(inputNameHint string) {
	pb.dumpMisc(inputNameHint)
}

func (pb pbLoader) dumpTimepoints(inputNameHint string) {
	outName := fmt.Sprintf("%v_timeinfo.pbdumptxt", inputNameHint)
	fout, err := os.Create(outName)
	if err != nil {
		panic(fmt.Errorf("could not open %v: %v", outName, err))
	}
	defer fout.Close()
	for _, tp := range pb.pbObj.Dtu.Device.SyncPoint {
		fmt.Fprintf(fout, "%v %v %v %v\n", int(tp.GetId()),
			uint64(tp.GetTimestamp()),
			int(tp.GetDeviceCycle()),
			uint64(tp.GetClockTimestamp()),
		)
	}
}

func (pb pbLoader) dumpRuntimeTasks(inputNameHint string) {
	outName := fmt.Sprintf("%v_tasks.pbdumptxt", inputNameHint)
	fout, err := os.Create(outName)
	if err != nil {
		panic(fmt.Errorf("could not open %v: %v", outName, err))
	}
	defer fout.Close()
	for _, task := range pb.pbObj.Dtu.Runtime.Task.TaskData {
		fmt.Fprintf(fout, "%v 0x%016x %v\n",
			task.GetTaskId(),
			task.GetExecUuid(),
			task.GetPgMask(),
		)
	}
}

func (pb pbLoader) dumpPgAffinityInfo(inputNameHint string) {
	affInfo := pb.pbObj.Dtu.Runtime.GetAffinity()
	if affInfo == nil {
		return
	}
	outName := fmt.Sprintf("%v_affinity_.pbdumptxt", inputNameHint)
	fout, err := os.Create(outName)
	if err != nil {
		panic(fmt.Errorf("could not open %v: %v", outName, err))
	}
	defer fout.Close()
	fmt.Fprintf(fout, "engine cluster_id engine_id pg_order\n")
	for _, cdma := range affInfo.CdmaAffinity {
		fmt.Fprintf(fout, "cdma %v %v %v\n", cdma.GetClusterId(), cdma.GetEngineId(), cdma.GetPgId())
	}
}

type DoradoCdmaAffinity struct {
	toPgIdx    []int
	archTarget codec.ArchTarget
}

func NewDoradoCdmaAffinity(
	cdmaAffinity []*topspb.EngineAffinity,
	arch codec.ArchTarget) DoradoCdmaAffinity {
	affinityMap := make([]int, arch.CdmaPerC*arch.ClusterPerD)

	def := affinity.DoradoCdmaAffinityDefault{}
	for cid := 0; cid < arch.ClusterPerD; cid++ {
		for eid := 0; eid < arch.CdmaPerC; eid++ {
			affinityMap[cid*arch.CdmaPerC+eid] = def.GetCdmaIdxToPg(cid, eid)
		}
	}

	for _, cdma := range cdmaAffinity {
		idx := arch.CdmaPerC*int(cdma.GetClusterId()) +
			int(cdma.GetEngineId())
		affinityMap[idx] = int(cdma.GetPgId())
	}
	return DoradoCdmaAffinity{
		toPgIdx:    affinityMap,
		archTarget: arch,
	}
}

func (c2p DoradoCdmaAffinity) GetCdmaIdxToPg(cid int, eid int) int {
	idx := cid*c2p.archTarget.CdmaPerC + eid
	return c2p.toPgIdx[idx]
}

func (pb pbLoader) GetCdmaAffinity() affinity.CdmaAffinitySet {
	affInfo := pb.pbObj.Dtu.Runtime.GetAffinity()
	if affInfo == nil {
		return affinity.DoradoCdmaAffinityDefault{}
	}
	return NewDoradoCdmaAffinity(affInfo.GetCdmaAffinity(),
		codec.NewDoradoArchTarget())
}

func (pb pbLoader) dumpMisc(inputNameHint string) {
	outName := fmt.Sprintf("%v_misc.pbdumptxt", inputNameHint)
	fout, err := os.Create(outName)
	if err != nil {
		panic(fmt.Errorf("could not open %v: %v", outName, err))
	}
	defer fout.Close()
	fmt.Fprintf(fout, "config:\n")
	for k, v := range pb.pbObj.Info.ConfigInfo.Config {
		fmt.Fprintf(fout, "%v %v\n", k, v)
	}

	fmt.Fprintf(fout, "\n")
	fmt.Fprintf(fout, "platform:\n")
	for _, pl := range pb.pbObj.GetInfo().GetPlatformInfo() {
		fmt.Fprintf(fout, "%v\n", pl.GetArch())
		fmt.Fprintf(fout, "%v\n", pl.GetPlatform())
		fmt.Fprintf(fout, "%v\n", pl.GetProduct())
	}
}

func (pb pbLoader) ExtractHostInfo() *mimicdefs.HostInfo {
	var verInfo mimicdefs.VersionInfo
	if pb.pbObj.GetInfo() != nil && pb.pbObj.GetInfo().GetVersionInfo() != nil {
		vi := pb.pbObj.GetInfo().GetVersionInfo()
		verInfo = mimicdefs.VersionInfo{
			SdkVersion:         vi.GetSdkVersion(),
			FrameworkVersion:   vi.GetFrameworkVersion(),
			ProfileDataName:    vi.GetProfileVersion().GetProfileDataName(),
			ProfileDataType:    vi.GetProfileVersion().GetProfileDataType(),
			ProfileDataVersion: vi.GetProfileVersion().GetProfileDataVersion(),
		}
	}

	var platInfo []mimicdefs.PlatformInfo
	if pb.pbObj.GetInfo() != nil && pb.pbObj.GetInfo().GetPlatformInfo() != nil {
		for _, pl := range pb.pbObj.GetInfo().GetPlatformInfo() {
			platInfo = append(platInfo, mimicdefs.PlatformInfo{
				Platform:         pl.GetPlatform(),
				OsName:           pl.GetOsName(),
				OsVersion:        pl.GetOsVersion(),
				Product:          pl.GetProduct(),
				OsRelease:        pl.GetOsRelease(),
				HostName:         pl.GetHostName(),
				Arch:             pl.GetArch(),
				CpuModel:         pl.GetCpuModel(),
				CpuVendor:        pl.GetCpuVendor(),
				DistributionName: pl.GetDistributionName(),
			})
		}
	}

	return &mimicdefs.HostInfo{
		CommandInfo: mimicdefs.CommandInfo{
			Command:        pb.pbObj.Info.CommandInfo.GetCommand(),
			StartTimestamp: int64(pb.pbObj.Info.CommandInfo.GetStartTimestamp()),
			EndTimestamp:   pb.pbObj.Info.CommandInfo.GetEndTimestamp(),
		},
		VersionInfo:  verInfo,
		PlatformInfo: platInfo,
	}
}

func (pb pbLoader) LoadWildcards(checkExist func(str string) bool,
	notifyNew func(uint64, *metadata.ExecScope)) {

	// no task record, will load all exec into wild
	for _, seri := range pb.pbObj.Dtu.Meta.GetExecutableProfileSerialize() {
		execMeta := ParseProfileSection(seri, DummyStdout{})
		notifyNew(execMeta.GetExecUuid(), execMeta)
	}
}

func (pb pbLoader) GetArchType() dtuarch.ArchType {
	if pls := pb.pbObj.Info.GetPlatformInfo(); pls != nil {
		for _, platform := range pls {
			switch strings.ToUpper(platform.GetProduct()) {
			case "T20":
				return dtuarch.EnflameT20
			case "I20", "X":
				return dtuarch.EnflameI20
			}
		}
	}
	return dtuarch.EnflameUnknownArch
}

type PbComplex struct {
	pbLoader
	ringbufferContentIdx int
}

// For now one-task-mode is ignored
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

func (pb PbComplex) GetRingBufferCount() int {
	return 1
}

func (pb PbComplex) GetCpuOpTraceSeq() []rtdata.CpuOpAct {
	var cpuOps []rtdata.CpuOpAct
	for _, cpuOp := range pb.pbObj.Cpu.Events {
		name := cpuOp.GetName() // expecting BACKEND
		for _, evt := range cpuOp.Event {
			opName := extractStringValueByKey(evt.GetArgs(), "op_name")
			startTs, endTs := evt.GetStartTimestamp(), evt.GetEndTimestamp()
			cpuOps = append(cpuOps, rtdata.CpuOpAct{
				Cat:            name,
				Name:           opName,
				StartTimestamp: uint64(startTs),
				EndTimestamp:   uint64(endTs),
			})
		}
	}
	return cpuOps
}

func (pb PbComplex) LoadRingBufferContent(cid int, idx int) []byte {
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
