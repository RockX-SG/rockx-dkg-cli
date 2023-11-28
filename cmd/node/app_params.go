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

package main

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	"github.com/RockX-SG/frost-dkg-demo/internal/utils"
	"github.com/bloxapp/ssv-spec/types"
)

type AppParams struct {
	HttpAddress        string
	OperatorID         types.OperatorID
	OperatorPrivateKey *rsa.PrivateKey
}

func (params *AppParams) loadFromEnv() error {
	params.loadOperatorID()
	params.loadHttpAddress()
	return params.loadOperatorPrivateKey()
}

func (params *AppParams) print() string {
	return fmt.Sprintf(
		"operatorID=%d http_addr=%s",
		params.OperatorID,
		params.HttpAddress,
	)
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

func (params *AppParams) loadOperatorPrivateKey() error {
	passwordFilePath := os.Getenv("OPERATOR_PRIVATE_KEY_PASSWORD_PATH")
	if passwordFilePath == "" {
		encodedKey := os.Getenv("OPERATOR_PRIVATE_KEY")
		if encodedKey == "" {
			return fmt.Errorf("missing operator private key in app env")
		}
		decodedKey, err := base64.StdEncoding.DecodeString(encodedKey)
		if err != nil {
			return fmt.Errorf("failed to decode base64 encoded operator private key: %w", err)
		}
		operatorPrivateKey, err := types.PemToPrivateKey(decodedKey)
		if err != nil {
			return fmt.Errorf("failed to convert pem block to rsa private key %w", err)
		}
		params.OperatorPrivateKey = operatorPrivateKey
	} else {
		lockedPrivateKey, err := os.ReadFile(os.Getenv("OPERATOR_PRIVATE_KEY_PATH"))
		if err != nil {
			return fmt.Errorf("failed to read operator private key file %w", err)
		}
		keyPassword, err := os.ReadFile(passwordFilePath)
		if err != nil {
			return fmt.Errorf("failed to read operator private key password file %w", err)
		}
		privateKey, err := utils.UnlockRSAJSON(lockedPrivateKey, string(keyPassword))
		if err != nil {
			return fmt.Errorf("failed to unlock operator private key %w", err)
		}
		params.OperatorPrivateKey = privateKey
	}
	return nil
}
