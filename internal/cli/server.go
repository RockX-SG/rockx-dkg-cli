package cli

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/urfave/cli/v2"
)

func GetRandRequestID() dkg.RequestID {
	requestID := dkg.RequestID{}
	for i := range requestID {
		rndInt, _ := rand.Int(rand.Reader, big.NewInt(255))
		if len(rndInt.Bytes()) == 0 {
			requestID[i] = 0
		} else {
			requestID[i] = rndInt.Bytes()[0]
		}
	}
	return requestID
}

type CliHandler struct {
	client *http.Client
}

func New() *CliHandler {
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		DisableKeepAlives:   true,
	}
	return &CliHandler{
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: netTransport,
		},
	}
}

type KeygenReq struct {
	Operators            map[types.OperatorID]string `json:"operators"`
	Threshold            int                         `json:"threshold"`
	WithdrawalCredential string                      `json:"withdrawal_credentials"`
	ForkVersion          string                      `json:"fork_version"`
}

func (h *CliHandler) HandleKeygen(c *cli.Context) error {
	req := KeygenReq{
		Operators:            make(map[types.OperatorID]string),
		Threshold:            c.Int("threshold"),
		WithdrawalCredential: c.String("withdrawal-credentials"),
		ForkVersion:          c.String("fork-version"),
	}
	operatorkv := c.StringSlice("operator")
	for _, op := range operatorkv {
		op = strings.Trim(op, " ")
		pair := strings.Split(op, "=")
		if len(pair) != 2 {
			return fmt.Errorf("operator %s is not in the form of key=value", op)
		}
		opID, err := strconv.Atoi(pair[0])
		if err != nil {
			return err
		}
		req.Operators[types.OperatorID(opID)] = pair[1]
	}

	withdrawalCred, _ := hex.DecodeString(req.WithdrawalCredential)
	forkVersion := types.NetworkFromString(req.ForkVersion).ForkVersion()

	ks := testingutils.TestingKeygenKeySet()
	requestID := GetRandRequestID()

	operators := []types.OperatorID{}
	for operatorID, _ := range req.Operators {
		operators = append(operators, operatorID)
	}

	messengerClient := messenger.NewMessengerClient(messenger.MessengerAddrFromEnv())
	if err := messengerClient.CreateTopic(hex.EncodeToString(requestID[:]), operators); err != nil {
		return err
	}

	init := testingutils.InitMessageData(
		operators,
		uint16(req.Threshold),
		withdrawalCred,
		forkVersion,
	)
	initBytes, _ := init.Encode()

	signedInitMsg := testingutils.SignDKGMsg(ks.DKGOperators[1].SK, 1, &dkg.Message{
		MsgType:    dkg.InitMsgType,
		Identifier: requestID,
		Data:       initBytes,
	})
	signedInitMsgBytes, _ := signedInitMsg.Encode()

	ssvMsg := &types.SSVMessage{
		MsgType: types.DKGMsgType,
		Data:    signedInitMsgBytes,
	}
	ssvMsgBytes, err := ssvMsg.Encode()
	if err != nil {
		return err
	}

	for operatorID, nodeAddr := range req.Operators {
		url := fmt.Sprintf("%s/consume", nodeAddr)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(ssvMsgBytes))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")

		resp, err := h.client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to send init message to the operator %d", operatorID)
		}
	}

	fmt.Println(hex.EncodeToString(requestID[:]))
	return nil
}

type ResharingReq struct {
	Operators    map[types.OperatorID]string `json:"operators"`
	Threshold    int                         `json:"threshold"`
	ValidatorPK  string                      `json:"validator_pk"`
	OperatorsOld map[types.OperatorID]string `json:"operators_old"`
}

func (h *CliHandler) HandleResharing(c *cli.Context) error {
	req := ResharingReq{
		Operators:    make(map[types.OperatorID]string),
		OperatorsOld: make(map[types.OperatorID]string),
		Threshold:    c.Int("threshold"),
		ValidatorPK:  c.String("validator-pk"),
	}

	operatorkv := c.StringSlice("operator")
	for _, op := range operatorkv {
		op = strings.Trim(op, " ")
		pair := strings.Split(op, "=")
		if len(pair) != 2 {
			return fmt.Errorf("operator %s is not in the form of key=value", op)
		}
		opID, err := strconv.Atoi(pair[0])
		if err != nil {
			return err
		}
		req.Operators[types.OperatorID(opID)] = pair[1]
	}

	oldoperatorkv := c.StringSlice("old-operator")
	for _, op := range oldoperatorkv {
		op = strings.Trim(op, " ")
		pair := strings.Split(op, "=")
		if len(pair) != 2 {
			return fmt.Errorf("operator %s is not in the form of key=value", op)
		}
		opID, err := strconv.Atoi(pair[0])
		if err != nil {
			return err
		}
		req.OperatorsOld[types.OperatorID(opID)] = pair[1]
	}

	vk, err := hex.DecodeString(req.ValidatorPK)
	if err != nil {
		return err
	}

	ks := testingutils.TestingResharingKeySet()
	requestID := GetRandRequestID()

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
		return err
	}

	reshare := testingutils.ReshareMessageData(
		operators,
		uint16(req.Threshold),
		vk,
		operatorsOld,
	)
	reshareBytes, _ := reshare.Encode()

	reshareMsg := testingutils.SignDKGMsg(ks.DKGOperators[5].SK, 5, &dkg.Message{
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

	for _, operatorID := range alloperators {
		var nodeAddr string
		_, ok := req.Operators[operatorID]
		if ok {
			nodeAddr = req.Operators[operatorID]
		} else {
			nodeAddr = req.OperatorsOld[operatorID]
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

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to send reshare message with code %d to operator %d", resp.StatusCode, operatorID)
		}
	}

	fmt.Println(hex.EncodeToString(requestID[:]))
	return nil
}

func (h *CliHandler) HandleGetData(c *cli.Context) error {
	requestID := c.String("request-id")
	if requestID == "" {
		return fmt.Errorf("`request_id` not found")
	}

	results, err := h.fetchDKGResults(requestID)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("dkg_results_%s_%d.json", requestID, time.Now().Unix())
	fmt.Printf("writing results to file: %s\n", filename)
	return WriteJSONToFile(results, filename)
}

func WriteJSONToFile(results *DKGResult, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(results)
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

func (h *CliHandler) HandleGetDepositData(c *cli.Context) error {
	requestID := c.String("request-id")
	if requestID == "" {
		return fmt.Errorf("`request_id` not found")
	}

	results, err := h.fetchDKGResults(requestID)
	if err != nil {
		return err
	}

	validatorPK, _ := hex.DecodeString(results.Output[1].Data.ValidatorPubKey)
	withdrawalCredentials, _ := hex.DecodeString(c.String("withdrawal-credentials"))
	fork := types.NetworkFromString(c.String("fork-version")).ForkVersion()
	amount := phase0.Gwei(types.MaxEffectiveBalanceInGwei)

	_, depositData, err := types.GenerateETHDepositData(validatorPK, withdrawalCredentials, fork, types.DomainDeposit)
	if err != nil {
		return err
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
		WithdrawalCredentials: c.String("withdrawal-credentials"),
		Amount:                amount,
		Signature:             results.Output[1].Data.DepositDataSignature,
		DepositMessageRoot:    hex.EncodeToString(depositMsgRoot[:]),
		DepositDataRoot:       hex.EncodeToString(depositDataRoot[:]),
		ForkVersion:           hex.EncodeToString(fork[:]),
		DepositCliVersion:     "2.3.0",
	}

	filename := fmt.Sprintf("deposit-data_%d.json", time.Now().UTC().Unix())
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(response); err != nil {
		return err
	}

	fmt.Printf("writing deposit data json to file %s\n", filename)
	return nil
}

func (h *CliHandler) fetchDKGResults(requestID string) (*DKGResult, error) {

	messengerAddr := os.Getenv("MESSENGER_SRV_ADDR")
	if messengerAddr == "" {
		messengerAddr = "http://0.0.0.0:3000"
	}

	url := fmt.Sprintf("%s/data/%s", messengerAddr, requestID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch dkg result for request %s with code %d", requestID, resp.StatusCode)
	}

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
