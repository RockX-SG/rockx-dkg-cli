package main

import (
	"fmt"
	"os"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/ping"
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
		Incoming: make(chan *messenger.Message, 5),
		Data:     make(map[string]*messenger.DataStore),
	}

	runner := workers.NewRunner()
	runner.AddJob(&workers.Job{
		ID: fmt.Sprintf("TOPIC__%s", messenger.DefaultTopic),
		Fn: m.ProcessIncomingMessageWorker,
	})

	go runner.Run()

	for _, sub := range m.Topics[messenger.DefaultTopic].Subscribers {
		sub.SubscribesTo[messenger.DefaultTopic] = m.Topics[messenger.DefaultTopic]
	}

	messengerAddr := os.Getenv("MESSENGER_ADDR")
	if messengerAddr == "" {
		messengerAddr = "0.0.0.0:3000"
	}

	r := gin.Default()
	r.GET("/ping", ping.HandlePing)

	// node registration
	r.POST("/register_node", messenger.HandleNodeRegistration(runner, m))

	// setup api routes
	r.POST("/publish", messenger.HandlePublish(m))
	r.POST("/stream/dkgoutput", messenger.HandleStreamDKGOutput(m))
	r.POST("/stream/dkgblame", messenger.HandleStreamDKGBlame(m))
	r.GET("/data/:request_id", messenger.HandleGetData(m))

	panic(r.Run(messengerAddr))
}
