package cli

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/urfave/cli/v2"
)

func getRandRequestID() dkg.RequestID {
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
	return &CliHandler{
		client: http.DefaultClient,
	}
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

	messengerAddr := messenger.MessengerAddrFromEnv()

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
