# Simple makefile for building a testing module

all: clean build test

build:
	go build ./...
	go build
	
test:
	go test ./...

swagger:
	swag init
	cd docs; go fmt

coverage:
	go test ./... -coverprofile=coverage.out | true
	go tool cover -html=coverage.out

clean:
	@rm -f activity activity.log
