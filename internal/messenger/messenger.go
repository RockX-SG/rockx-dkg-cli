/*
 * ==================================================================
 *Copyright (C) 2022-2023 Altstake Technology Pte. Ltd. (RockX)
 *This file is part of rockx-dkg-cli <https://github.com/RockX-SG/rockx-dkg-cli>
 *CAUTION: THESE CODES HAVE NOT BEEN AUDITED
 *
 *rockx-dkg-cli is free software: you can redistribute it and/or modify
 *it under the terms of the GNU General Public License as published by
 *the Free Software Foundation, either version 3 of the License, or
 *(at your option) any later version.
 *
 *rockx-dkg-cli is distributed in the hope that it will be useful,
 *but WITHOUT ANY WARRANTY; without even the implied warranty of
 *MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *GNU General Public License for more details.
 *
 *You should have received a copy of the GNU General Public License
 *along with rockx-dkg-cli. If not, see <http://www.gnu.org/licenses/>.
 *==================================================================
 */

package messenger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/workers"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/dkg/frost"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/sirupsen/logrus"
)

var (
	DefaultTopic = "default"
)

type Messenger struct {
	Topics map[string]*Topic
	Data   map[string]*DataStore

	Incoming chan *Message

	logger *logrus.Logger
}

func (m *Messenger) WithLogger(logger *logrus.Logger) {
	m.logger = logger
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
		m.logger.Errorf("Publish: topic %s already exists", topicName)
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
			m.logger.Errorf("ProcessIncomingMessageWorker: %v", err)
			continue
		}

		ssvMsg := &types.SSVMessage{}
		if err := ssvMsg.Decode(msg.Data); err != nil {
			m.logger.Errorf("ProcessIncomingMessageWorker: %v", err)
			continue
		}
		signedMsg := &dkg.SignedMessage{}
		if err := signedMsg.Decode(ssvMsg.Data); err != nil {
			m.logger.Errorf("ProcessIncomingMessageWorker: failed to decode signed message: %v", err)
			continue
		}
		protocolMsg := &frost.ProtocolMsg{}
		if err := protocolMsg.Decode(signedMsg.Message.Data); err != nil {
			m.logger.Errorf("ProcessIncomingMessageWorker: failed to decode protocol message: %v", err)
			continue
		}

		m.logger.Debugf(
			"received message from %d for msgType %d round %d",
			signedMsg.Signer,
			signedMsg.Message.MsgType,
			protocolMsg.Round,
		)

		for _, subscriber := range tp.Subscribers {
			operatorID := strconv.Itoa(int(signedMsg.Signer))
			if operatorID == subscriber.Name {
				continue
			}
			subscriber.Outgoing <- msg
		}
	}
}

const (
	maxRetriesAllowed = 10
)

func (s *Subscriber) ProcessOutgoingMessageWorker(ctx *context.Context) {

	log := (*ctx).Value(workers.Ctxlog("logger"))
	if log == nil {
		panic("logger not found in context")
	}
	logger := log.(*logrus.Logger)
	logger.Infof("ProcessOutgoingMessageWorker: logger loaded successfully")

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
			time.Sleep(2 * (time.Second))
		}

		_, exist := s.SubscribesTo[msg.Topic]
		if !exist {
			var err = &ErrTopicNotFound{TopicName: msg.Topic}
			logger.Errorf("ProcessOutgoingMessageWorker: %v", err)
			continue
		}

		// TODO: replace this client
		resp, err := http.Post(fmt.Sprintf("%s/consume", s.SrvAddr), "application/json", bytes.NewBuffer(msg.Data))
		if err != nil {
			logger.Errorf("ProcessOutgoingMessageWorker: %v", err)
			continue
		}

		respbody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			s.Outgoing <- msg

			err := fmt.Errorf("failed to publish message to the subscriber %s %v", s.Name, string(respbody))
			logger.Errorf("ProcessOutgoingMessageWorker: %v", err)
		} else {
			logger.Infof("ProcessOutgoingMessageWorker: message sent to %s successfully", s.Name)
		}
		resp.Body.Close()
	}
}

func MessengerAddrFromEnv() string {
	messengerAddr := os.Getenv("MESSENGER_SRV_ADDR")
	if messengerAddr == "" {
		messengerAddr = "http://dkg-messenger.rockx.com"
	}
	return messengerAddr
}
