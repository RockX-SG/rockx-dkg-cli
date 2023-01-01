package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
)

func main() {
	m := messenger.Messenger{
		Topics: map[string]*messenger.Topic{
			"default": {
				Name: "default",
				Subscribers: map[string]*messenger.Subscriber{
					"sub1": {
						Name:         "sub1",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message),
						ChMutex:      &sync.Mutex{},
					},
					"sub2": {
						Name:         "sub2",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message),
						ChMutex:      &sync.Mutex{},
					},
					"sub3": {
						Name:         "sub3",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message),
						ChMutex:      &sync.Mutex{},
					},
					"sub4": {
						Name:         "sub4",
						SubscribesTo: make(map[string]*messenger.Topic),
						Outgoing:     make(chan *messenger.Message),
						ChMutex:      &sync.Mutex{},
					},
				},
			},
		},
		Incoming: make(chan *messenger.Message),
		ChMutex:  &sync.Mutex{},
	}

	go m.ProcessIncomingMessageWorker(1)

	for _, sub := range m.Topics["default"].Subscribers {
		sub.SubscribesTo["default"] = m.Topics["default"]
		go sub.ProcessOutgoingMessageWorker(1)
	}

	m.Publish("default", []byte("hello world 1"))
	m.Publish("default", []byte("hello world 2"))
	m.Publish("default", []byte("hello world 3"))
	m.Publish("default", []byte("hello world 4"))
	m.Publish("default", []byte("hello world 5"))

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	fmt.Println("Received signal, shutting down...")
}
