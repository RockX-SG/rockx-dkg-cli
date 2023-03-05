package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/gin-gonic/gin"
)

type Apihandler struct {
	client            *http.Client
	requests          map[string]*KeygenReq
	resharingrequests map[string]*ResharingReq
}

func New() *Apihandler {
	return &Apihandler{
		client:            http.DefaultClient,
		requests:          make(map[string]*KeygenReq),
		resharingrequests: make(map[string]*ResharingReq),
	}
}

type KeygenReq struct {
	Operators            map[types.OperatorID]string `json:"operators"`
	Threshold            int                         `json:"threshold"`
	WithdrawalCredential string                      `json:"withdrawal_credentials"`
	ForkVersion          string                      `json:"fork_version"`
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

	withdrawalCred, _ := hex.DecodeString(req.WithdrawalCredential)
	forkVersion := types.NetworkFromString(req.ForkVersion).ForkVersion()

	ks := testingutils.TestingKeygenKeySet()
	requestID := testingutils.GetRandRequestID()

	operators := []types.OperatorID{}
	for operatorID, _ := range req.Operators {
		operators = append(operators, operatorID)
	}

	messengerClient := messenger.NewMessengerClient(messenger.MessengerAddrFromEnv())
	if err := messengerClient.CreateTopic(hex.EncodeToString(requestID[:]), operators); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to create new topic for this keygen req",
			"error":   err.Error(),
		})
		return
	}

	for operatorID, nodeAddr := range req.Operators {

		init := testingutils.InitMessageData(
			operators,
			uint16(req.Threshold),
			withdrawalCred,
			forkVersion,
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

	h.requests[hex.EncodeToString(requestID[:])] = req
	c.JSON(http.StatusOK, gin.H{
		"request_id": hex.EncodeToString(requestID[:]),
	})
}

type ResharingReq struct {
	Operators    map[types.OperatorID]string `json:"operators"`
	Threshold    int                         `json:"threshold"`
	ValidatorPK  string                      `json:"validator_pk"`
	OperatorsOld map[types.OperatorID]string `json:"operators_old"`
	KeygenReqID  string                      `json:"keygen_request_id"`
}

func (h *Apihandler) HandleResharing(c *gin.Context) {

	req := &ResharingReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to parse resharing req from body",
			"error":   err.Error(),
		})
		return
	}

	vk, _ := hex.DecodeString(req.ValidatorPK)

	ks := testingutils.TestingResharingKeySet()
	requestID := testingutils.GetRandRequestID()

	operators := []types.OperatorID{}
	for operatorID, _ := range req.Operators {
		operators = append(operators, operatorID)
	}

	operatorsOld := []types.OperatorID{}
	for operatorID, _ := range req.OperatorsOld {
		operatorsOld = append(operatorsOld, operatorID)
	}

	alloperators := append(operators, operatorsOld...)

	messengerClient := messenger.NewMessengerClient(messenger.MessengerAddrFromEnv())
	if err := messengerClient.CreateTopic(hex.EncodeToString(requestID[:]), alloperators); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to create new topic for this resharing req",
			"error":   err.Error(),
		})
		return
	}

	for _, operatorID := range alloperators {

		var nodeAddr string
		_, ok := req.Operators[operatorID]
		if ok {
			nodeAddr = req.Operators[operatorID]
		} else {
			nodeAddr = req.OperatorsOld[operatorID]
		}

		reshare := testingutils.ReshareMessageData(
			operators,
			uint16(req.Threshold),
			vk,
			operatorsOld,
		)
		reshareBytes, _ := reshare.Encode()

		reshareMsg := testingutils.SignDKGMsg(ks.DKGOperators[operatorID].SK, operatorID, &dkg.Message{
			MsgType:    dkg.ReshareMsgType,
			Identifier: requestID,
			Data:       reshareBytes,
		})
		byts, _ := reshareMsg.Encode()

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
				"message": fmt.Sprintf("failed to send reshare message to operator %d", operatorID),
				"error":   string(respBody),
			})
			return
		}
	}

	h.resharingrequests[hex.EncodeToString(requestID[:])] = req
	c.JSON(http.StatusOK, gin.H{
		"request_id": hex.EncodeToString(requestID[:]),
	})
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

	results, err := h.fetchDKGResults(requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to fetch results for keygen request",
			"error":   err.Error(),
		})
	}

	c.JSON(http.StatusOK, results)
}

func (h *Apihandler) HandleGetDataByOperatorID(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "empty requestID in the http request",
			"error":   "query parameter `request_id` not found in the request",
		})
		return
	}

	operatorIDParam := c.Param("operator_id")
	if operatorIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "empty operatorID in the http request",
			"error":   "query parameter `operator_id` not found in the request",
		})
		return
	}

	operatorID, _ := strconv.Atoi(operatorIDParam)

	results, err := h.fetchDKGResults(requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to fetch results for keygen request",
			"error":   err.Error(),
		})
	}

	c.JSON(http.StatusOK, results.Output[types.OperatorID(operatorID)])
}

type DepositDataJson struct {
	PubKey                string      `json:"pubkey"`
	WithdrawalCredentials string      `json:"withdrawal_credentials"`
	Amount                phase0.Gwei `json:"amount"`
	Signature             string      `json:"signature"`
	DepositMessageRoot    string      `json:"deposit_message_root"`
	DepositDataRoot       string      `json:"deposit_data_root"`
	ForkVersion           string      `json:"fork_version"`
	NetworkName           string      `json:"network_name"`
	DepositCliVersion     string      `json:"deposit_cli_version"`
}

func (h *Apihandler) HandleGetDepositData(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "empty requestID in the http request",
			"error":   "query parameter `request_id` not found in the request",
		})
		return
	}

	req, ok := h.requests[requestID]
	if !ok {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	results, err := h.fetchDKGResults(requestID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	validatorPK, _ := hex.DecodeString(results.Output[1].Data.ValidatorPubKey)
	withdrawalCredentials, _ := hex.DecodeString(req.WithdrawalCredential)
	fork := types.NetworkFromString(req.ForkVersion).ForkVersion()
	amount := phase0.Gwei(types.MaxEffectiveBalanceInGwei)

	_, depositData, err := types.GenerateETHDepositData(validatorPK, withdrawalCredentials, fork, types.DomainDeposit)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	depositMsg := &phase0.DepositMessage{
		PublicKey:             depositData.PublicKey,
		WithdrawalCredentials: withdrawalCredentials,
		Amount:                amount,
	}
	depositMsgRoot, _ := depositMsg.HashTreeRoot()

	blsSigBytes, _ := hex.DecodeString(results.Output[1].Data.DepositDataSignature)
	blsSig := phase0.BLSSignature{}
	copy(blsSig[:], blsSigBytes)
	depositData.Signature = blsSig

	depositDataRoot, _ := depositData.HashTreeRoot()

	response := DepositDataJson{
		PubKey:                results.Output[1].Data.ValidatorPubKey,
		WithdrawalCredentials: req.WithdrawalCredential,
		Amount:                amount,
		Signature:             results.Output[1].Data.DepositDataSignature,
		DepositMessageRoot:    hex.EncodeToString(depositMsgRoot[:]),
		DepositDataRoot:       hex.EncodeToString(depositDataRoot[:]),
		ForkVersion:           hex.EncodeToString(fork[:]),
		DepositCliVersion:     "2.3.0",
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=deposit-data_%d.json", time.Now().UTC().Unix()))
	c.JSON(http.StatusOK, []DepositDataJson{response})
}

func (h *Apihandler) fetchDKGResults(requestID string) (*DKGResult, error) {

	messengerAddr := os.Getenv("MESSENGER_SRV_ADDR")
	if messengerAddr == "" {
		messengerAddr = "http://0.0.0.0:3000"
	}

	resp, err := http.Get(fmt.Sprintf("%s/data/%s", messengerAddr, requestID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data := &messenger.DataStore{}

	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	formattedOutput := formatResults(data)
	return &formattedOutput, nil
}

type DKGResult struct {
	Output map[types.OperatorID]SignedOutput `json:"output"`
	Blame  *dkg.BlameOutput                  `json:"blame"`
}

type Output struct {
	RequestID            string
	EncryptedShare       string
	SharePubKey          string
	ValidatorPubKey      string
	DepositDataSignature string
}

type SignedOutput struct {
	Data      Output
	Signer    string
	Signature string
}

func formatResults(data *messenger.DataStore) DKGResult {
	if data.BlameOutput != nil {
		return formatBlameResults(data.BlameOutput)
	}

	output := make(map[types.OperatorID]SignedOutput)
	for operatorID, signedOutput := range data.DKGOutputs {
		getHex := hex.EncodeToString
		v := SignedOutput{
			Data: Output{
				RequestID:            getHex(signedOutput.Data.RequestID[:]),
				EncryptedShare:       getHex(signedOutput.Data.EncryptedShare),
				SharePubKey:          getHex(signedOutput.Data.SharePubKey),
				ValidatorPubKey:      getHex(signedOutput.Data.ValidatorPubKey),
				DepositDataSignature: getHex(signedOutput.Data.DepositDataSignature),
			},
			Signer:    strconv.Itoa(int(signedOutput.Signer)),
			Signature: hex.EncodeToString(signedOutput.Signature),
		}
		output[operatorID] = v
	}

	return DKGResult{Output: output}
}

func formatBlameResults(blameOutput *dkg.BlameOutput) DKGResult {
	return DKGResult{Blame: blameOutput}
}
