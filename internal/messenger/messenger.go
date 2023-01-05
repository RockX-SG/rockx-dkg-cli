package messenger

import (
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
)

var (
	DefaultTopic = "default"
)

type Messenger struct {
	Topics map[string]*Topic
	Data   map[string]*DataStore

	Incoming chan *Message
}

type Topic struct {
	Name        string
	Subscribers map[string]*Subscriber
}

type Subscriber struct {
	Name         string            `json:"name"`
	SrvAddr      string            `json:"srv_addr"`
	SubscribesTo map[string]*Topic `json:"-"`
	Outgoing     chan *Message     `json:"-"`
	RetryData    map[string]int    `json:"-"`
}

type Message struct {
	Topic string
	Data  []byte
}

type DataStore struct {
	DKGOutputs  map[types.OperatorID]*dkg.SignedOutput
	BlameOutput *dkg.BlameOutput
}

func (m *Messenger) Publish(topicName string, data []byte) error {
	tp, exist := m.Topics[topicName]
	if !exist {
		return &ErrTopicNotFound{TopicName: topicName}
	}

	m.Incoming <- &Message{Topic: tp.Name, Data: data}
	return nil
}
