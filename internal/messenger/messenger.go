package messenger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/dkg/frost"
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

func (m *Messenger) ProcessIncomingMessageWorker(ctx *context.Context) {
	for msg := range m.Incoming {
		tp, exist := m.Topics[msg.Topic]
		if !exist {
			var err = &ErrTopicNotFound{TopicName: msg.Topic}
			log.Printf("Error: %s\n", err.Error())
		}

		ssvMsg := &types.SSVMessage{}
		if err := ssvMsg.Decode(msg.Data); err != nil {
			log.Printf("Error: %s\n", err.Error())
		}

		signedMsg := &dkg.SignedMessage{}
		if err := signedMsg.Decode(ssvMsg.Data); err != nil {
			log.Printf("Error: %s\n", err.Error())
		}

		protocolMsg := &frost.ProtocolMsg{}
		if err := protocolMsg.Decode(signedMsg.Message.Data); err != nil {
			log.Printf("Error: %s\n", err.Error())
		}

		log.Printf("received message from %d for msgType %d round %d \n", signedMsg.Signer, signedMsg.Message.MsgType, protocolMsg.Round)

		for _, subscriber := range tp.Subscribers {
			operatorID := strconv.Itoa(int(signedMsg.Signer))
			if operatorID == subscriber.Name {
				continue
			}
			subscriber.Outgoing <- msg
		}
	}
}

var (
	maxRetriesAllowed = 10
)

func (s *Subscriber) ProcessOutgoingMessageWorker(ctx *context.Context) {
	for msg := range s.Outgoing {

		h := sha256.Sum256(msg.Data)
		k := base64.RawStdEncoding.EncodeToString(h[:])

		numRetries, ok := s.RetryData[k]
		if ok {
			if numRetries >= maxRetriesAllowed {
				continue
			} else {
				s.RetryData[k]++
			}
		} else {
			s.RetryData[k] = 0
		}

		if numRetries > 0 {
			time.Sleep(time.Duration(numRetries) * (time.Second))
		}

		_, exist := s.SubscribesTo[msg.Topic]
		if !exist {
			var err = &ErrTopicNotFound{TopicName: msg.Topic}
			log.Printf("Error: %s\n", err.Error())
			continue
		}

		resp, err := http.Post(fmt.Sprintf("%s/consume", s.SrvAddr), "application/json", bytes.NewBuffer(msg.Data))
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
			continue
		}

		// TODO: remove this after testing
		respbody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			s.Outgoing <- msg

			err := fmt.Errorf("failed to publish message to the subscriber %s %v", s.Name, string(respbody))
			log.Printf("Error: %s\n", err.Error())
		}
		resp.Body.Close()
	}
}

func MessengerAddrFromEnv() string {
	messengerAddr := os.Getenv("MESSENGER_SRV_ADDR")
	if messengerAddr == "" {
		messengerAddr = "https://dkg-messenger.rockx.com"
	}
	return messengerAddr
}
