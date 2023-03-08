package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/RockX-SG/frost-dkg-demo/internal/keymanager"
	"github.com/RockX-SG/frost-dkg-demo/internal/keystore"
	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/node"
	"github.com/RockX-SG/frost-dkg-demo/internal/ping"
	store "github.com/RockX-SG/frost-dkg-demo/internal/storage"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"

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
	log := setupLogger()

	params := &AppParams{}
	params.loadFromEnv()

	log.Infof("app env: %s", params.print())

	// set up db for storage
	db, err := setupDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ks := &keystore.KeyStoreV4{}
	if err := ks.DecodeFromFile(params.KeystoreFilePath); err != nil {
		panic(err)
	}

	sk, err := ks.ToECDSAPrivateKey(params.keystorePassword)
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

func setupLogger() *logrus.Logger {
	// Create a new Logrus logger instance
	logger := logrus.New()

	// Set the log level to Info
	logger.SetLevel(logrus.InfoLevel)

	// Create a new LFS hook to write log messages to a file
	logFilePath := "/var/log/dkg_node.log"
	fileHook := lfshook.NewHook(lfshook.PathMap{
		logrus.InfoLevel:  logFilePath,
		logrus.WarnLevel:  logFilePath,
		logrus.ErrorLevel: logFilePath,
	}, &logrus.JSONFormatter{})

	// Add the LFS hook to the logger
	logger.AddHook(fileHook)

	// Set the logger to not print to the console
	logger.SetOutput(os.Stdout)

	return logger
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
