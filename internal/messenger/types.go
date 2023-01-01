package messenger

import "sync"

var (
	defaultTopic = "topic-keygen-default"
)

type Messenger struct {
	Topics map[string]*Topic

	ChMutex  *sync.Mutex
	Incoming chan *Message
}

type Topic struct {
	Name        string
	Subscribers map[string]*Subscriber
}

type Subscriber struct {
	Name         string
	SubscribesTo map[string]*Topic

	ChMutex  *sync.Mutex
	Outgoing chan *Message
}

type Message struct {
	Topic string
	Data  []byte
}
