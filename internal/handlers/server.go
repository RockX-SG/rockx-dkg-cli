package handlers

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/gin-gonic/gin"
)

func HandlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func HandleConsume(node *dkg.Node) func(*gin.Context) {
	return func(c *gin.Context) {
		data, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Error: %s\n", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to load data from request body",
				"error":   err.Error(),
			})
			return
		}

		msg := &types.SSVMessage{}
		if err = msg.Decode(data); err != nil {
			log.Printf("Error: %s\n", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to parse data from request body",
				"error":   err.Error(),
			})
			return
		}

		if err = node.ProcessMessage(msg); err != nil {
			log.Printf("Error: %s\n", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to process message",
				"error":   err.Error(),
			})
			return
		}

		log.Printf("HandleConsume finished successfully\n")
		c.JSON(http.StatusOK, gin.H{
			"message": "processed message successfully",
			"error":   nil,
		})
	}
}
