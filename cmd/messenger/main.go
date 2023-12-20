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

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/RockX-SG/frost-dkg-demo/internal/logger"
	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/ping"
	"github.com/RockX-SG/frost-dkg-demo/internal/workers"
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

	worker := workers.NewRunner(log)
	go worker.Run()

	worker.AddJob(&workers.Job{
		ID: fmt.Sprintf("TOPIC__%s", messenger.DefaultTopic),
		Fn: m.ProcessIncomingMessageWorker,
	})

	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.Use(logger.GinLogger(log))

	InitializeAPIEndpoints(r, m, worker)

	log.Infof("Starting %s on %s", serviceName, httpAddr)
	panic(r.Run(httpAddr))
}

func InitializeAPIEndpoints(r *gin.Engine, m *messenger.Messenger, w *workers.Runner) {

	// Message Topic - CRUD APIs
	topicsGroup := r.Group("/topics")
	{
		topicsGroup.GET("", m.GetTopics())
		topicsGroup.POST("", m.CreateOrUpdateTopic())
		topicsGroup.GET("/:topic_name", m.GetTopic())
		topicsGroup.DELETE("/:topic_name", m.DeleteTopic())
	}

	// DKG Node Registration
	r.POST("/register_node", m.HandleNodeRegistration(w))

	// DKG network layer actions
	r.POST("/publish", m.HandlePublish())
	r.POST("/stream/dkgoutput", m.HandleStreamDKGOutput())
	r.POST("/stream/dkgblame", m.HandleStreamDKGBlame())
	r.GET("/data/:request_id", m.HandleGetData())

	// Service Health and Monitoring APIs
	r.GET("/ping", ping.HandlePing)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/version", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"version": version,
		})
	})
}
