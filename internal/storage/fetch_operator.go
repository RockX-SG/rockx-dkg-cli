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

package storage

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
)

func FetchOperatorByID(operatorID types.OperatorID) (*dkg.Operator, error) {
	// Note, this is just for testing, to be removed before moving to staging
	if isUsingHardcodedOperators() {
		return hardCodedOperatorInfo(operatorID)
	}

	operator, err := GetOperatorFromRegistryByID(operatorID)
	if err != nil {
		return nil, err
	}

	publicKey, err := ParsePublicKeyFromBase64(operator.PublicKey)
	if err != nil {
		return nil, err
	}

	return &dkg.Operator{
		OperatorID:       operatorID,
		ETHAddress:       ethAddressFromHex(operator.Owner[2:]),
		EncryptionPubKey: publicKey,
	}, nil
}

type operatorResponse struct {
	ID        uint32 `json:"id"`
	Owner     string `json:"owner_address"`
	PublicKey string `json:"public_key"`
}

func GetOperatorFromRegistryByID(operatorID types.OperatorID) (*operatorResponse, error) {
	var operator = new(operatorResponse)
	respBody, err := getResponse(fmt.Sprintf("https://api.ssv.network/api/v4/prater/operators/%d", operatorID))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(respBody, operator); err != nil {
		return nil, err
	}
	return operator, nil
}

func isUsingHardcodedOperators() bool {
	isHardcoded := os.Getenv("USE_HARDCODED_OPERATORS")
	if isHardcoded == "" {
		isHardcoded = "false"
	}
	return os.Getenv("USE_HARDCODED_OPERATORS") == "true"
}

func getResponse(url string) ([]byte, error) {
	cl := getHttpClient()
	resp, err := cl.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func ParsePublicKeyFromBase64(base64Key string) (*rsa.PublicKey, error) {
	// Decode the Base64-encoded key
	keyBytes, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, err
	}

	// Parse the PEM block
	pemBlock, _ := pem.Decode(keyBytes)
	if pemBlock == nil {
		return nil, errors.New("failed to parse PEM block containing public key")
	}

	// Parse the RSA public key
	publicKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey.(*rsa.PublicKey), nil
}

func PublicKeyToBase64(publicKey *rsa.PublicKey) (string, error) {
	// Marshal the RSA public key to DER format
	derBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	// Create a PEM block
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derBytes,
	}

	// Encode the PEM block to Base64
	base64Encoded := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(pemBlock))

	return base64Encoded, nil
}

func getHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		IdleConnTimeout: 5 * time.Minute, // Close idle connections after 30 seconds
	}

	// Create an HTTP client with the custom transport
	return &http.Client{Transport: tr, Timeout: 5 * time.Minute}
}
