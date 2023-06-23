package cli

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
)

type DKGResult struct {
	Output map[types.OperatorID]SignedOutput `json:"output,omitempty"`
	Blame  *dkg.BlameOutput                  `json:"blame,omitempty"`
}

type Output struct {
	RequestID            string `json:"RequestID,omitempty"`
	EncryptedShare       string `json:"EncryptedShare,omitempty"`
	SharePubKey          string `json:"SharePubKey,omitempty"`
	ValidatorPubKey      string `json:"ValidatorPubKey,omitempty"`
	DepositDataSignature string `json:"DepositDataSignature,omitempty"`
}

type KeySignOutput struct {
	RequestID       string `json:"RequestID,omitempty"`
	ValidatorPubKey string `json:"ValidatorPubKey,omitempty"`
	Signature       string `json:"Signature,omitempty"`
}

type SignedOutput struct {
	KeySignData KeySignOutput `json:"KeySignData,omitempty"`
	Data        Output        `json:"Data,omitempty"`
	Signer      string        `json:"Signer,omitempty"`
	Signature   string        `json:"Signature,omitempty"`
}

func (r *DKGResult) GetValidatorPK() (types.ValidatorPK, error) {
	var vk types.ValidatorPK
	for _, output := range r.Output {
		vkbytes, err := hex.DecodeString(output.Data.ValidatorPubKey)
		if err != nil {
			return nil, fmt.Errorf("GetValidatorPK: failed to decode validator PK from its hex value: %w", err)
		}

		if vk != nil {
			if !bytes.Equal(vk, vkbytes) {
				return nil, fmt.Errorf("GetValidatorPK: invalid dkg result, vk from all operators are not equal")
			}
		}

		vk = vkbytes
	}
	return vk, nil
}

func formatResults(data *messenger.DataStore) *DKGResult {
	if data.BlameOutput != nil {
		return formatBlameResults(data.BlameOutput)
	}
	output := make(map[types.OperatorID]SignedOutput)

	for operatorID, signedOutput := range data.DKGOutputs {
		getHex := hex.EncodeToString
		if signedOutput.KeySignData != nil {
			v := SignedOutput{
				KeySignData: KeySignOutput{
					RequestID:       getHex(signedOutput.KeySignData.RequestID[:]),
					Signature:       getHex(signedOutput.KeySignData.Signature),
					ValidatorPubKey: getHex(signedOutput.KeySignData.ValidatorPK),
				},
				Signer:    strconv.Itoa(int(signedOutput.Signer)),
				Signature: hex.EncodeToString(signedOutput.Signature),
			}
			output[operatorID] = v
		} else {
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
	}

	return &DKGResult{Output: output}
}

func formatBlameResults(blameOutput *dkg.BlameOutput) *DKGResult {
	return &DKGResult{Blame: blameOutput}
}
