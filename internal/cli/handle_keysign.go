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

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/urfave/cli/v2"
)

func (h *CliHandler) GenerateSignature(c *cli.Context, vk types.ValidatorPK, signingRoot []byte) (dkg.RequestID, error) {
	requestID := getRandRequestID()

	keySign := dkg.KeySign{
		ValidatorPK: vk,
		SigningRoot: signingRoot,
	}
	keySignBytes, _ := keySign.Encode()

	initBytes, err := initMsgForKeySign(requestID, keySignBytes)
	if err != nil {
		return [24]byte{}, fmt.Errorf("HandleKeySign: failed to generate init msg for KeySign: %w", err)
	}

	operators, err := parseOperatorList(c)
	if err != nil {
		return [24]byte{}, fmt.Errorf("HandleKeySign: failed to parse operator list from command: %w", err)
	}

	ol := make([]types.OperatorID, 0)
	for operatorID := range operators {
		ol = append(ol, operatorID)
	}

	messengerClient := messenger.NewMessengerClient(messenger.MessengerAddrFromEnv())
	if err := messengerClient.CreateTopic(hex.EncodeToString(requestID[:]), ol); err != nil {
		return [24]byte{}, fmt.Errorf("HandleKeygen: failed to create a new topic on messenger service: %w", err)
	}

	for operatorID, addr := range operators {
		if err := h.sendKeySignMsg(operatorID, addr, initBytes); err != nil {
			return [24]byte{}, fmt.Errorf("HandleKeySign: failed to send init message to operatorID %d: %w", operatorID, err)
		}
	}

	return requestID, nil
}

func (h *CliHandler) sendKeySignMsg(operatorID types.OperatorID, addr string, data []byte) error {
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

func initMsgForKeySign(requestID dkg.RequestID, data []byte) ([]byte, error) {
	ks := testingutils.TestingKeygenKeySet()
	signedInitMsg := testingutils.SignDKGMsg(ks.DKGOperators[1].EncryptionKey, 1, &dkg.Message{
		MsgType:    dkg.KeySignMsgType,
		Identifier: requestID,
		Data:       data,
	})
	signedInitMsgBytes, _ := signedInitMsg.Encode()

	msg := &types.SSVMessage{
		MsgType: types.DKGMsgType,
		Data:    signedInitMsgBytes,
	}
	return msg.Encode()
}
