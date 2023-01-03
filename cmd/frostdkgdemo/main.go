package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

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

	log.Println(hex.EncodeToString(requestID[:]))

	for _, operatorID := range operators {
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
		msgBytes, _ := msg.Encode()

		resp, err := http.Post(fmt.Sprintf("%s/consume", nodes[operatorID]), "application/json", bytes.NewReader(msgBytes))
		if err != nil {
			panic(err)
		}

		fmt.Println(resp.StatusCode)
		resp.Body.Close()
	}
}
