/*
 * ==================================================================
 *Copyright (C) 2022-2023 Altstake Technology Pte. Ltd. (RockX)
 *This file is part of rockx-dkg-cli <https://github.com/RockX-SG/rockx-dkg-cli>
 *CAUTION: THESE CODES HAVE NOT BEEN AUDITED
 *
 *rockx-dkg-cli is free software: you can redistribute it and/or modify
 *it under the terms of the GNU General Public License as published by
 *the Free Software Foundation, either version 3 of the License, or
 *(at your option) any later version.
 *
 *rockx-dkg-cli is distributed in the hope that it will be useful,
 *but WITHOUT ANY WARRANTY; without even the implied warranty of
 *MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *GNU General Public License for more details.
 *
 *You should have received a copy of the GNU General Public License
 *along with rockx-dkg-cli. If not, see <http://www.gnu.org/licenses/>.
 *==================================================================
 */

package cli

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/RockX-SG/frost-dkg-demo/internal/messenger"
	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type CliHandler struct {
	client        *http.Client
	logger        *logrus.Logger
	messengerAddr string
}

func New(logger *logrus.Logger) *CliHandler {
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

func (h CliHandler) CommandKeygen() *cli.Command {
	return &cli.Command{
		Name:    "keygen",
		Aliases: []string{"k"},
		Usage:   "start keygen process",
		Action:  h.HandleKeygen,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "operator",
				Aliases:  []string{"o"},
				Usage:    "operator key-value pair",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "threshold",
				Aliases:  []string{"t"},
				Usage:    "threshold value",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "withdrawal-credentials",
				Aliases:  []string{"w"},
				Usage:    "withdrawal credential value",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "fork-version",
				Aliases:  []string{"f"},
				Usage:    "fork version",
				Required: true,
			},
		},
	}
}

func (h CliHandler) CommandResharing() *cli.Command {
	return &cli.Command{
		Name:    "resharing",
		Aliases: []string{"r"},
		Usage:   "start resharing process",
		Action:  h.HandleResharing,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "operator",
				Aliases:  []string{"o"},
				Usage:    "operator key-value pair",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "old-operator",
				Aliases:  []string{"oo"},
				Usage:    "old operator key-value pair",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "threshold",
				Aliases:  []string{"t"},
				Usage:    "threshold value",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "validator-pk",
				Aliases:  []string{"vk"},
				Usage:    "validator public key value",
				Required: true,
			},
		},
	}
}

func (h CliHandler) CommandGetDKGResults() *cli.Command {
	return &cli.Command{
		Name:    "get-dkg-results",
		Aliases: []string{"gr"},
		Usage:   "get validator-pk and key shares data for all operators",
		Action:  h.HandleGetData,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "request-id",
				Aliases:  []string{"req"},
				Usage:    "request id for keygen/resharing",
				Required: true,
			},
		},
	}
}

func (h CliHandler) CommandGetKeyshares() *cli.Command {
	return &cli.Command{
		Name:    "get-keyshares",
		Aliases: []string{"gks"},
		Usage:   "generates a keyshare for registering the validator on ssv UI",
		Action:  h.HandleGetKeyShares,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "request-id",
				Aliases:  []string{"req"},
				Usage:    "request id for keygen/resharing",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "operator",
				Aliases:  []string{"o"},
				Usage:    "operator key-value pair",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "owner-address",
				Aliases:  []string{"oa"},
				Usage:    "The cluster owner address (in the SSV contract)",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "owner-nonce",
				Aliases:  []string{"on"},
				Usage:    "The validator registration nonce of the account (owner address) within the SSV contract (increments after each validator registration), obtained using the ssv-scanner tool.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "network",
				Aliases:  []string{"net"},
				Usage:    "ETH network: prater, holesky, mainnet",
				Required: true,
			},
		},
	}
}

func (h CliHandler) CommandGenerateDepositData() *cli.Command {
	return &cli.Command{
		Name:    "generate-deposit-data",
		Aliases: []string{"gdd"},
		Usage:   "generate deposit data in json format",
		Action:  h.HandleGetDepositData,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "request-id",
				Aliases:  []string{"req"},
				Usage:    "request id for keygen/resharing",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "withdrawal-credentials",
				Aliases:  []string{"w"},
				Usage:    "withdrawal credential",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "fork-version",
				Aliases:  []string{"f"},
				Usage:    "fork version",
				Required: true,
			},
		},
	}
}

func (h *CliHandler) DKGResultByRequestID(requestID string) (*DKGResult, error) {
	log := h.logger.WithFields(logrus.Fields{"request-id": requestID})
	log.Debug("DKGResultByRequestID: fetching dkg results for keygen/resharing")

	resp, err := h.client.Get(fmt.Sprintf("%s/data/%s", h.messengerAddr, requestID))
	if err != nil {
		log.Errorf("failed to request messenger server for dkg result: %s", err.Error())
		return nil, fmt.Errorf("DKGResultByRequestID: failed to request messenger server for dkg result %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Errorf("failed to fetch keygen/resharing results with status %s", resp.Status)
		log.Debugf("request failed with body %s", string(respBody))
		return nil, fmt.Errorf("DKGResultByRequestID: failed to fetch dkg result for request %s with code %d", requestID, resp.StatusCode)
	}

	data := &messenger.DataStore{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorf("failed to parse response json: %s", err.Error())
		return nil, fmt.Errorf("DKGResultByRequestID: failed to parse dkg result from api response")
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
