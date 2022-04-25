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
