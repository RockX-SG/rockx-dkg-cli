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

	"github.com/RockX-SG/frost-dkg-demo/internal/logger"
	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/sirupsen/logrus"
)

type CliHandler struct {
	client        *http.Client
	logger        *logger.Logger
	messengerAddr string
}

func New(logger *logger.Logger) *CliHandler {
	logger.WithFields(logrus.Fields{"messenger-server-address": messenger.MessengerAddrFromEnv()}).
		Debug("created new cli handler")

	return &CliHandler{
		client: &http.Client{
			Timeout: 5 * time.Minute,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		logger:        logger,
		messengerAddr: messenger.MessengerAddrFromEnv(),
	}
}

func (h *CliHandler) DKGResultByRequestID(requestID string) (*DKGResult, error) {

	log := h.logger.WithFields(logrus.Fields{"request-id": requestID})
	log.Debug("fetching dkg results for keygen/resharing")

	resp, err := h.client.Get(fmt.Sprintf("%s/data/%s", h.messengerAddr, requestID))
	if err != nil {
		log.Errorf("failed to request messenger server for dkg result: %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Errorf("failed to fetch keygen/resharing results with status %s", resp.Status)
		log.Debugf("request failed with body %s", string(respBody))
		return nil, fmt.Errorf("failed to fetch dkg result for request %s with code %d", requestID, resp.StatusCode)
	}

	data := &messenger.DataStore{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorf("failed to parse response json: %s", err.Error())
		return nil, err
	}

	return formatResults(data), nil
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
