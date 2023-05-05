package main

import (
	"fmt"
	"os"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/ping"
	"github.com/RockX-SG/frost-dkg-demo/internal/workers"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// logger := logger.New()

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

	runner := workers.NewRunner()
	go runner.Run()

	runner.AddJob(&workers.Job{
		ID: fmt.Sprintf("TOPIC__%s", messenger.DefaultTopic),
		Fn: m.ProcessIncomingMessageWorker,
	})

	messengerAddr := os.Getenv("MESSENGER_ADDR")
	if messengerAddr == "" {
		messengerAddr = "0.0.0.0:3000"
	}

	r := gin.Default()
	setRoutes(r, m, runner)

	panic(r.Run(messengerAddr))
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
}
