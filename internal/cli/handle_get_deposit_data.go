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
	"encoding/hex"
	"fmt"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/utils"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/urfave/cli/v2"
)

type DepositDataJson struct {
	PubKey                string      `json:"pubkey"`
	WithdrawalCredentials string      `json:"withdrawal_credentials"`
	Amount                phase0.Gwei `json:"amount"`
	Signature             string      `json:"signature"`
	DepositMessageRoot    string      `json:"deposit_message_root"`
	DepositDataRoot       string      `json:"deposit_data_root"`
	ForkVersion           string      `json:"fork_version"`
	NetworkName           string      `json:"network_name"`
	DepositCliVersion     string      `json:"deposit_cli_version"`
}

func (h *CliHandler) HandleGetDepositData(c *cli.Context) error {
	requestID := c.String("request-id")

	results, err := h.DKGResultByRequestID(requestID)
	if err != nil {
		return fmt.Errorf("HandleGetDepositData: failed to get dkg result for requestID %s: %w", requestID, err)
	}

	// all operators will have same validatorPK in their result
	var firstOperator types.OperatorID
	for k := range results.Output {
		firstOperator = k
		break
	}

	validatorPK, _ := hex.DecodeString(results.Output[firstOperator].Data.ValidatorPubKey)
	withdrawalCredentials, _ := hex.DecodeString(c.String("withdrawal-credentials"))
	fork := types.NetworkFromString(c.String("fork-version")).ForkVersion()
	amount := phase0.Gwei(types.MaxEffectiveBalanceInGwei)

	_, depositData, err := types.GenerateETHDepositData(validatorPK, withdrawalCredentials, fork, types.DomainDeposit)
	if err != nil {
		return fmt.Errorf("HandleGetDepositData: failed to generate eth deposit data: %w", err)
	}

	depositMsg := &phase0.DepositMessage{
		PublicKey:             depositData.PublicKey,
		WithdrawalCredentials: withdrawalCredentials,
		Amount:                amount,
	}
	depositMsgRoot, _ := depositMsg.HashTreeRoot()

	blsSigBytes, _ := hex.DecodeString(results.Output[firstOperator].Data.DepositDataSignature)
	blsSig := phase0.BLSSignature{}
	copy(blsSig[:], blsSigBytes)
	depositData.Signature = blsSig

	depositDataRoot, _ := depositData.HashTreeRoot()

	depositDataJson := DepositDataJson{
		PubKey:                results.Output[firstOperator].Data.ValidatorPubKey,
		WithdrawalCredentials: c.String("withdrawal-credentials"),
		Amount:                amount,
		Signature:             results.Output[firstOperator].Data.DepositDataSignature,
		DepositMessageRoot:    hex.EncodeToString(depositMsgRoot[:]),
		DepositDataRoot:       hex.EncodeToString(depositDataRoot[:]),
		ForkVersion:           hex.EncodeToString(fork[:]),
		NetworkName:           c.String("fork-version"),
		DepositCliVersion:     "2.3.0",
	}

	filepath := fmt.Sprintf("deposit-data_%d.json", time.Now().UTC().Unix())
	fmt.Printf("writing deposit data json to file %s\n", filepath)
	return utils.WriteJSON(filepath, []DepositDataJson{depositDataJson})
}
