package messenger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/gin-gonic/gin"
)

func HandlePublish(m *Messenger) func(*gin.Context) {
	return func(c *gin.Context) {
		topicName := c.Query("topic_name")
		data, err := ioutil.ReadAll(c.Request.Body)
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

		data, _ := json.Marshal(m.Data[requestID])
		c.JSON(http.StatusOK, data)
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

		body, _ := ioutil.ReadAll(c.Request.Body)
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

		body, _ := ioutil.ReadAll(c.Request.Body)
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
