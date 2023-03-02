package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/bloxapp/ssv-spec/types"
)

func createTopic(requestID string, l []types.OperatorID) error {
	messengerAddr := messengerAddrFromEnv()

	topic := messenger.CreateTopicReq{
		TopicName:   requestID,
		Subscribers: make([]string, 0),
	}
	for _, operatorID := range l {
		topic.Subscribers = append(topic.Subscribers, strconv.Itoa(int(operatorID)))
	}
	data, _ := json.Marshal(topic)

	resp, err := http.Post(fmt.Sprintf("%s/topics", messengerAddr), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to call createTopic on messenger")
	}
	return nil
}

func messengerAddrFromEnv() string {
	messengerAddr := os.Getenv("MESSENGER_SRV_ADDR")
	if messengerAddr == "" {
		messengerAddr = "http://0.0.0.0:3000"
	}
	return messengerAddr
}
