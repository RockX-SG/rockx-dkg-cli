package cli

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/utils"
	"github.com/urfave/cli/v2"
)

type KeyShares struct {
	Version   string           `json:"version"`
	Data      KeySharesData    `json:"data"`
	Payload   KeySharesPayload `json:"payload"`
	CreatedAt time.Time        `json:"createdAt"`
}

type KeySharesData struct {
	PublicKey string         `json:"publicKey"`
	Operators []OperatorData `json:"operators"`
	Shares    KeySharesKeys  `json:"shares"`
}

type OperatorData struct {
	ID        uint32 `json:"id"`
	PublicKey string `json:"publicKey"`
}

type KeySharesKeys struct {
	PublicKeys    []string `json:"publicKeys"`
	EncryptedKeys []string `json:"encryptedKeys"`
}

type ReadablePayload struct {
	PublicKey   string   `json:"publicKey"`
	OperatorIDs []uint32 `json:"operatorIds"`
	Shares      string   `json:"shares"`
	Amount      string   `json:"amount"`
	Cluster     string   `json:"cluster"`
}

type KeySharesPayload struct {
	Readable ReadablePayload `json:"readable"`
}

func (h *CliHandler) HandleGetKeyShares(c *cli.Context) error {
	requestID := c.String("request-id")

	results, err := h.DKGResultByRequestID(requestID)
	if err != nil {
		return err
	}

	keyshares, err := results.toKeyShares()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("keyshares-%d.json", time.Now().Unix())
	fmt.Printf("writing keyshares to file: %s\n", filename)
	return utils.WriteJSON(filename, keyshares)
}

// Convert a slice of strings to a slice of byte slices, where each string is converted to a byte slice
// using hex decoding
func toArrayByteSlices(input []string) [][]byte {
	var result [][]byte
	for _, str := range input {
		bytes, _ := hex.DecodeString(str[2:]) // remove the '0x' prefix and decode the hex string to bytes
		result = append(result, bytes)
	}
	return result
}

func sharesToBytes(publicKeys []string, privateKeys []string) string {
	encryptedShares, _ := decodeEncryptedShares(privateKeys)
	arrayPublicKeys := bytes.Join(toArrayByteSlices(publicKeys), []byte{})
	arrayEncryptedShares := bytes.Join(toArrayByteSlices(encryptedShares), []byte{})

	// public keys hex encoded
	pkHex := hex.EncodeToString(arrayPublicKeys)
	// length of the public keys (hex), hex encoded
	pkHexLength := fmt.Sprintf("%04x", len(pkHex)/2)

	// join arrays
	pkPsBytes := append(arrayPublicKeys, arrayEncryptedShares...)

	// add length of the public keys at the beginning
	// this is the variable that is sent to the contract as bytes, prefixed with 0x
	return "0x" + pkHexLength + hex.EncodeToString(pkPsBytes)
}

func decodeEncryptedShares(encodedEncryptedShares []string) ([]string, error) {
	var result []string
	for _, item := range encodedEncryptedShares {
		// Decode the base64 string
		decoded, err := base64.StdEncoding.DecodeString(item)
		if err != nil {
			return nil, err
		}

		// Encode the decoded bytes as a hexadecimal string with '0x' prefix
		result = append(result, "0x"+hex.EncodeToString(decoded))
	}
	return result, nil
}
