

RELEASE_DATE=$(shell date +%H%M%S_%Y%m%d)

all:
	go build -o build/dmaster

install:	all
	cp build/dmaster ${HOME}/bin

linux:
	GOARCH=amd64 GOOS=linux go build -o build_linux/dmaster

pack:	all
	cp scripts/* build
	tar czvf dmaster.tar.gz build/*
	sha256sum dmaster.tar.gz

exec:
	xgo --targets=windows/amd64 .

linux_release:
	go build -o release.01/dmaster
	tar czvf "dmaster_${RELEASE_DATE}.tar.gz" release.01/dmaster
	rm -rf release.01
