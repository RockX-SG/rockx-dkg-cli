package main

import (
	"fmt"
	"log"
	"os"

	clihandler "github.com/RockX-SG/frost-dkg-demo/internal/cli"
	"github.com/RockX-SG/frost-dkg-demo/internal/logger"
	"github.com/urfave/cli/v2"
)

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
				},
			},
			{
				Name:    "keysign",
				Aliases: []string{"ks"},
				Action:  h.HandleKeySign,
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "operator",
						Aliases:  []string{"o"},
						Usage:    "operator key-value pair",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "keygen-request-id",
						Aliases:  []string{"req"},
						Usage:    "request id from a previous keygen/resharing",
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
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
