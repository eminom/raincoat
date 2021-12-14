all:
	go build -o build/dmaster

install:	all
	cp build/dmaster ${HOME}/bin