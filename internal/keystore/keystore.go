package keystore

import (
	"crypto/ecdsa"
	"encoding/json"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
)

type KeyStoreV4 struct {
	Crypto      CryptoInfo `json:"crypto"`
	Description string     `json:"description"`
	PubKey      string     `json:"pubkey"`
	Path        string     `json:"path"`
	UUID        string     `json:"uuid"`
	Version     int        `json:"version"`
}

type CryptoInfo struct {
	KDF      KDFInfo      `json:"kdf"`
	Checksum ChecksumInfo `json:"checksum"`
	Cipher   CipherInfo   `json:"cipher"`
}

type KDFInfo struct {
	Function string    `json:"function"`
	Params   KDFParams `json:"params"`
	Message  string    `json:"message"`
}

type KDFParams struct {
	DKLen int    `json:"dklen"`
	N     int    `json:"n"`
	R     int    `json:"r"`
	P     int    `json:"p"`
	Salt  string `json:"salt"`
}

type ChecksumInfo struct {
	Function string   `json:"function"`
	Params   struct{} `json:"params"`
	Message  string   `json:"message"`
}

type CipherInfo struct {
	Function string       `json:"function"`
	Params   CipherParams `json:"params"`
	Message  string       `json:"message"`
}

type CipherParams struct {
	IV string `json:"iv"`
}

func (data *KeyStoreV4) DecodeFromFile(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&data)
}

func (data *KeyStoreV4) ToECDSAPrivateKey(password string) (*ecdsa.PrivateKey, error) {
	key, err := keystorev4.New(keystorev4.WithCipher("scrypt")).
		Decrypt(
			map[string]interface{}{
				"checksum": data.Crypto.Checksum,
				"cipher":   data.Crypto.Cipher,
				"kdf":      data.Crypto.KDF,
			},
			password,
		)

	if err != nil {
		return nil, err
	}

	return crypto.ToECDSA(key)
}
