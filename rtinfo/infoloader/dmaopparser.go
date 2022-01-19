package infoloader

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"git.enflame.cn/hai.bai/dmaster/meta/metadata"
)

func newDmaInfoMap(dc map[int]metadata.DmaOp) metadata.DmaInfoMap {
	return metadata.DmaInfoMap{
		Info: dc,
	}
}

type DmaOpFormatFetcher interface {
	FetchDmaOpDict(string) metadata.DmaInfoMap
}

type compatibleDmaFetcher struct{}

func NewCompatibleDmaInfoLoader() DmaOpFormatFetcher {
	return compatibleDmaFetcher{}
}

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
) metadata.DmaInfoMap {
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
		text := scan.Text()
		pktId, dmaOp, err := parseSingleLineV0(text)
		if err != nil {
			log.Printf("error: parsing %v for text %v", filename, text)
			panic(err)
		}
		if _, ok := dmaOpDict[pktId]; ok {
			panic(fmt.Errorf("duplicated dma op id: %v", pktId))
		}
		dmaOpDict[pktId] = dmaOp
	}
	return newDmaInfoMap(dmaOpDict)
}

// Split into 7 segments
func elementSplitAndCombines(text string) []string {
	vs := strings.Fields(text)
	if len(vs) == 0 {
		return nil
	}
	var rvs = []string{vs[0]}
	pos := 0
	for i := 1; i < len(vs); i++ {
		if strings.HasSuffix(rvs[pos], ",") {
			rvs[pos] += vs[i]
			continue
		}
		pos++
		rvs = append(rvs, vs[i])
	}
	return rvs
}

func specialCombines(vs []string) []string {
	var rv []string
	for i := 0; i < len(vs); i++ {
		if vs[i] == "sip" && i < len(vs)-1 && vs[i+1] == "launch" {
			rv = append(rv, "sip launch")
			i++
			continue
		}
		rv = append(rv, vs[i])
	}
	return rv
}

func parseSingleLineV0(text string) (int, metadata.DmaOp, error) {
	vs := elementSplitAndCombines(text)
	vs = specialCombines(vs)

	pktId, err := strconv.ParseInt(vs[0], 10, 32)
	if err != nil {
		panic(err)
	}
	dmaOp, engineTy := vs[1], vs[2]
	engineIdx, err := strconv.ParseInt(vs[3], 10, 32)
	if err != nil {
		return 0, metadata.DmaOp{},
			fmt.Errorf("error parsing\n%v\nerr: %v", text, err)
	}
	attrsStr := ""
	if len(vs) >= 7 {
		attrsStr = vs[6]
	}
	return int(pktId), metadata.DmaOp{
		PktId:       int(pktId),
		DmaOpString: dmaOp,
		EngineTy:    engineTy,
		EngineIndex: int(engineIdx),
		Input:       vs[4],
		Output:      vs[5],
		Attrs:       parserAttrsV0(attrsStr),
	}, nil
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

func NewPbDmaInfoLoader() DmaOpFormatFetcher {
	return pbDmaMetaFetcher{}
}

func (pbDmaMetaFetcher) FetchDmaOpDict(
	filename string,
) metadata.DmaInfoMap {
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
			// Special extraction
			if dmaOp, ok := curDma.Attrs["dma_op"]; ok {
				curDma.DmaOpString = dmaOp
			}
			if engine, ok := curDma.Attrs["engine"]; ok {
				curDma.EngineTy = engine
			}
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
			if len(vs) < 4 {
				panic(fmt.Errorf(
					"not expecting for dma op input line: %v", text))
			}
			vs[3] = strings.Join(vs[3:], ",")
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
	return newDmaInfoMap(dmaOpDict)
}
