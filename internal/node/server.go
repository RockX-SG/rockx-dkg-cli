package node

import (
	"encoding/hex"
	"io"
	"net/http"

	"github.com/RockX-SG/frost-dkg-demo/internal/logger"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/gin-gonic/gin"
)

type ApiHandler struct {
	logger *logger.Logger
}

func New(logger *logger.Logger) *ApiHandler {
	return &ApiHandler{logger: logger}
}

func (h *ApiHandler) HandleConsume(node *dkg.Node) func(*gin.Context) {
	return func(c *gin.Context) {
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			h.logger.Errorf("HandleConsume: failed to read request body: %w", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to load data from request body",
				"error":   err.Error(),
			})
			return
		}

		msg := &types.SSVMessage{}
		if err = msg.Decode(data); err != nil {
			h.logger.Errorf("HandleConsume: failed to parse data from request body: %w", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to parse data from request body",
				"error":   err.Error(),
			})
			return
		}

		if err = node.ProcessMessage(msg); err != nil {
			h.logger.Errorf("HandleConsume: dkg node failed to process incoming message: %w", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "dkg node failed to process message",
				"error":   err.Error(),
			})
			return
		}

		h.logger.Infof("HandleConsume: dkg node processed incoming message successfully")
		c.JSON(http.StatusOK, gin.H{
			"message": "processed message successfully",
			"error":   nil,
		})
	}
}

func (h *ApiHandler) HandleGetDKGResults(node *dkg.Node) func(*gin.Context) {
	return func(c *gin.Context) {
		vkByte, _ := hex.DecodeString(c.Param("vk"))
		output, err := node.GetConfig().GetStorage().GetKeyGenOutput(vkByte)
		if err != nil {
			h.logger.Errorf("HandleGetDKGResults: failed to get dkg result for vk %s: %w", c.Param("vk"), err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, output)
	}
}
