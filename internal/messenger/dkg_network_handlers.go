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
