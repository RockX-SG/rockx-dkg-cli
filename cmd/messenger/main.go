package main

import (
	"os"

	"github.com/RockX-SG/frost-dkg-demo/internal/handlers"
	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/workers"
	"github.com/gin-gonic/gin"
)

func main() {
	m := &messenger.Messenger{
		Topics: map[string]*messenger.Topic{
			messenger.DefaultTopic: {
				Name: messenger.DefaultTopic,
				Subscribers: map[string]*messenger.Subscriber{
					"1": {
						Name:         "1",
						SrvAddr:      "http://0.0.0.0:8081",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message, 5),
						RetryData:    make(map[string]int),
					},
					"2": {
						Name:         "2",
						SrvAddr:      "http://0.0.0.0:8082",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message, 5),
						RetryData:    make(map[string]int),
					},
					"3": {
						Name:         "3",
						SrvAddr:      "http://0.0.0.0:8083",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message, 5),
						RetryData:    make(map[string]int),
					},
					"4": {
						Name:         "4",
						SrvAddr:      "http://0.0.0.0:8084",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message, 5),
						RetryData:    make(map[string]int),
					},
				},
			},
		},
		Incoming: make(chan *messenger.Message),
		Data:     make(map[string]*messenger.DataStore),
	}

	go workers.ProcessIncomingMessageWorker(1, m)

	for _, sub := range m.Topics[messenger.DefaultTopic].Subscribers {
		sub.SubscribesTo[messenger.DefaultTopic] = m.Topics[messenger.DefaultTopic]
		go workers.ProcessOutgoingMessageWorker(sub)
	}

	r := gin.Default()
	r.GET("/ping", handlers.HandlePing)
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
