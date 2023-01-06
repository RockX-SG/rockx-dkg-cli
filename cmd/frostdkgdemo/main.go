package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv-spec/types/testingutils"
)

func main() {
	nodes := make(map[types.OperatorID]string)
	nodes[1] = "http://0.0.0.0:8081"
	nodes[2] = "http://0.0.0.0:8082"
	nodes[3] = "http://0.0.0.0:8083"
	nodes[4] = "http://0.0.0.0:8084"

	operators := []types.OperatorID{1, 2, 3, 4}
	threshold := 3
	ks := testingutils.TestingKeygenKeySet()
	requestID := testingutils.GetRandRequestID()

	log.Printf("RequestID: %s\n", hex.EncodeToString(requestID[:]))

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		DisableKeepAlives:   true,
	}

	netClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: netTransport,
	}

	for _, operatorID := range operators {
		nodeAddr := nodes[operatorID]

		init := testingutils.InitMessageData(
			operators,
			uint16(threshold),
			testingutils.TestingWithdrawalCredentials,
			testingutils.TestingForkVersion,
		)
		initBytes, _ := init.Encode()

		initMsg := testingutils.SignDKGMsg(ks.DKGOperators[operatorID].SK, operatorID, &dkg.Message{
			MsgType:    dkg.InitMsgType,
			Identifier: requestID,
			Data:       initBytes,
		})
		byts, _ := initMsg.Encode()

		msg := &types.SSVMessage{
			MsgType: types.DKGMsgType,
			Data:    byts,
		}

		msgBytes, err := msg.Encode()
		if err != nil {
			panic(err)
		}

		url := fmt.Sprintf("%s/consume", nodeAddr)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(msgBytes))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")

		resp, err := netClient.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	}
}
