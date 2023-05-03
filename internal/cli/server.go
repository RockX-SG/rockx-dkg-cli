package cli

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/storage"
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

type KeyShares struct {
	Version   string           `json:"version"`
	Data      KeySharesData    `json:"data"`
	Payload   KeySharesPayload `json:"payload"`
	CreatedAt time.Time        `json:"createdAt"`
}

type KeySharesData struct {
	PublicKey string         `json:"publicKey"`
	Operators []OperatorData `json:"operators"`
	Shares    KeySharesKeys  `json:"shares"`
}

type OperatorData struct {
	ID        uint32 `json:"id"`
	PublicKey string `json:"publicKey"`
}

type KeySharesKeys struct {
	PublicKeys    []string `json:"publicKeys"`
	EncryptedKeys []string `json:"encryptedKeys"`
}

type ReadablePayload struct {
	PublicKey   string   `json:"publicKey"`
	OperatorIDs []uint32 `json:"operatorIds"`
	Shares      string   `json:"shares"`
	Amount      string   `json:"amount"`
	Cluster     string   `json:"cluster"`
}

type KeySharesPayload struct {
	Readable ReadablePayload `json:"readable"`
}

func (h *CliHandler) HandleGetKeyShares(c *cli.Context) error {
	requestID := c.String("request-id")
	if requestID == "" {
		return fmt.Errorf("`request_id` not found")
	}

	results, err := h.fetchDKGResults(requestID)
	if err != nil {
		return err
	}

	keyshares, err := results.toKeyShares()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("keyshares-%d.json", time.Now().Unix())
	fmt.Printf("writing keyshares to file: %s\n", filename)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(keyshares)
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

	// all operators will have same validatorPK in their result
	var firstOperator types.OperatorID
	for k := range results.Output {
		firstOperator = k
		break
	}

	validatorPK, _ := hex.DecodeString(results.Output[firstOperator].Data.ValidatorPubKey)
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

	blsSigBytes, _ := hex.DecodeString(results.Output[firstOperator].Data.DepositDataSignature)
	blsSig := phase0.BLSSignature{}
	copy(blsSig[:], blsSigBytes)
	depositData.Signature = blsSig

	depositDataRoot, _ := depositData.HashTreeRoot()

	response := DepositDataJson{
		PubKey:                results.Output[firstOperator].Data.ValidatorPubKey,
		WithdrawalCredentials: c.String("withdrawal-credentials"),
		Amount:                amount,
		Signature:             results.Output[firstOperator].Data.DepositDataSignature,
		DepositMessageRoot:    hex.EncodeToString(depositMsgRoot[:]),
		DepositDataRoot:       hex.EncodeToString(depositDataRoot[:]),
		ForkVersion:           hex.EncodeToString(fork[:]),
		NetworkName:           c.String("fork-version"),
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
	body, _ := io.ReadAll(resp.Body)
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

func (results *DKGResult) toKeyShares() (*KeyShares, error) {
	if results.Blame != nil {
		return nil, fmt.Errorf("results contains blame output")
	}

	if len(results.Output) == 0 {
		return nil, fmt.Errorf("invalid dkg output")
	}

	operatorData := make([]OperatorData, 0)
	operatorIds := make([]uint32, 0)
	for operatorID := range results.Output {
		od := OperatorData{
			ID: uint32(operatorID),
		}
		operatorIds = append(operatorIds, uint32(operatorID))

		operator, err := storage.GetOperatorFromRegistryByID(operatorID)
		if err != nil {
			return nil, err
		}
		od.PublicKey = operator.PublicKey

		operatorData = append(operatorData, od)
	}

	shares := KeySharesKeys{
		PublicKeys:    make([]string, 0),
		EncryptedKeys: make([]string, 0),
	}

	for _, output := range results.Output {
		shares.PublicKeys = append(shares.PublicKeys, fmt.Sprintf("0x%s", output.Data.SharePubKey))
		shares.EncryptedKeys = append(shares.EncryptedKeys, output.Data.EncryptedShare)
	}

	data := KeySharesData{
		PublicKey: "0x" + results.Output[types.OperatorID(operatorIds[0])].Data.ValidatorPubKey,
		Operators: operatorData,
		Shares:    shares,
	}

	payload := KeySharesPayload{
		Readable: ReadablePayload{
			PublicKey:   "0x" + results.Output[types.OperatorID(operatorIds[0])].Data.ValidatorPubKey,
			OperatorIDs: operatorIds,
			Shares:      sharesToBytes(data.Shares.PublicKeys, shares.EncryptedKeys),
			Amount:      "Amount of SSV tokens to be deposited to your validator's cluster balance (mandatory only for 1st validator in a cluster)",
			Cluster:     "The latest cluster snapshot data, obtained using the cluster-scanner tool. If this is the cluster's 1st validator then use - {0,0,0,0,0,false}",
		},
	}

	return &KeyShares{
		Version:   "v3",
		Data:      data,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}, nil
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

// Convert a slice of strings to a slice of byte slices, where each string is converted to a byte slice
// using hex decoding
func toArrayByteSlices(input []string) [][]byte {
	var result [][]byte
	for _, str := range input {
		bytes, _ := hex.DecodeString(str[2:]) // remove the '0x' prefix and decode the hex string to bytes
		result = append(result, bytes)
	}
	return result
}

func sharesToBytes(publicKeys []string, privateKeys []string) string {
	encryptedShares, _ := decodeEncryptedShares(privateKeys)
	arrayPublicKeys := bytes.Join(toArrayByteSlices(publicKeys), []byte{})
	arrayEncryptedShares := bytes.Join(toArrayByteSlices(encryptedShares), []byte{})

	// public keys hex encoded
	pkHex := hex.EncodeToString(arrayPublicKeys)
	// length of the public keys (hex), hex encoded
	pkHexLength := fmt.Sprintf("%04x", len(pkHex)/2)

	// join arrays
	pkPsBytes := append(arrayPublicKeys, arrayEncryptedShares...)

	// add length of the public keys at the beginning
	// this is the variable that is sent to the contract as bytes, prefixed with 0x
	return "0x" + pkHexLength + hex.EncodeToString(pkPsBytes)
}

func decodeEncryptedShares(encodedEncryptedShares []string) ([]string, error) {
	var result []string
	for _, item := range encodedEncryptedShares {
		// Decode the base64 string
		decoded, err := base64.StdEncoding.DecodeString(item)
		if err != nil {
			return nil, err
		}

		// Encode the decoded bytes as a hexadecimal string with '0x' prefix
		result = append(result, "0x"+hex.EncodeToString(decoded))
	}
	return result, nil
}
