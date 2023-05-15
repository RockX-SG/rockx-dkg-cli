package messenger

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/gin-gonic/gin"
)

func (m *Messenger) HandlePublish() func(*gin.Context) {

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
			"error":   "",
		})
	}
}

func (m *Messenger) HandleGetData() func(*gin.Context) {

	return func(c *gin.Context) {
		requestID := c.Param("request_id")

		_, ok := m.Data[requestID]
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusOK, m.Data[requestID])
	}
}

func (m *Messenger) HandleStreamDKGOutput() func(*gin.Context) {

	return func(c *gin.Context) {
		data := make(map[types.OperatorID]*dkg.SignedOutput)
		requestID := c.Query("request_id")

		body, _ := io.ReadAll(c.Request.Body)
		if err := json.Unmarshal(body, &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to parse request body",
				"error":   err.Error(),
			})
			return
		}

		m.Data[requestID] = &DataStore{DKGOutputs: data}
		c.JSON(http.StatusOK, nil)
	}
}

func (m *Messenger) HandleStreamDKGBlame() func(*gin.Context) {

	return func(c *gin.Context) {
		data := new(dkg.BlameOutput)
		requestID := c.Query("request_id")

		body, _ := io.ReadAll(c.Request.Body)
		if err := json.Unmarshal(body, &data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to parse request body",
				"error":   fmt.Sprintf("Error: %s", err.Error()),
			})
			return
		}

		m.Data[requestID] = &DataStore{BlameOutput: data}
		c.JSON(http.StatusOK, nil)
	}
}
