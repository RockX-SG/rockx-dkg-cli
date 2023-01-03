package main

import (
	"os"
	"sync"

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
					"operator-1": {
						Name:         "operator-1",
						SrvAddr:      "http://0.0.0.0:8081",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message),
						ChMutex:      &sync.Mutex{},
					},
					"operator-2": {
						Name:         "operator-2",
						SrvAddr:      "http://0.0.0.0:8082",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message),
						ChMutex:      &sync.Mutex{},
					},
					"operator-3": {
						Name:         "operator-3",
						SrvAddr:      "http://0.0.0.0:8083",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message),
						ChMutex:      &sync.Mutex{},
					},
					"operator-4": {
						Name:         "operator-4",
						SrvAddr:      "http://0.0.0.0:8084",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message),
						ChMutex:      &sync.Mutex{},
					},
				},
			},
		},
		Incoming: make(chan *messenger.Message),
		ChMutex:  &sync.Mutex{},
		Data:     make(map[string]*messenger.DataStore),
	}

	go workers.ProcessIncomingMessageWorker(1, m)

	for _, sub := range m.Topics[messenger.DefaultTopic].Subscribers {
		sub.SubscribesTo[messenger.DefaultTopic] = m.Topics[messenger.DefaultTopic]
		go workers.ProcessOutgoingMessageWorker(1, sub)
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
