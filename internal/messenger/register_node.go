package messenger

import (
	"fmt"
	"net/http"

	"github.com/RockX-SG/frost-dkg-demo/internal/workers"
	"github.com/gin-gonic/gin"
)

func (m *Messenger) HandleNodeRegistration(runner *workers.Runner) func(*gin.Context) {

	return func(c *gin.Context) {

		subscribesTo := c.Query("subscribes_to")

		_, exist := m.Topics[subscribesTo]
		if !exist {
			err := &ErrTopicNotFound{TopicName: subscribesTo}
			m.logger.Errorf("HandleNodeRegistration: %w", err)
			c.JSON(http.StatusNotFound, gin.H{
				"message": fmt.Sprintf("topic %s doesn't exist", subscribesTo),
				"error":   err.Error(),
			})
			return
		}

		subscriber := &Subscriber{
			SubscribesTo: map[string]*Topic{},
			Outgoing:     make(chan *Message, 50),
			RetryData:    make(map[string]int),
		}

		if err := c.ShouldBindJSON(subscriber); err != nil {
			m.logger.Errorf("HandleNodeRegistration: failed to parse subscriber from request body: %w", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to parse subscriber data from the request body",
				"error":   err.Error(),
			})
			return
		}

		if subscriber.Name == "" || subscriber.SrvAddr == "" {
			err := fmt.Errorf("empty name %s or subscriber's address %s", subscriber.Name, subscriber.SrvAddr)
			m.logger.Errorf("HandleNodeRegistration: %w", err)
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
			subscriber.Outgoing = make(chan *Message, 50)
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
