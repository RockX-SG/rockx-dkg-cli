package main

import (
	"net/http"
	"os"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/workers"
	"github.com/gin-gonic/gin"
)

func main() {
	m := &messenger.Messenger{
		Topics: map[string]*messenger.Topic{
			messenger.DefaultTopic: {
				Name:        messenger.DefaultTopic,
				Subscribers: make(map[string]*messenger.Subscriber),
			},
		},
		Incoming: make(chan *messenger.Message),
		Data:     make(map[string]*messenger.DataStore),
	}

	go workers.ProcessIncomingMessageWorker(1, m)
	go workers.ProcessOutgoingMessageWorker(m)

	for _, sub := range m.Topics[messenger.DefaultTopic].Subscribers {
		sub.SubscribesTo[messenger.DefaultTopic] = m.Topics[messenger.DefaultTopic]
	}

	r := gin.Default()
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// node registration
	r.PUT("/register_node", messenger.HandleNodeRegistration(m))

	// setup api routes
	r.POST("/publish", messenger.HandlePublish(m))
	r.POST("/stream/dkgoutput", messenger.HandleStreamDKGOutput(m))
	r.POST("/stream/dkgblame", messenger.HandleStreamDKGBlame(m))
	r.GET("/data/:request_id", messenger.HandleGetData(m))

	HttpAddr := os.Getenv("MESSENGER_ADDR")
	if HttpAddr == "" {
		HttpAddr = "0.0.0.0:3000"
	}
	panic(r.Run(HttpAddr))
}
