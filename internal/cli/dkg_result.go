package cli

import (
	"encoding/hex"
	"strconv"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
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
