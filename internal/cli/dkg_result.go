package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/RockX-SG/frost-dkg-demo/internal/storage"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
)

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

func formatResults(data *messenger.DataStore) *DKGResult {
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

	return &DKGResult{Output: output}
}

func formatBlameResults(blameOutput *dkg.BlameOutput) *DKGResult {
	return &DKGResult{Blame: blameOutput}
}
