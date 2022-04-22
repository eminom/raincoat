package topsdev

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

/*
const size_t MAGIC_SIZE = 8;  // libprofile data file identification
const size_t TAG_SIZE = 32;   // src code git tag, e.g. 202004081830.1.2
const size_t BNUM_SIZE = 8;   // CI build number
const size_t SHA_SIZE = 16;   // src code SHA
const size_t MD5_SIZE = 32;   // cereal serialized data MD5
const size_t RESERVED_SIZE = 32;

int headerSize() {
	return MAGIC_SIZE + TAG_SIZE + BNUM_SIZE + SHA_SIZE + MD5_SIZE +
           RESERVED_SIZE;
}
*/
import "C"

var (
	errIncompleteHeader = errors.New("incomplete header")
	errLessThanExpected = errors.New("less-than-expected")
)

func GetProfileDataHeaderSize() int {
	return int(C.headerSize())
}

type ProfHeader struct {
	Magic  []byte
	Tag    []byte
	BNum   []byte
	Sha256 []byte
	MD5    []byte
	Reserv []byte
}

func (ph ProfHeader) ToString() string {
	return fmt.Sprintf("%v+%v", string(ph.Magic), string(ph.Tag))
}

type TokenScanner struct {
	buf *bytes.Buffer
}

func newTokenScanner(inbuf []byte) *TokenScanner {
	return &TokenScanner{
		buf: bytes.NewBuffer(inbuf),
	}
}

func (t *TokenScanner) NextChunk(need int) []byte {
	rv := make([]byte, need)
	read, err := t.buf.Read(rv)
	if err != nil {
		panic(err)
	}
	if read != len(rv) {
		log.Printf("less than expected")
		panic(errLessThanExpected)
	}
	return rv
}

func openZipFile(inputFile string) (*os.File, error) {
	zipReader, zipErr := zip.OpenReader(inputFile)
	if zipErr != nil {
		return os.Open(inputFile)
	}
	var fileNames []string
	for _, file := range zipReader.Reader.File {
		zippedFile, err := file.Open()
		if err != nil {
			panic(err)
		}
		tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix-")
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(tmpFile, zippedFile)
		if err != nil {
			panic(err)
		}
		tmpFile.Close()
		fileNames = append(fileNames, tmpFile.Name())
	}

	if len(fileNames) <= 0 {
		return nil, errors.New("no suitable file for content")
	}
	if len(fileNames) > 1 {
		log.Printf("warning: INPUT more THAN 1")
	}
	return os.Open(fileNames[0])
}

func DecodeFile(inputFile string) (hd ProfHeader, body []byte, err error) {
	fin, err := openZipFile(inputFile)
	if err != nil {
		log.Printf("error open \"%v\":%v", inputFile, err)
		return
	}
	defer fin.Close()

	st, err := fin.Stat()
	if err != nil {
		log.Printf("could not get file info: %v", err)
		return
	}
	bodySize := st.Size() - int64(GetProfileDataHeaderSize())

	headBytes := make([]byte, GetProfileDataHeaderSize())
	dwRead, err := fin.Read(headBytes)
	if err != nil {
		log.Printf("error read file: %v\n", err)
		return
	}
	if dwRead != len(headBytes) {
		log.Printf("incomplete header")
		err = errIncompleteHeader
		return
	}
	coco := newTokenScanner(headBytes)
	hd = ProfHeader{
		Magic:  coco.NextChunk(int(C.MAGIC_SIZE)),
		Tag:    coco.NextChunk(int(C.TAG_SIZE)),
		BNum:   coco.NextChunk(int(C.BNUM_SIZE)),
		Sha256: coco.NextChunk(int(C.SHA_SIZE)),
		MD5:    coco.NextChunk(int(C.MD5_SIZE)),
		Reserv: coco.NextChunk(int(C.RESERVED_SIZE)),
	}

	body = make([]byte, bodySize)

	left := bodySize
	var writePtr int64 = 0
	for left > 0 {
		dwRead, err = fin.Read(body[writePtr:])
		log.Printf("read rawfile: %v byte(s) read", dwRead)
		if dwRead == 0 {
			break
		}
		writePtr += int64(dwRead)
		left -= int64(dwRead)
	}
	if left > 0 {
		log.Fatalf("%v bytes missed\n", left)
	}
	return
}
