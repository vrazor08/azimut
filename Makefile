B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags)
REV=$(GITREV)-$(BRANCH)

info:
	- @echo "revision $(REV)"

test:
	go test -v ./...

test-race:
	go test -race -timeout=60s -count 1 ./...

lint:
	@golangci-lint run

run:
	@go run -v .

run-race:
	@go run -race .

build:
	@go build -ldflags "-s -w" -o ./azimut.exe

build-linux:
	@CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o ./azimut
	
build-linux-for-ps:
	@go env -w CGO_ENABLED=0
	@go env -w GOOS=linux
	@go build -ldflags "-s -w" -o ./azimut
	@go env -w CGO_ENABLED=1
	@go env -w GOOS=windows
	
.PHONY: info test test-race lint run run-race build build-linux build-linux-for-ps
