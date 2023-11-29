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

func (m *Messenger) CreateOrUpdateTopic() func(*gin.Context) {
	return func(c *gin.Context) {
		topicJSON := &TopicJSON{}
		if err := c.ShouldBindJSON(topicJSON); err != nil {
			m.logger.Errorf("HandleCreateTopic: failed to parse topic from request body: %v", err)
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
