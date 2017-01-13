all: build

build:
	go build

push:
	git push -u origin master
