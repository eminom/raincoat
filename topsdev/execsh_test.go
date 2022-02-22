package topsdev

import (
	"testing"
)

func TestExecLoad(t *testing.T) {
	inpath := "/home/hai.bai/data_dorado/exec/vg_test_six_thread_ocr_2_3_whole_network_binary.bin"
	DumpProfileSectionFromExecutable(inpath)
}
