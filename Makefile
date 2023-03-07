GOBASE = $(shell pwd)
GOBIN = $(GOBASE)/build/bin
GOCMD = $(GOBASE)/cmd
VERSION = 0.0.1

build:
	go build -o $(GOBIN)/rockx-dkg-cli  $(GOCMD)/cli/main.go

build_messenger:
	go build -o $(GOBIN)/messenger  $(GOCMD)/messenger/main.go

build_node:
	go build -o $(GOBIN)/node  $(GOCMD)/node/main.go $(GOCMD)/node/app_params.go

build_verify:
	go build -o $(GOBIN)/verify  $(GOCMD)/verify/main.go

release:
	GOOS=linux GOARCH=amd64 go build -o $(GOBIN)/rockx-dkg-messenger  $(GOCMD)/messenger/main.go
	GOOS=linux GOARCH=amd64 go build -o $(GOBIN)/rockx-dkg-node  $(GOCMD)/node/main.go $(GOCMD)/node/app_params.go
	GOOS=linux GOARCH=amd64 go build -o $(GOBIN)/rockx-dkg-cli  $(GOCMD)/cli/main.go
	tar -czf rockx-messenger.$(VERSION).tar.gz $(GOBIN)/rockx-dkg-messenger
	tar -czf rockx-node.$(VERSION).tar.gz $(GOBIN)/rockx-dkg-node
	tar -czf rockx-cli.$(VERSION).tar.gz $(GOBIN)/rockx-dkg-cli

deps:
	go mod download

test:
	go test -v -cover ./...  -coverprofile .testCoverage.txt

clean:
	rm deposit-data_*
	rm dkg_results_*
	rm $(GOBIN)/*
	rm *.log

all: test build

.PHONY: all test clean build
