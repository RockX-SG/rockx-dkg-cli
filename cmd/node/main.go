/*
 * ==================================================================
 *Copyright (C) 2022-2023 Altstake Technology Pte. Ltd. (RockX)
 *This file is part of rockx-dkg-cli <https://github.com/RockX-SG/rockx-dkg-cli>
 *CAUTION: THESE CODES HAVE NOT BEEN AUDITED
 *
 *rockx-dkg-cli is free software: you can redistribute it and/or modify
 *it under the terms of the GNU General Public License as published by
 *the Free Software Foundation, either version 3 of the License, or
 *(at your option) any later version.
 *
 *rockx-dkg-cli is distributed in the hope that it will be useful,
 *but WITHOUT ANY WARRANTY; without even the implied warranty of
 *MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *GNU General Public License for more details.
 *
 *You should have received a copy of the GNU General Public License
 *along with rockx-dkg-cli. If not, see <http://www.gnu.org/licenses/>.
 *==================================================================
 */

package main

import (
	"fmt"
	"net/http"
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
	"github.com/bloxapp/ssv-spec/dkg/keysign"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/dgraph-io/badger/v3"
	"github.com/gin-gonic/gin"
)

const serviceName = "node"

var version string

func init() {
	types.InitBLS()
}

func main() {
	log := logger.New(serviceName)

	params := &AppParams{}
	params.loadFromEnv()

	log.Debugf("app env: %s messenger addr: %s", params.print(), messenger.MessengerAddrFromEnv())

	// set up db for storage
	db, err := setupDB()
	if err != nil {
		log.Errorf("Main: failed to setup DB: %s", err.Error())
		panic(err)
	}
	defer db.Close()
	storage := store.NewStorage(db)

	// TODO: add a check to verify the node operator is a valid node operator
	operatorPrivateKey, err := params.loadDecryptedPrivateKey()
	if err != nil {
		log.Errorf("Main: failed to load decrypted private key: %s", err.Error())
		panic(err)
	}
	signer := keymanager.NewKeyManager(types.PrimusTestnet, operatorPrivateKey)

	network := messenger.NewMessengerClient(messenger.MessengerAddrFromEnv())

	config := &dkg.Config{
		KeygenProtocol:      frost.New,
		ReshareProtocol:     frost.NewResharing,
		KeySign:             keysign.NewSignature,
		Network:             network,
		Signer:              signer,
		Storage:             storage,
		SignatureDomainType: types.PrimusTestnet,
	}

	thisOperator, err := thisOperator(uint32(params.OperatorID), storage)
	if err != nil {
		log.Errorf("Main: failed to get operator %d from operator registry: %s", params.OperatorID, err.Error())
		panic(err)
	}
	dkgnode := dkg.NewNode(thisOperator, config)

	// register dkg operator node with the messenger
	if err := network.RegisterOperatorNode(strconv.Itoa(int(params.OperatorID)), os.Getenv("NODE_BROADCAST_ADDR")); err != nil {
		log.Errorf("Main: %s", err.Error())
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

	r.GET("/version", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"version": version,
		})
	})

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
