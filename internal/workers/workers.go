package workers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
)

func ProcessIncomingMessageWorker(id int, m *messenger.Messenger) {
	for msg := range m.Incoming {
		tp, exist := m.Topics[msg.Topic]
		if !exist {
			var err = &messenger.ErrTopicNotFound{TopicName: msg.Topic}
			log.Printf("Error: %s\n", err.Error())
		}

		log.Printf("received message %+v\n", msg.Data)
		ssvMsg := &types.SSVMessage{}
		if err := ssvMsg.Decode(msg.Data); err != nil {
			log.Printf("Error: %s\n", err.Error())
		}

		signedMsg := &dkg.SignedMessage{}
		if err := signedMsg.Decode(ssvMsg.Data); err != nil {
			log.Printf("Error: %s\n", err.Error())
		}

		protocolMsg := signedMsg.Message.Data

		for _, subscriber := range tp.Subscribers {
			subscriber.ChMutex.Lock()
			subscriber.Outgoing <- msg
			subscriber.ChMutex.Unlock()
		}
	}
}

func ProcessOutgoingMessageWorker(id int, s *messenger.Subscriber) {
	for msg := range s.Outgoing {
		_, exist := s.SubscribesTo[msg.Topic]
		if !exist {
			var err = &messenger.ErrTopicNotFound{TopicName: msg.Topic}
			log.Printf("Error: %s\n", err.Error())
			continue
		}

		resp, err := http.Post(fmt.Sprintf("%s/consume", s.SrvAddr), "application/json", bytes.NewReader(msg.Data))
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
			continue
		}

		if resp.StatusCode != http.StatusOK {
			var err = fmt.Errorf("failed to publish message to the subscriber %s", s.Name)
			log.Printf("Error: %s\n", err.Error())
		}
		resp.Body.Close()
	}
}
