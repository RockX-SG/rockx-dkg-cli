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
	"flag"
	"fmt"
	"net/http"

	"github.com/RockX-SG/frost-dkg-demo/internal/logger"
	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/ping"
	"github.com/RockX-SG/frost-dkg-demo/internal/workers"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const serviceName = "messenger"

var (
	version  string
	httpAddr string
)

func init() {
	flag.StringVar(&httpAddr, "http-addr", "0.0.0.0:3000", "host:port of the application")
}

func main() {
	flag.Parse()

	log := logger.New(serviceName)
	m := &messenger.Messenger{
		Topics: map[string]*messenger.Topic{
			messenger.DefaultTopic: {
				Name:        messenger.DefaultTopic,
				Subscribers: make(map[string]*messenger.Subscriber),
			},
		},
		Incoming: make(chan *messenger.Message, 50),
		Data:     make(map[string]*messenger.DataStore),
	}
	m.WithLogger(log)

	runner := workers.NewRunner(log)
	go runner.Run()

	runner.AddJob(&workers.Job{
		ID: fmt.Sprintf("TOPIC__%s", messenger.DefaultTopic),
		Fn: m.ProcessIncomingMessageWorker,
	})

	r := gin.Default()
	r.Use(logger.GinLogger(log))
	setRoutes(r, m, runner)

	panic(r.Run(httpAddr))
}

func setRoutes(r *gin.Engine, m *messenger.Messenger, runner *workers.Runner) {

	r.GET("/ping", ping.HandlePing)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// CRUD APIs for Topics
	r.GET("/topics", m.GetTopics())
	r.POST("/topics", m.HandleCreateTopic())
	r.GET("/topics/:topic_name", m.GetTopic())
	r.DELETE("/topics/:topic_name", m.DeleteTopic())

	// Register a node
	r.POST("/register_node", m.HandleNodeRegistration(runner))

	// DKG network implementation
	r.POST("/publish", m.HandlePublish())
	r.POST("/stream/dkgoutput", m.HandleStreamDKGOutput())
	r.POST("/stream/dkgblame", m.HandleStreamDKGBlame())
	r.GET("/data/:request_id", m.HandleGetData())

	r.GET("/version", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"version": version,
		})
	})
}
