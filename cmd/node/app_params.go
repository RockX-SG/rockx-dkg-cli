package main

import (
	"os"
	"strconv"

	"github.com/bloxapp/ssv-spec/types"
)

type AppParams struct {
	OperatorID           types.OperatorID
	HttpAddress          string
	MessengerHttpAddress string
	KeystoreFilePath     string
	keystorePassword     string
}

func (params *AppParams) loadFromEnv() {
	params.loadOperatorID()
	params.loadHttpAddress()
	params.loadMessengerHttpAddress()
	params.loadKeystoreFilePath()
	params.loadKeystorePassword()
}

func (params *AppParams) loadOperatorID() {
	operatorID, err := strconv.ParseUint(os.Getenv("NODE_OPERATOR_ID"), 10, 32)
	if err != nil {
		panic(err)
	}
	params.OperatorID = types.OperatorID(operatorID)
}

func (params *AppParams) loadHttpAddress() {
	nodeAddr := os.Getenv("NODE_ADDR")
	if nodeAddr == "" {
		nodeAddr = "0.0.0.0:8080"
	}
	params.HttpAddress = nodeAddr
}

func (params *AppParams) loadMessengerHttpAddress() {
	srvaddr := os.Getenv("MESSENGER_SRV_ADDR")
	if srvaddr == "" {
		srvaddr = "http://0.0.0.0:3000"
	}
	params.MessengerHttpAddress = srvaddr
}

func (params *AppParams) loadKeystoreFilePath() {
	keystoreFilePath := os.Getenv("KEYSTORE_FILE_PATH")
	if keystoreFilePath == "" {
		keystoreFilePath = "keystore.json"
	}
	params.KeystoreFilePath = keystoreFilePath
}

func (params *AppParams) loadKeystorePassword() {
	params.keystorePassword = os.Getenv("KEYSTORE_PASSWORD")
}
