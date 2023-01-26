package messenger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/RockX-SG/frost-dkg-demo/internal/workers"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/gin-gonic/gin"
)

func HandleNodeRegistration(runner *workers.Runner, m *Messenger) func(*gin.Context) {
	return func(c *gin.Context) {

		subscribesTo := c.Query("subscribes_to")
		_, exist := m.Topics[subscribesTo]
		if !exist {
			err := &ErrTopicNotFound{TopicName: subscribesTo}
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "topic for subscription doesn't exist",
				"error":   err.Error(),
			})
			return
		}

		subscriber := &Subscriber{
			SubscribesTo: map[string]*Topic{},
			Outgoing:     make(chan *Message, 5),
			RetryData:    make(map[string]int),
		}

		body, _ := io.ReadAll(c.Request.Body)
		if err := json.Unmarshal(body, subscriber); err != nil {
			log.Println("failed to parse subscriber")
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to parse subscriber data from the request body",
				"error":   err.Error(),
			})
			return
		}

		if subscriber.Name == "" || subscriber.SrvAddr == "" {
			err := fmt.Errorf("Error: empty name %s or addr %s", subscriber.Name, subscriber.SrvAddr)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid subscriber data: empty name or addr",
				"error":   err.Error(),
			})
			return
		}

		existingSubscriber, ok := m.Topics[subscribesTo].Subscribers[subscriber.Name]
		if ok {
			existingSubscriber.SrvAddr = subscriber.SrvAddr
			m.Topics[subscribesTo].Subscribers[subscriber.Name] = existingSubscriber
		} else {
			subscriber.Outgoing = make(chan *Message, 5)
			subscriber.RetryData = make(map[string]int)
			subscriber.SubscribesTo[subscribesTo] = m.Topics[subscribesTo]
			m.Topics[subscribesTo].Subscribers[subscriber.Name] = subscriber

			runner.AddJob(&workers.Job{
				ID: fmt.Sprintf("SUBSCRIBER__%s", subscriber.Name),
				Fn: m.Topics[subscribesTo].Subscribers[subscriber.Name].ProcessOutgoingMessageWorker,
			})
		}
		c.JSON(http.StatusOK, nil)
	}
}

type CreateTopicReq struct {
	TopicName   string   `json:"topic_name"`
	Subscribers []string `json:"subscribers"`
}

func HandleCreateTopic(m *Messenger) func(*gin.Context) {
	return func(c *gin.Context) {
		req := &CreateTopicReq{}
		if err := c.ShouldBindJSON(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to load data from request body",
				"error":   err.Error(),
			})
			return
		}

		topic := Topic{
			Name:        req.TopicName,
			Subscribers: make(map[string]*Subscriber),
		}

		for _, sub := range req.Subscribers {
			subscriber, ok := m.Topics[DefaultTopic].Subscribers[sub]
			if ok {
				subscriber.SubscribesTo[req.TopicName] = &topic
				topic.Subscribers[sub] = subscriber
			}
		}

		m.Topics[req.TopicName] = &topic

		c.JSON(http.StatusOK, topic)
	}
}

func HandlePublish(m *Messenger) func(*gin.Context) {
	return func(c *gin.Context) {

		topicName := c.Query("topic_name")
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to load data from request body",
				"error":   err.Error(),
			})
			return
		}

		err = m.Publish(topicName, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("failed to publish data to topic %s", topicName),
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("message successfully published to topic %s", topicName),
			"error":   nil,
		})
	}
}

func HandleGetData(m *Messenger) func(*gin.Context) {
	return func(c *gin.Context) {

		requestID := c.Param("request_id")
		if requestID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "empty requestID in the http request",
				"error":   "query parameter `request_id` not found in the request",
			})
			return
		}

		_, ok := m.Data[requestID]
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusOK, m.Data[requestID])
	}
}

func HandleStreamDKGOutput(m *Messenger) func(*gin.Context) {
	return func(c *gin.Context) {

		requestID := c.Query("request_id")
		if requestID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "empty requestID in the http request",
				"error":   "query parameter `request_id` not found in the request",
			})
			return
		}

		body, _ := io.ReadAll(c.Request.Body)
		data := make(map[types.OperatorID]*dkg.SignedOutput)
		if err := json.Unmarshal(body, &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to parse request body",
				"error":   fmt.Sprintf("Error: %s", err.Error()),
			})
			return
		}

		m.Data[requestID] = &DataStore{
			DKGOutputs: data,
		}
		c.JSON(http.StatusOK, nil)
	}
}

func HandleStreamDKGBlame(m *Messenger) func(*gin.Context) {
	return func(c *gin.Context) {

		requestID := c.Query("request_id")
		if requestID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "empty requestID in the http request",
				"error":   "query parameter `request_id` not found in the request",
			})
			return
		}

		body, _ := io.ReadAll(c.Request.Body)
		data := make(map[types.OperatorID]*dkg.SignedOutput)
		if err := json.Unmarshal(body, &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to parse request body",
				"error":   fmt.Sprintf("Error: %s", err.Error()),
			})
			return
		}

		m.Data[requestID] = &DataStore{
			DKGOutputs: data,
		}
		c.JSON(http.StatusOK, nil)
	}
}
