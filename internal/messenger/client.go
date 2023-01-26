package messenger

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
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

	return cl.stream("dkgblame", requestID, data)
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
	return cl.stream("dkgoutput", requestID, data)
}

func (cl *Client) BroadcastDKGMessage(msg *dkg.SignedMessage) error {
	msgBytes, err := msg.Encode()
	if err != nil {
		return err
	}

	ssvMsg := types.SSVMessage{
		MsgType: types.DKGMsgType,
		Data:    msgBytes,
	}
	ssvMsgBytes, _ := ssvMsg.Encode()

	fmt.Printf("signer %d, requestID %s\n", msg.Signer, hex.EncodeToString(msg.Message.Identifier[:]))
	requestID := hex.EncodeToString(msg.Message.Identifier[:])
	return cl.publish(requestID, ssvMsgBytes)
}

func (cl *Client) RegisterOperatorNode(id, addr string) error {
	numtries := 3
	try := 1

	for ; try <= numtries; try++ {
		sub := &Subscriber{
			Name:    id,
			SrvAddr: addr,
		}
		byts, _ := json.Marshal(sub)

		url := fmt.Sprintf("%s/register_node?subscribes_to=%s", cl.SrvAddr, DefaultTopic)
		resp, err := http.Post(url, "application/json", bytes.NewReader(byts))
		if err != nil {
			err := fmt.Errorf("failed to make request to messenger")
			log.Printf("Error: %s\n", err.Error())
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err := fmt.Errorf("failed to register operator of ID %s with the messenger on %d try", sub.Name, try)
			log.Printf("Error: %s\n", err.Error())
		} else {
			break
		}
	}

	if try > numtries {
		return fmt.Errorf("failed to register this node even after %d tried", numtries)
	}
	return nil
}

func (cl *Client) publish(topicName string, data []byte) error {
	resp, err := http.Post(fmt.Sprintf("%s/publish?topic_name=%s", cl.SrvAddr, topicName), "application/json", bytes.NewBuffer(data))
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
	resp, err := http.Post(fmt.Sprintf("%s/stream/%s?request_id=%s", cl.SrvAddr, urlparam, requestID), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to call stream %s request to messenger", urlparam)
	}
	return nil
}
