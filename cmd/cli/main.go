package main

import (
	"log"
	"os"

	clihandler "github.com/RockX-SG/frost-dkg-demo/internal/cli"
	"github.com/urfave/cli/v2"
)

func main() {
	h := clihandler.New()
	app := &cli.App{
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
						Name:    "threshold",
						Aliases: []string{"t"},
						Usage:   "threshold value",
						Value:   3,
					},
					&cli.StringFlag{
						Name:    "withdrawal",
						Aliases: []string{"w"},
						Usage:   "withdrawal credential",
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "fork",
						Aliases: []string{"f"},
						Usage:   "fork version",
						Value:   "",
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
						Aliases:  []string{"p"},
						Usage:    "old operator key-value pair",
						Required: true,
					},
					&cli.IntFlag{
						Name:    "threshold",
						Aliases: []string{"t"},
						Usage:   "threshold value",
						Value:   3,
					},
					&cli.StringFlag{
						Name:    "validator-pk",
						Aliases: []string{"vk"},
						Usage:   "validator public key value",
						Value:   "",
					},
				},
			},
			{
				Name:    "get-results",
				Aliases: []string{"gr"},
				Usage:   "get results of keygen/resharing request",
				Action:  h.HandleGetData,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "request-id",
						Aliases:  []string{"2"},
						Usage:    "request id for keygen/resharing",
						Required: true,
					},
				},
			},
			// {
			// 	Name:    "generate-deposit-data",
			// 	Aliases: []string{"gdd"},
			// 	Usage:   "generate deposit data in json format",
			// 	Action:  h.HandleGetDepositData,
			// },
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
