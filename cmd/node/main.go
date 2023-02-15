package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/node"
	"github.com/RockX-SG/frost-dkg-demo/internal/ping"
	"github.com/RockX-SG/frost-dkg-demo/internal/storage"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/dkg/frost"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/dgraph-io/badger/v3"

	"github.com/gin-gonic/gin"
)

func main() {
	operatorID, nodeAddr, messengerAddr := loadEnv()
	network := messenger.NewMessengerClient(messengerAddr)
	signer := testingutils.NewTestingKeyManager()
	db, err := setupDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	storage := storage.NewStorage(db)

	config := &dkg.Config{
		KeygenProtocol:      frost.New,
		ReshareProtocol:     frost.NewResharing,
		Network:             network,
		Signer:              signer,
		Storage:             storage,
		SignatureDomainType: types.PrimusTestnet,
	}
	thisNode := dkg.NewNode(thisOperator(uint32(operatorID), storage), config)

	// register with the messenger
	if err := network.RegisterOperatorNode(strconv.Itoa(int(operatorID)), fmt.Sprintf("http://%s", os.Getenv("NODE_BROADCAST_ADDR"))); err != nil {
		panic(err)
	}

	r := gin.Default()
	r.GET("/ping", ping.HandlePing)
	r.POST("/consume", node.HandleConsume(thisNode))
	r.GET("/dkg_results/:vk", node.HandleGetDKGResults(thisNode))

	panic(r.Run(nodeAddr))
}

func loadEnv() (operatorID uint64, nodeAddr, messengerAddr string) {
	nodeAddr = os.Getenv("NODE_ADDR")
	if nodeAddr == "" {
		nodeAddr = "0.0.0.0:8080"
	}

	operatorID, err := strconv.ParseUint(os.Getenv("NODE_OPERATOR_ID"), 10, 32)
	if err != nil {
		panic(err)
	}

	hostname := os.Getenv("MESSENGER_SRV_ADDR")
	if hostname == "" {
		hostname = "http://0.0.0.0:3000"
	}
	port := os.Getenv("MESSENGER_SRV_ADDR_PORT")
	if port == "" {
		port = "3000"
	}

	messengerAddr = fmt.Sprintf("http://%s:%s", hostname, port)
	return
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

func setupDB() (*badger.DB, error) {
	return badger.Open(badger.DefaultOptions("/tmp/frost-dkg-data"))
}
