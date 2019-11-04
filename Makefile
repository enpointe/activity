# Simple makefile for building a testing module

all: build test

build:
	go build
test:
	go test -v