package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/RockX-SG/frost-dkg-demo/internal/handlers"
	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/dkg/frost"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/ping", handlers.HandlePing)

	storage := testingutils.NewTestingStorage()
	operatorID, err := strconv.ParseInt(os.Getenv("NODE_OPERATOR_ID"), 10, 32)
	if err != nil {
		panic(err)
	}
	exist, operator, err := storage.GetDKGOperator(types.OperatorID(operatorID))
	if err != nil {
		panic(err)
	}
	if !exist {
		panic(fmt.Sprintf("operator with ID %d doesn't exist", operatorID))
	}

	messengerSrvAddr := os.Getenv("MESSENGER_SRV_ADDR")
	if messengerSrvAddr == "" {
		messengerSrvAddr = "http://0.0.0.0:3000"
	}

	signer := testingutils.NewTestingKeyManager()
	network := messenger.NewMessengerClient(messengerSrvAddr)

	config := &dkg.Config{
		KeygenProtocol:  frost.New,
		ReshareProtocol: frost.NewResharing,
		Network:         network,
		Signer:          signer,
		Storage:         storage,
	}
	node := dkg.NewNode(operator, config)
	SetRoutes(r.Group(""), node)

	HttpAddr := os.Getenv("NODE_ADDR")
	if HttpAddr == "" {
		HttpAddr = "0.0.0.0:8080"
	}
	panic(r.Run(HttpAddr))
}

func SetRoutes(r *gin.RouterGroup, node *dkg.Node) {
	r.POST("/consume", handlers.HandleConsume(node))
}
