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
	"github.com/urfave/cli/v2"
)

func (h *CliHandler) HandleGetKeyShares(c *cli.Context) error {
	keygenRequestID := c.String("request-id")

	keygenOutput, err := h.DKGResultByRequestID(keygenRequestID)
	if err != nil {
		return fmt.Errorf("HandleGetKeyShares: failed to get dkg result for requestID %s: %w", keygenRequestID, err)
	}

	vk, err := keygenOutput.GetValidatorPK()
	if err != nil {
		return fmt.Errorf("HandleGetKeyShares: failed to get ValidatorPK from keygen results: %w", err)
	}

	ownerAddress := c.String("owner-address")
	ownerNonce := c.Int("owner-nonce")
	signingRoot := []byte(fmt.Sprintf("%s:%d", ownerAddress, ownerNonce))

	signatureRequestID, err := h.GenerateSignature(c, vk, signingRoot)
	if err != nil {
		return fmt.Errorf("HandleGetKeyShares: failed to send signingRoot for signature: %w", err)
	}

	var signatureResult *DKGResult

	try := 0
	sleepTime := 2

	for {
		if try == 4 {
			break
		}

		signatureResult, err = h.DKGResultByRequestID(hex.EncodeToString(signatureRequestID[:]))
		if err != nil {
			time.Sleep(time.Duration(sleepTime) * time.Second)
			try++
			continue
		} else {
			break
		}
	}

	if signatureResult == nil {
		return fmt.Errorf("HandleGetKeyShares: failed to sign owner prefix: %w", err)
	}

	ownerPrefix, err := signatureResult.GetSignatureFromKeySign()
	if err != nil {
		return fmt.Errorf("HandleGetKeyShares: failed to parse owner prefix from signature result: %w", err)
	}

	keyshares := &KeyShares{}
	if err := keyshares.GenerateKeyshareV4(keygenOutput, ownerPrefix); err != nil {
		return fmt.Errorf("HandleGetKeyShares: failed to parse keyshare from dkg results: %w", err)
	}

	filename := fmt.Sprintf("keyshares-%d.json", time.Now().Unix())
	fmt.Printf("writing keyshares to file: %s\n", filename)
	return utils.WriteJSON(filename, keyshares)
}
