package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/RockX-SG/frost-dkg-demo/internal/keymanager"
	"github.com/RockX-SG/frost-dkg-demo/internal/keystore"
	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/node"
	"github.com/RockX-SG/frost-dkg-demo/internal/ping"
	store "github.com/RockX-SG/frost-dkg-demo/internal/storage"
	"github.com/ethereum/go-ethereum/crypto"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/dkg/frost"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/dgraph-io/badger/v3"
	"github.com/gin-gonic/gin"
)

func init() {
	types.InitBLS()
}

func main() {
	params := &AppParams{}
	params.loadFromEnv()

	// set up db for storage
	db, err := setupDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	f, err := os.Open(params.KeystoreFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	sk, err := SKFromFile(params.KeystoreFilePath, params.keystorePassword)
	if err != nil {
		panic(err)
	}

	storage := store.NewStorage(db)
	network := messenger.NewMessengerClient(params.MessengerHttpAddress)
	signer := keymanager.NewKeyManager(types.PrimusTestnet, sk)

	config := &dkg.Config{
		KeygenProtocol:      frost.New,
		ReshareProtocol:     frost.NewResharing,
		Network:             network,
		Signer:              signer,
		Storage:             storage,
		SignatureDomainType: types.PrimusTestnet,
	}
	dkgnode := dkg.NewNode(thisOperator(uint32(params.OperatorID), storage), config)

	// register dkg operator node with the messenger
	if err := network.RegisterOperatorNode(strconv.Itoa(int(params.OperatorID)), fmt.Sprintf("http://%s", os.Getenv("NODE_BROADCAST_ADDR"))); err != nil {
		panic(err)
	}

	// register api routes
	r := gin.Default()
	r.GET("/ping", ping.HandlePing)

	// handle incoming message
	r.POST("/consume", node.HandleConsume(dkgnode))

	// get dkg results
	r.GET("/dkg_results/:vk", node.HandleGetDKGResults(dkgnode))

	panic(r.Run(params.HttpAddress))
}

func setupDB() (*badger.DB, error) {
	return badger.Open(badger.DefaultOptions("/frost-dkg-data"))
}

func thisOperator(operatorID uint32, storage dkg.Storage) *dkg.Operator {
	exist, operator, err := storage.GetDKGOperator(types.OperatorID(operatorID))
	if err != nil {
		panic(err)
	}
	if !exist {
		panic(fmt.Sprintf("operator with ID %d doesn't exist", operatorID))
	}
	return operator
}

func SKFromFile(filepath, password string) (*ecdsa.PrivateKey, error) {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	e := keystorev4.New(keystorev4.WithCipher("scrypt"))

	data := keystore.KeyStoreV4{}
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		panic(err)
	}

	data2 := make(map[string]interface{})
	data2["checksum"] = data.Crypto.Checksum
	data2["cipher"] = data.Crypto.Cipher
	data2["kdf"] = data.Crypto.KDF

	key, err := e.Decrypt(data2, password)
	if err != nil {
		panic(err)
	}
	return crypto.ToECDSA(key)
}
