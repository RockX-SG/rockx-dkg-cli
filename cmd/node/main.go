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
	log := logger.New()

	params := &AppParams{}
	params.loadFromEnv()

	log.Infof("app env: %s messenger: %s", params.print(), messenger.MessengerAddrFromEnv())

	// set up db for storage
	db, err := setupDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// TODO: add a check to verify the node operator is a valid node operator
	operatorPrivateKey, err := params.loadDecryptedPrivateKey()
	if err != nil {
		panic(err)
	}

	storage := store.NewStorage(db)
	network := messenger.NewMessengerClient(messenger.MessengerAddrFromEnv())
	signer := keymanager.NewKeyManager(types.PrimusTestnet, operatorPrivateKey)

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
	if err := network.RegisterOperatorNode(strconv.Itoa(int(params.OperatorID)), os.Getenv("NODE_BROADCAST_ADDR")); err != nil {
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
