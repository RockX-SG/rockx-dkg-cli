package messenger

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
)

type Client struct {
	SrvAddr string
}

func NewMessengerClient(srvAddr string) *Client {
	return &Client{
		SrvAddr: srvAddr,
	}
}

func (cl *Client) StreamDKGBlame(blame *dkg.BlameOutput) error {
	requestID := hex.EncodeToString(blame.BlameMessage.Message.Identifier[:])
	data, err := json.Marshal(blame)
	if err != nil {
		return err
	}

	return cl.stream("dkgoutput", requestID, data)
}
func (cl *Client) StreamDKGOutput(output map[types.OperatorID]*dkg.SignedOutput) error {
	var requestID string

	// assuming all signed output have same identifier. skipping validation here
	for _, output := range output {
		requestID = hex.EncodeToString(output.Data.RequestID[:])
	}

	data, err := json.Marshal(output)
	if err != nil {
		return err
	}
	return cl.stream("dkgblame", requestID, data)
}
func (cl *Client) BroadcastDKGMessage(msg *dkg.SignedMessage) error {
	data, err := msg.Encode()
	if err != nil {
		return err
	}

	return cl.publish(DefaultTopic, data)
}

func (cl *Client) publish(topicName string, data []byte) error {
	msg := types.SSVMessage{
		MsgType: types.DKGMsgType,
		Data:    data,
	}

	byts, _ := msg.Encode()

	resp, err := http.Post(fmt.Sprintf("%s/publish?topic_name=%s", cl.SrvAddr, DefaultTopic), "application/json", bytes.NewReader(byts))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to call publish request to messenger")
	}
	return nil
}

func (cl *Client) stream(urlparam string, requestID string, data []byte) error {
	resp, err := http.Post(fmt.Sprintf("%s/stream/%s?request_id=%s", cl.SrvAddr, urlparam, requestID), "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to call stream %s request to messenger", urlparam)
	}
	return nil
}
