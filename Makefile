VERSION = 0.2.4
GOBASE = $(shell pwd)
GOBIN = $(GOBASE)/build/bin
GOCMD = $(GOBASE)/cmd

build:
	go build -o $(GOBIN)/rockx-dkg-cli  $(GOCMD)/cli/main.go

build_messenger:
	go build -o $(GOBIN)/messenger  $(GOCMD)/messenger/main.go

build_node:
	go build -o $(GOBIN)/node  $(GOCMD)/node/main.go $(GOCMD)/node/app_params.go

build_verify:
	go build -o $(GOBIN)/verify  $(GOCMD)/verify/main.go

release_darwin_arm64:
	GOOS=darwin GOARCH=arm64 go build -o $(GOBIN)/darwin_arm64/rockx-dkg-messenger  $(GOCMD)/messenger/main.go
	GOOS=darwin GOARCH=arm64 go build -o $(GOBIN)/darwin_arm64/rockx-dkg-node  $(GOCMD)/node/main.go $(GOCMD)/node/app_params.go
	GOOS=darwin GOARCH=arm64 go build -o $(GOBIN)/darwin_arm64/rockx-dkg-cli  $(GOCMD)/cli/main.go
	
	mkdir -p $(GOBASE)/release/$(VERSION)

	cd $(GOBIN)/darwin_arm64 && pwd && \
	tar -czvf $(GOBASE)/release/$(VERSION)/rockx-dkg-messenger.$(VERSION).darwin.arm64.tar.gz rockx-dkg-messenger && \
	tar -czvf $(GOBASE)/release/$(VERSION)/rockx-dkg-node.$(VERSION).darwin.arm64.tar.gz rockx-dkg-node && \
	tar -czvf $(GOBASE)/release/$(VERSION)/rockx-dkg-cli.$(VERSION).darwin.arm64.tar.gz rockx-dkg-cli && \
	cd $(GOBASE)

release_linux_amd64:
	GOOS=linux GOARCH=amd64 go build -o $(GOBIN)/linux_amd64/rockx-dkg-messenger  $(GOCMD)/messenger/main.go
	GOOS=linux GOARCH=amd64 go build -o $(GOBIN)/linux_amd64/rockx-dkg-node  $(GOCMD)/node/main.go $(GOCMD)/node/app_params.go
	GOOS=linux GOARCH=amd64 go build -o $(GOBIN)/linux_amd64/rockx-dkg-cli  $(GOCMD)/cli/main.go
	
	mkdir -p $(GOBASE)/release/$(VERSION)

	cd $(GOBIN)/linux_amd64 && pwd && \
	tar -czvf $(GOBASE)/release/$(VERSION)/rockx-dkg-messenger.$(VERSION).linux.amd64.tar.gz rockx-dkg-messenger && \
	tar -czvf $(GOBASE)/release/$(VERSION)/rockx-dkg-node.$(VERSION).linux.amd64.tar.gz rockx-dkg-node && \
	tar -czvf $(GOBASE)/release/$(VERSION)/rockx-dkg-cli.$(VERSION).linux.amd64.tar.gz rockx-dkg-cli && \
	cd $(GOBASE)

deps:
	go mod download

test:
	go test -v -cover ./...  -coverprofile .testCoverage.txt

clean:
	rm deposit-data_*
	rm dkg_results_*
	rm keyshares_*
	rm $(GOBIN)/*
	rm *.log

all: test build

.PHONY: all test clean build
