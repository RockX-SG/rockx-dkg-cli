GOBASE = $(shell pwd)
GOBIN = $(GOBASE)/build/bin
GOCMD = $(GOBASE)/cmd

build:
	go build -o $(GOBIN)/api  $(GOCMD)/api/main.go

build_messenger:
	go build -o $(GOBIN)/messenger  $(GOCMD)/messenger/main.go

build_node:
	go build -o $(GOBIN)/node  $(GOCMD)/node/main.go $(GOCMD)/node/app_params.go

build_verify:
	go build -o $(GOBIN)/verify  $(GOCMD)/verify/main.go

deps:
	go mod download

test:
	go test -v -cover ./...  -coverprofile .testCoverage.txt

clean:
	rm $(GOBIN)/frostdkgdemo
	rm *.log

all: test build

.PHONY: all test clean build