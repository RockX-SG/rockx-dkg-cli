package cli

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/storage"
	"github.com/bloxapp/ssv-spec/types"
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

type KeySharesPayload struct {
	Readable ReadablePayload `json:"readable"`
}

type ReadablePayload struct {
	PublicKey   string   `json:"publicKey"`
	OperatorIDs []uint32 `json:"operatorIds"`
	Shares      string   `json:"shares"`
	Amount      string   `json:"amount"`
	Cluster     string   `json:"cluster"`
}

func (ks *KeyShares) ParseDKGResult(result *DKGResult) error {

	if result.Blame != nil {
		return fmt.Errorf("ParseDKGResult: result contains blame output")
	}

	if len(result.Output) == 0 {
		return fmt.Errorf("ParseDKGResult: dkg result is empty")
	}

	operatorData := make([]OperatorData, 0)
	operatorIds := make([]uint32, 0)

	for operatorID := range result.Output {
		operator, err := storage.GetOperatorFromRegistryByID(operatorID)
		if err != nil {
			return fmt.Errorf("ParseDKGResult: failed to get operator %d from operator registry: %w", operatorID, err)
		}
		operatorData = append(operatorData, OperatorData{
			ID:        uint32(operatorID),
			PublicKey: operator.PublicKey,
		})

		operatorIds = append(operatorIds, uint32(operatorID))
	}

	shares := KeySharesKeys{
		PublicKeys:    make([]string, 0),
		EncryptedKeys: make([]string, 0),
	}

	for _, output := range result.Output {
		shares.PublicKeys = append(shares.PublicKeys, fmt.Sprintf("0x%s", output.Data.SharePubKey))
		shares.EncryptedKeys = append(shares.EncryptedKeys, output.Data.EncryptedShare)
	}

	data := KeySharesData{
		PublicKey: "0x" + result.Output[types.OperatorID(operatorIds[0])].Data.ValidatorPubKey,
		Operators: operatorData,
		Shares:    shares,
	}

	payload := KeySharesPayload{
		Readable: ReadablePayload{
			PublicKey:   "0x" + result.Output[types.OperatorID(operatorIds[0])].Data.ValidatorPubKey,
			OperatorIDs: operatorIds,
			Shares:      sharesToBytes(data.Shares.PublicKeys, shares.EncryptedKeys),
			Amount:      "Amount of SSV tokens to be deposited to your validator's cluster balance (mandatory only for 1st validator in a cluster)",
			Cluster:     "The latest cluster snapshot data, obtained using the cluster-scanner tool. If this is the cluster's 1st validator then use - {0,0,0,0,0,false}",
		},
	}

	ks.Version = "v3"
	ks.Data = data
	ks.Payload = payload
	ks.CreatedAt = time.Now().UTC()
	return nil
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
