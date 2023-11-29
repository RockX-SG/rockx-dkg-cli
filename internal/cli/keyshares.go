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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/storage"
	"github.com/bloxapp/ssv-spec/types"
)

type KeyShares struct {
	Version   string        `json:"version"`
	Data      KeySharesData `json:"data"`
	Payload   Payload       `json:"payload"`
	CreatedAt time.Time     `json:"createdAt"`
}

type KeySharesData struct {
	PublicKey string         `json:"publicKey"`
	Operators []OperatorData `json:"operators"`
}

type OperatorData struct {
	ID          uint32 `json:"id"`
	OperatorKey string `json:"operatorKey"`
}

type KeySharesKeys struct {
	PublicKeys    []string `json:"publicKeys"`
	EncryptedKeys []string `json:"encryptedKeys"`
}

type Payload struct {
	PublicKey   string   `json:"publicKey"`
	OperatorIDs []uint32 `json:"operatorIds"`
	Shares      string   `json:"sharesData"`
	Owner       string   `json:"ownerAddress"`
	Nonce       int      `json:"ownerNonce"`
}

func (ks *KeyShares) GenerateKeyshareV4(result *DKGResult, ownerSig, ownerAddress string, ownerNonce int) error {

	if result.Blame != nil {
		return fmt.Errorf("ParseDKGResultV4: result contains blame output")
	}

	if len(result.Output) == 0 {
		return fmt.Errorf("ParseDKGResultV4: dkg result is empty")
	}

	operatorData := make([]OperatorData, 0)
	operatorIds := make([]uint32, 0)

	for operatorID := range result.Output {
		operator, err := storage.FetchOperatorByID(operatorID)
		if err != nil {
			return fmt.Errorf("ParseDKGResultV4: failed to get operator %d from operator registry: %w", operatorID, err)
		}

		publicKey, _ := storage.PublicKeyToBase64(operator.EncryptionPubKey)

		operatorData = append(operatorData, OperatorData{
			ID:          uint32(operatorID),
			OperatorKey: publicKey,
		})

		operatorIds = append(operatorIds, uint32(operatorID))
	}

	sort.SliceStable(operatorIds, func(i, j int) bool {
		return operatorIds[i] < operatorIds[j]
	})

	sort.SliceStable(operatorData, func(i, j int) bool {
		return operatorData[i].ID < operatorData[j].ID
	})

	shares := KeySharesKeys{
		PublicKeys:    make([]string, 0),
		EncryptedKeys: make([]string, 0),
	}

	for _, id := range operatorIds {
		output := result.Output[types.OperatorID(id)]
		shares.PublicKeys = append(shares.PublicKeys, "0x"+output.Data.SharePubKey)
		encryptedShare, _ := hex.DecodeString(output.Data.EncryptedShare)
		shares.EncryptedKeys = append(shares.EncryptedKeys, base64.StdEncoding.EncodeToString(encryptedShare))
	}

	data := KeySharesData{
		PublicKey: "0x" + result.Output[types.OperatorID(operatorIds[0])].Data.ValidatorPubKey,
		Operators: operatorData,
	}

	payload := Payload{
		PublicKey:   "0x" + result.Output[types.OperatorID(operatorIds[0])].Data.ValidatorPubKey,
		OperatorIDs: operatorIds,
		Shares:      sharesToBytes(shares.PublicKeys, shares.EncryptedKeys, ownerSig),
		Owner:       ownerAddress,
		Nonce:       ownerNonce,
	}

	ks.Version = "v4"
	ks.Data = data
	ks.Payload = payload
	ks.CreatedAt = time.Now().UTC()
	return nil
}

func sharesToBytes(publicKeys []string, privateKeys []string, ownerSig string) string {
	encryptedShares, _ := decodeEncryptedShares(privateKeys)
	arrayPublicKeys := bytes.Join(toArrayByteSlices(publicKeys), []byte{})
	arrayEncryptedShares := bytes.Join(toArrayByteSlices(encryptedShares), []byte{})
	pkPsBytes := append(arrayPublicKeys, arrayEncryptedShares...)
	return "0x" + ownerSig + hex.EncodeToString(pkPsBytes)
}

func decodeEncryptedShares(encodedEncryptedShares []string) ([]string, error) {
	var result []string
	for _, item := range encodedEncryptedShares {
		decoded, err := base64.StdEncoding.DecodeString(item)
		if err != nil {
			return nil, err
		}
		result = append(result, "0x"+hex.EncodeToString(decoded))
	}
	return result, nil
}

func toArrayByteSlices(input []string) [][]byte {
	var result [][]byte
	for _, str := range input {
		bytes, _ := hex.DecodeString(str[2:]) // remove the '0x' prefix and decode the hex string to bytes
		result = append(result, bytes)
	}
	return result
}
