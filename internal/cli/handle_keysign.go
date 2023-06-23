package cli

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/bloxapp/ssv-spec/types/testingutils"
	"github.com/urfave/cli/v2"
)

func (h *CliHandler) HandleKeySign(c *cli.Context) error {

	keygenRequestID := c.String("keygen-request-id")
	dkgResult, err := h.DKGResultByRequestID(keygenRequestID)
	if err != nil {
		return fmt.Errorf("HandleKeySign: failed to get dkg result for requestID %s: %w", keygenRequestID, err)
	}
	vk, err := dkgResult.GetValidatorPK()
	if err != nil {
		return fmt.Errorf("HandleKeySign: failed to get validatorPK from DKG result: %w", err)
	}

	ownerAddress := c.String("owner-address")
	ownerNonce := c.Int("owner-nonce")
	signingRoot := []byte(fmt.Sprintf("%s:%d", ownerAddress, ownerNonce))

	requestID := getRandRequestID()
	requestIDInHex := hex.EncodeToString(requestID[:])

	keySign := dkg.KeySign{
		ValidatorPK: vk,
		SigningRoot: signingRoot,
	}
	keySignBytes, _ := keySign.Encode()

	initBytes, err := initMsgForKeySign(requestID, keySignBytes)
	if err != nil {
		return fmt.Errorf("HandleKeySign: failed to generate init msg for KeySign: %w", err)
	}

	operators, err := parseOperatorList(c)
	if err != nil {
		return fmt.Errorf("HandleKeySign: failed to parse operator list from command: %w", err)
	}

	ol := make([]types.OperatorID, 0)
	for operatorID := range operators {
		ol = append(ol, operatorID)
	}

	messengerClient := messenger.NewMessengerClient(messenger.MessengerAddrFromEnv())
	if err := messengerClient.CreateTopic(requestIDInHex, ol); err != nil {
		return fmt.Errorf("HandleKeygen: failed to create a new topic on messenger service: %w", err)
	}

	for operatorID, addr := range operators {
		if err := h.sendKeySignMsg(operatorID, addr, initBytes); err != nil {
			return fmt.Errorf("HandleKeySign: failed to send init message to operatorID %d: %w", operatorID, err)
		}
	}

	fmt.Printf("keysign init request sent with ID: %s\n", requestIDInHex)
	return nil
}

func (h *CliHandler) sendKeySignMsg(operatorID types.OperatorID, addr string, data []byte) error {
	url := fmt.Sprintf("%s/consume", addr)
	resp, err := h.client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request to operator %d to consume init message failed with status %s", operatorID, resp.Status)
	}
	return nil
}

func initMsgForKeySign(requestID dkg.RequestID, data []byte) ([]byte, error) {
	ks := testingutils.TestingKeygenKeySet()
	signedInitMsg := testingutils.SignDKGMsg(ks.DKGOperators[1].SK, 1, &dkg.Message{
		MsgType:    dkg.KeySignMsgType,
		Identifier: requestID,
		Data:       data,
	})
	signedInitMsgBytes, _ := signedInitMsg.Encode()

	msg := &types.SSVMessage{
		MsgType: types.DKGMsgType,
		Data:    signedInitMsgBytes,
	}
	return msg.Encode()
}
