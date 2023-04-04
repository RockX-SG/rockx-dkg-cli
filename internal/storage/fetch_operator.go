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

	fmt.Println("owner", operator.Owner)

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
	respBody, err := getResponse(fmt.Sprintf("https://api.ssv.network/api/v3/prater/operators/%d", operatorID))
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

func getHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		IdleConnTimeout: 30 * time.Second, // Close idle connections after 30 seconds
	}

	// Create an HTTP client with the custom transport
	return &http.Client{Transport: tr, Timeout: 10 * time.Second}
}
