all:
	go build -o build/dmaster

install:	all
	cp build/dmaster ${HOME}/bin

pack:	all
	cp scripts/* build
	tar czvf dmaster.tar.gz build/*