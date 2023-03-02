GOBASE = $(shell pwd)
GOBIN = $(GOBASE)/build/bin
GOCMD = $(GOBASE)/cmd

build:
	go build -o $(GOBIN)/api  $(GOCMD)/api/main.go

build_cli:
	go build -o $(GOBIN)/cli  $(GOCMD)/cli/main.go

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
#  --operator 1="http://0.0.0.0:8081" --operator 2="http://0.0.0.0:8082" --operator 3="http://0.0.0.0:8083" --operator 4="http://0.0.0.0:8084" --threshold 3 --withdrawal "010000000000000000000000535953b5a6040074948cf185eaa7d2abbd66808f" --fork "prater"
# {
#     "operators": {
#         "1": "http://host.docker.internal:8081",
#         "2": "http://host.docker.internal:8082",
#         "3": "http://host.docker.internal:8083",
#         "4": "http://host.docker.internal:8084"
#     },
#     "threshold": 3,
#     "withdrawal_credentials": "010000000000000000000000535953b5a6040074948cf185eaa7d2abbd66808f",
#     "fork_version": "prater"
# }