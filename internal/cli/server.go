package cli

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/sirupsen/logrus"
)

type CliHandler struct {
	client *http.Client
	logger *logrus.Logger
}

func New(logger *logrus.Logger) *CliHandler {
	return &CliHandler{
		client: &http.Client{
			Timeout: 5 * time.Minute,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		logger: logger,
	}
}

func (h *CliHandler) DKGResultByRequestID(requestID string) (*DKGResult, error) {
	resp, err := h.client.Get(fmt.Sprintf("%s/data/%s", messenger.MessengerAddrFromEnv(), requestID))
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

func writeJSON(filepath string, data any) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(data)
}
