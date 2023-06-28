/*
 * ==================================================================
 *Copyright (C) 2022-2023 Altstake Technology Pte. Ltd. (RockX)
 *This file is part of rockx-dkg-cli <https://github.com/RockX-SG/rockx-dkg-cli>
 *CAUTION: THESE CODES HAVE NOT BEEN AUDITED
 *
 *rockx-dkg-cli is free software: you can redistribute it and/or modify
 *it under the terms of the GNU General Public License as published by
 *the Free Software Foundation, either version 3 of the License, or
 *(at your option) any later version.
 *
 *rockx-dkg-cli is distributed in the hope that it will be useful,
 *but WITHOUT ANY WARRANTY; without even the implied warranty of
 *MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *GNU General Public License for more details.
 *
 *You should have received a copy of the GNU General Public License
 *along with rockx-dkg-cli. If not, see <http://www.gnu.org/licenses/>.
 *==================================================================
 */

package cli

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/urfave/cli/v2"
)

func (h *CliHandler) HandleKeygen(c *cli.Context) error {
	keygenRequest := &KeygenRequest{}
	if err := keygenRequest.parseKeygenRequest(c); err != nil {
		return fmt.Errorf("HandleKeygen: failed to parse keygen request: %w", err)
	}

	requestID := getRandRequestID()
	requestIDInHex := hex.EncodeToString(requestID[:])

	messengerClient := messenger.NewMessengerClient(messenger.MessengerAddrFromEnv())
	if err := messengerClient.CreateTopic(requestIDInHex, keygenRequest.allOperators()); err != nil {
		return fmt.Errorf("HandleKeygen: failed to create a new topic on messenger service: %w", err)
	}

	initMsgBytes, err := keygenRequest.initMsgForKeygen(requestID)
	if err != nil {
		return fmt.Errorf("HandleKeygen: failed to generate init message for keygen: %w", err)
	}

	for operatorID, nodeAddr := range keygenRequest.Operators {
		if err := h.sendInitMsg(operatorID, nodeAddr, initMsgBytes); err != nil {
			return fmt.Errorf("HandleKeygen: failed to send init message to operatorID %d: %w", operatorID, err)
		}
	}

	fmt.Printf("keygen init request sent with ID: %s\n", requestIDInHex)
	return nil
}

func (h *CliHandler) sendInitMsg(operatorID types.OperatorID, addr string, data []byte) error {
	url := fmt.Sprintf("%s/consume", addr)
	resp, err := h.client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request to operator %d to consume init message failed with status %s", operatorID, resp.Status)
	}
	return nil
}

type KeygenRequest struct {
	Operators            map[types.OperatorID]string `json:"operators"`
	Threshold            int                         `json:"threshold"`
	WithdrawalCredential string                      `json:"withdrawal_credentials"`
	ForkVersion          string                      `json:"fork_version"`
}

func (request *KeygenRequest) allOperators() []types.OperatorID {
	operators := []types.OperatorID{}
	for operatorID := range request.Operators {
		operators = append(operators, operatorID)
	}
	return operators
}

func (request *KeygenRequest) parseKeygenRequest(c *cli.Context) error {
	operators, err := parseOperatorList(c)
	if err != nil {
		return err
	}

	request.Operators = operators
	request.Threshold = c.Int("threshold")
	request.WithdrawalCredential = c.String("withdrawal-credentials")
	request.ForkVersion = c.String("fork-version")
	return nil
}

func parseOperatorList(c *cli.Context) (map[types.OperatorID]string, error) {
	operators := make(map[types.OperatorID]string)
	for _, o := range c.StringSlice("operator") {

		operator := strings.Trim(o, " ")

		pair := strings.Split(operator, "=")
		if len(pair) != 2 {
			return nil, fmt.Errorf("operator %s is not in the form of key=value", operator)
		}

		operatorID, err := strconv.Atoi(pair[0])
		if err != nil {
			return nil, err
		}
		operators[types.OperatorID(operatorID)] = pair[1]
	}
	return operators, nil
}

func (request *KeygenRequest) initMsgForKeygen(requestID dkg.RequestID) ([]byte, error) {
	withdrawalCred, _ := hex.DecodeString(request.WithdrawalCredential)
	forkVersion := types.NetworkFromString(request.ForkVersion).ForkVersion()

	init := testingutils.InitMessageData(
		request.allOperators(),
		uint16(request.Threshold),
		withdrawalCred,
		forkVersion,
	)
	initBytes, _ := init.Encode()

	// TODO: TBD who signs this init msg
	ks := testingutils.TestingKeygenKeySet()
	signedInitMsg := testingutils.SignDKGMsg(ks.DKGOperators[1].SK, 1, &dkg.Message{
		MsgType:    dkg.InitMsgType,
		Identifier: requestID,
		Data:       initBytes,
	})
	signedInitMsgBytes, _ := signedInitMsg.Encode()

	msg := &types.SSVMessage{
		MsgType: types.DKGMsgType,
		Data:    signedInitMsgBytes,
	}
	return msg.Encode()
}
