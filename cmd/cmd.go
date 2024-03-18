package cmd

import (
	"log"
	"os"

	"github.com/ssd39/smart-vault-sgx-app/app/entrypoint"
	"github.com/urfave/cli/v2"
)

func Start() {
	app := &cli.App{
		Version: "v1.0",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "init the network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "key",
						Usage: "json file path of private key",
					},
				},
				Action: func(cCtx *cli.Context) error {
					keyPath := cCtx.String("key")
					err := entrypoint.Init(keyPath)
					return err
				},
			},
			{
				Name:  "join",
				Usage: "join the network",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "key",
						Usage: "json file path of private key",
					},
				},
				Action: func(cCtx *cli.Context) error {
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
