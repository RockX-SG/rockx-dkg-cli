package keystore

import (
	"encoding/json"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
)

func ReadKeystoreFromFile(filepath string) (*keystore.Key, error) {
	filedata, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	k := &keystore.Key{}
	return k, json.Unmarshal(filedata, k)
}
