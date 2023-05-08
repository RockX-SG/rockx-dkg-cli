package messenger

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type TopicJSON struct {
	TopicName   string   `json:"topic_name"`
	Subscribers []string `json:"subscribers"`
}

func (m *Messenger) GetTopics() func(*gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, m.Topics)
	}
}

func (m *Messenger) HandleCreateTopic() func(*gin.Context) {
	return func(c *gin.Context) {
		topicJSON := &TopicJSON{}
		if err := c.ShouldBindJSON(topicJSON); err != nil {
			m.logger.Errorf("HandleCreateTopic: failed to parse topic from request body: %w", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to load data from request body",
				"error":   err.Error(),
			})
			return
		}

		topic := Topic{
			Name:        topicJSON.TopicName,
			Subscribers: make(map[string]*Subscriber),
		}

		for _, sub := range topicJSON.Subscribers {
			subscriber, ok := m.Topics[DefaultTopic].Subscribers[sub]
			if ok {
				subscriber.SubscribesTo[topicJSON.TopicName] = &topic
				topic.Subscribers[sub] = subscriber
			}
		}
		m.Topics[topicJSON.TopicName] = &topic
		c.JSON(http.StatusOK, topic)
	}
}

func (m *Messenger) GetTopic() func(*gin.Context) {
	return func(c *gin.Context) {
		topic, exist := m.Topics[c.Param("topic_name")]
		if !exist {
			c.JSON(http.StatusNotFound, nil)
			return
		}
		c.JSON(http.StatusOK, topic)
	}
}

func (m *Messenger) DeleteTopic() func(*gin.Context) {
	return func(ctx *gin.Context) {
		topic, exist := m.Topics[ctx.Param("topic_name")]
		if !exist {
			ctx.JSON(http.StatusNotFound, nil)
			return
		}
		delete(m.Topics, topic.Name)
	}
}
