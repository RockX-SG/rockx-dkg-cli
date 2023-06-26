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

package main

import (
	"fmt"
	"log"
	"os"

	clihandler "github.com/RockX-SG/frost-dkg-demo/internal/cli"
	"github.com/RockX-SG/frost-dkg-demo/internal/logger"
	"github.com/urfave/cli/v2"
)

var version string

func main() {
	basePath := "/var/log"
	if os.Getenv("DKG_LOG_PATH") != "" {
		basePath = os.Getenv("DKG_LOG_PATH")
	}

	logger := logger.New(fmt.Sprintf("%s/dkg_cli.log", basePath))
	h := clihandler.New(logger)

	app := &cli.App{
		Name:  "rockx-dkg-cli",
		Usage: "A cli tool to run DKG for keygen and resharing and generate deposit data",
		Commands: []*cli.Command{
			{
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
			},
			{
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
			},
			{
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
			},
			{
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
				},
			},
			{
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
			},
			{
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
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "cli version",
				Action: func(ctx *cli.Context) error {
					fmt.Println(version)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
