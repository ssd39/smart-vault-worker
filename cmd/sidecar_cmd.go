package cmd

import (
	"log"
	"os"

	"github.com/ssd39/smart-vault-sgx-app/app/sidecar"
	"github.com/urfave/cli/v2"
)

func StartSidecar() {
	app := &cli.App{
		Version: "v1.0",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "run the sidecar",
				Flags: []cli.Flag{},
				Action: func(cCtx *cli.Context) error {
					sidecar.StartListner()
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
