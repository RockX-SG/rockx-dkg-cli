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

package node

import (
	"encoding/hex"
	"io"
	"net/http"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ApiHandler struct {
	logger *logrus.Logger
}

func New(logger *logrus.Logger) *ApiHandler {
	return &ApiHandler{logger: logger}
}

func (h *ApiHandler) HandleConsume(node *dkg.Node) func(*gin.Context) {
	return func(c *gin.Context) {
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			h.logger.Errorf("HandleConsume: failed to read request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to load data from request body",
				"error":   err.Error(),
			})
			return
		}

		msg := &types.SSVMessage{}
		if err = msg.Decode(data); err != nil {
			h.logger.Errorf("HandleConsume: failed to parse data from request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "failed to parse data from request body",
				"error":   err.Error(),
			})
			return
		}

		if err = node.ProcessMessage(msg); err != nil {
			h.logger.Errorf("HandleConsume: dkg node failed to process incoming message: %v", err)
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
			h.logger.Errorf("HandleGetDKGResults: failed to get dkg result for vk %s: %v", c.Param("vk"), err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, output)
	}
}
