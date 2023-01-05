package workers

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/dkg/frost"
	"github.com/bloxapp/ssv-spec/types"
)

var (
	maxRetriesAllowed = 10
)

func ProcessIncomingMessageWorker(id int, m *messenger.Messenger) {
	for msg := range m.Incoming {
		tp, exist := m.Topics[msg.Topic]
		if !exist {
			var err = &messenger.ErrTopicNotFound{TopicName: msg.Topic}
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

func ProcessOutgoingMessageWorker(m *messenger.Messenger) {
	for _, s := range m.Topics[messenger.DefaultTopic].Subscribers {
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
				var err = &messenger.ErrTopicNotFound{TopicName: msg.Topic}
				log.Printf("Error: %s\n", err.Error())
				continue
			}

			resp, err := http.Post(fmt.Sprintf("%s/consume", s.SrvAddr), "application/json", bytes.NewReader(msg.Data))
			if err != nil {
				log.Printf("Error: %s\n", err.Error())
				continue
			}

			respbody, _ := ioutil.ReadAll(resp.Body)
			if resp.StatusCode != http.StatusOK {
				s.Outgoing <- msg
				var err = fmt.Errorf("failed to publish message to the subscriber %s %v", s.Name, string(respbody))
				log.Printf("Error: %s\n", err.Error())
			}
			resp.Body.Close()
		}
	}
}
