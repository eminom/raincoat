package infoloader

import (
	"git.enflame.cn/hai.bai/dmaster/efintf/efconst"
	"git.enflame.cn/hai.bai/dmaster/rtinfo/rtdata"
)

type OneSolidTaskLoader struct{}

func (OneSolidTaskLoader) LoadTask(bool) (dc map[int]*rtdata.RuntimeTask,
	taskSequentials []int,
	ok bool,
) {
	const theTaskID = efconst.SolidTaskID
	dc = make(map[int]*rtdata.RuntimeTask)
	dc[theTaskID] = &rtdata.RuntimeTask{
		RuntimeTaskBase: rtdata.RuntimeTaskBase{TaskID: theTaskID,
			ExecutableUUID: 0,
			PgMask:         0,
		},
		StartCycle: 0,
		EndCycle:   1 << 63, //
		CycleValid: true,
		MetaValid:  true,
	}
	taskSequentials = []int{theTaskID}
	ok = true
	return
}
