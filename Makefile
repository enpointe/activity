# Simple makefile for building a testing module

all: build test

build:
	go build ./...
test:
	go test ./...

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

