package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/gin-gonic/gin"
)

type KeygenReq struct {
	Operators map[types.OperatorID]string `json:"operators"`
	Threshold int                         `json:"threshold"`
}

type Apihandler struct {
	client *http.Client
}

func New() *Apihandler {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		DisableKeepAlives:   true,
	}

	return &Apihandler{
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: netTransport,
		},
	}
}

func (h *Apihandler) HandleGetData(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "empty requestID in the http request",
			"error":   "query parameter `request_id` not found in the request",
		})
		return
	}

	messengerAddr := os.Getenv("MESSENGER_SRV_ADDR")
	if messengerAddr == "" {
		messengerAddr = "0.0.0.0:3000"
	}

	resp, err := http.Get(fmt.Sprintf("http://%s/data/%s", messengerAddr, requestID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to make get data call to messenger",
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	data := messenger.DataStore{}
	if err := json.Unmarshal(body, &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to parse response",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Apihandler) HandleKeygen(c *gin.Context) {
	req := &KeygenReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to parse keygenn req from body",
			"error":   err.Error(),
		})
		return
	}

	ks := testingutils.TestingKeygenKeySet()
	requestID := testingutils.GetRandRequestID()

	operators := []types.OperatorID{}
	for operatorID, _ := range req.Operators {
		operators = append(operators, operatorID)
	}

	for operatorID, nodeAddr := range req.Operators {

		init := testingutils.InitMessageData(
			operators,
			uint16(req.Threshold),
			testingutils.TestingWithdrawalCredentials,
			testingutils.TestingForkVersion,
		)
		initBytes, _ := init.Encode()

		initMsg := testingutils.SignDKGMsg(ks.DKGOperators[operatorID].SK, operatorID, &dkg.Message{
			MsgType:    dkg.InitMsgType,
			Identifier: requestID,
			Data:       initBytes,
		})
		byts, _ := initMsg.Encode()

		msg := &types.SSVMessage{
			MsgType: types.DKGMsgType,
			Data:    byts,
		}

		msgBytes, err := msg.Encode()
		if err != nil {
			panic(err)
		}

		url := fmt.Sprintf("%s/consume", nodeAddr)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(msgBytes))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")

		resp, err := h.client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("failed to send init message to operator %d", operatorID),
				"error":   string(respBody),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"request_id": hex.EncodeToString(requestID[:]),
	})
}
