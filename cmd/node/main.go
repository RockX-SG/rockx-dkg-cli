package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/RockX-SG/frost-dkg-demo/internal/keymanager"
	"github.com/RockX-SG/frost-dkg-demo/internal/logger"
	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/node"
	"github.com/RockX-SG/frost-dkg-demo/internal/ping"
	store "github.com/RockX-SG/frost-dkg-demo/internal/storage"

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
	log := logger.New("/var/log/dkg_node.log")

	params := &AppParams{}
	params.loadFromEnv()
	log.Debugf("app env: %s messenger: %s", params.print(), messenger.MessengerAddrFromEnv())

	// set up db for storage
	db, err := setupDB()
	if err != nil {
		log.Errorf("Main: failed to setup DB: %w", err)
		panic(err)
	}
	defer db.Close()
	storage := store.NewStorage(db)

	// TODO: add a check to verify the node operator is a valid node operator
	operatorPrivateKey, err := params.loadDecryptedPrivateKey()
	if err != nil {
		log.Errorf("Main: failed to load decrypted private key: %w", err)
		panic(err)
	}
	signer := keymanager.NewKeyManager(types.PrimusTestnet, operatorPrivateKey)

	network := messenger.NewMessengerClient(messenger.MessengerAddrFromEnv())

	config := &dkg.Config{
		KeygenProtocol:      frost.New,
		ReshareProtocol:     frost.NewResharing,
		Network:             network,
		Signer:              signer,
		Storage:             storage,
		SignatureDomainType: types.PrimusTestnet,
	}

	thisOperator, err := thisOperator(uint32(params.OperatorID), storage)
	if err != nil {
		log.Errorf("Main: failed to get operator %d from operator registry: %w", err)
		panic(err)
	}
	dkgnode := dkg.NewNode(thisOperator, config)

	// register dkg operator node with the messenger
	if err := network.RegisterOperatorNode(strconv.Itoa(int(params.OperatorID)), os.Getenv("NODE_BROADCAST_ADDR")); err != nil {
		log.Errorf("Main: %w", err)
		panic(err)
	}

	h := node.New(log)

	// register api routes
	r := gin.Default()
	r.Use(logger.GinLogger(log))

	r.GET("/ping", ping.HandlePing)

	// handle incoming message
	r.POST("/consume", h.HandleConsume(dkgnode))

	// get dkg results
	r.GET("/dkg_results/:vk", h.HandleGetDKGResults(dkgnode))

	panic(r.Run(params.HttpAddress))
}

func setupDB() (*badger.DB, error) {
	return badger.Open(badger.DefaultOptions("/frost-dkg-data"))
}

func thisOperator(operatorID uint32, storage dkg.Storage) (*dkg.Operator, error) {
	exist, operator, err := storage.GetDKGOperator(types.OperatorID(operatorID))
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("operator with ID %d doesn't exist", operatorID)
	}
	return operator, nil
}
