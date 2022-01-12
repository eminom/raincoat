package infoloader

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
)

type DmaOpFormatFetcher interface {
	FetchDmaOpDict(string) map[int]metadata.DmaOp
}

type compatibleDmaFetcher struct{}

/*
ofs << sec->pkt << " " << strConverter.CheckoutStringAt(sec->dma_op)
            << " " << strConverter.CheckoutStringAt(sec->engine_ty) << " "
            << sec->engine_index << " "
            << strConverter.CheckoutStringAt(sec->input) << " "
            << strConverter.CheckoutStringAt(sec->output) << " "
            << strConverter.CheckoutStringAt(sec->attrs) << "\n";
*/
func (compatibleDmaFetcher) FetchDmaOpDict(
	filename string,
) map[int]metadata.DmaOp {
	fin, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fin.Close()

	dmaOpDict := make(map[int]metadata.DmaOp)
	scan := bufio.NewScanner(fin)
	for {
		if !scan.Scan() {
			break
		}
		pktId, dmaOp := parseSingleLineV0(scan.Text())
		if _, ok := dmaOpDict[pktId]; ok {
			panic(fmt.Errorf("duplicated dma op id: %v", pktId))
		}
		dmaOpDict[pktId] = dmaOp
	}
	return dmaOpDict
}

func parseSingleLineV0(text string) (int, metadata.DmaOp) {
	vs := XSplit(text, 7)
	pktId, err := strconv.ParseInt(vs[0], 10, 32)
	if err != nil {
		panic(err)
	}
	dmaOp, engineTy := vs[1], vs[2]
	engineIdx, err := strconv.ParseInt(vs[3], 10, 32)
	if err != nil {
		panic(err)
	}
	return int(pktId), metadata.DmaOp{
		PktId:       int(pktId),
		DmaOpString: dmaOp,
		EngineTy:    engineTy,
		EngineIndex: int(engineIdx),
		Input:       vs[4],
		Output:      vs[5],
		Attrs:       parserAttrsV0(vs[6]),
	}
}

func parserAttrsV0(attrText string) map[string]string {
	fields := strings.Fields(attrText)
	dc := make(map[string]string)
	for _, chunk := range fields {
		chunk = strings.TrimRight(chunk, ",")
		vs := strings.Split(chunk, "=")
		if len(vs) == 2 {
			// duplicated not concerned for now.
			dc[vs[0]] = vs[1]
		}
	}
	return dc
}

type pbDmaMetaFetcher struct{}

func (pbDmaMetaFetcher) FetchDmaOpDict(
	filename string,
) map[int]metadata.DmaOp {
	fin, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fin.Close()

	dmaOpDict := make(map[int]metadata.DmaOp)
	scan := bufio.NewScanner(fin)
	var curDma metadata.DmaOp
	curDma.PktId = -1
	appendNewItem := func() {
		if curDma.PktId >= 0 {
			if _, ok := dmaOpDict[curDma.PktId]; ok {
				panic("duplicated")
			}
			dmaOpDict[curDma.PktId] = curDma
		}
		curDma = metadata.DmaOp{} // dummy, dict is null
	}

	for {
		if !scan.Scan() {
			break
		}
		text := scan.Text()
		if strings.HasPrefix(text, " ") {
			// key, and the rest are all values
			text = strings.TrimLeft(text, " ")
			vs := XSplit(text, 2)
			// Attach to current dma Op.
			curDma.Attrs[vs[0]] = vs[1]
			continue
		} else {
			appendNewItem()

			// OK. Create a new session for DmaOp
			vs := XSplit(text, 4)
			if len(vs) != 4 {
				panic(fmt.Errorf(
					"not expecting for dma op input line: %v", text))
			}
			pktId, err := strconv.ParseInt(vs[0], 10, 32)
			if err != nil {
				panic(err)
			}
			engineIndex, err := strconv.ParseInt(vs[2], 10, 32)
			if err != nil {
				panic(err)
			}
			// And Op id is missing
			curDma = metadata.DmaOp{
				PktId:       int(pktId),
				EngineTy:    vs[3],
				EngineIndex: int(engineIndex),
				Attrs:       make(map[string]string),
			}
		}
	}
	appendNewItem()
	return dmaOpDict
}
