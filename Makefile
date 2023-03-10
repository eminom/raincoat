

RELEASE_DATE=$(shell date +%H%M%S_%Y%m%d)

all:
	go build -o build/dmaster
	go build -o build/fakeraw cmds/fake.go

install:	all
	cp build/dmaster ${HOME}/bin
	cp build/fakeraw ${HOME}/bin
	cp cmds/fakebat ${HOME}/bin

linux:
	GOARCH=amd64 GOOS=linux go build -o build_linux/dmaster

pack:	all
	cp scripts/* build
	cp cmds/fakebat build
	tar czvf dmaster.tar.gz build/*
	sha256sum dmaster.tar.gz

exec:
	xgo --targets=windows/amd64 .

linux_release:
	go build -o release.01/dmaster
	tar czvf "dmaster_${RELEASE_DATE}.tar.gz" release.01/dmaster
	rm -rf release.01
