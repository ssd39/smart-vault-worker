package cmd

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/ssd39/smart-vault-sgx-app/app/entrypoint"
	"github.com/ssd39/smart-vault-sgx-app/app/ipfs"
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
					&cli.StringFlag{
						Name:     "ipfs",
						Usage:    "ipfs uploader service details",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					keyPath := cCtx.String("key")
					ipfsUplaoderArg := cCtx.String("ipfs")
					ipfsServiceDetails := strings.Split(ipfsUplaoderArg, ":")

					if len(ipfsServiceDetails) <= 0 {
						return errors.New("Ipfs uploder required to init")
					}

					var ipfsUploader ipfs.IpfsUploader

					if ipfsServiceDetails[0] == "filebase" {
						// Example: "filebase:{AccesKey}:{SecretKey}:{BucketName}"
						if len(ipfsServiceDetails) < 4 {
							return errors.New("Un-sufficient arguments for filebase ipfs uploader")
						}
						ipfsUploader = &ipfs.FilebaseUploader{
							AccessKey: ipfsServiceDetails[1],
							SecertKey: ipfsServiceDetails[2],
							Bucket:    ipfsServiceDetails[3],
							Name:      "init-attestation",
						}
					} else {
						return errors.New("Unkonw ipfs uploader provided")
					}
					err := entrypoint.Init(keyPath, ipfsUploader)
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
			{

				Name:  "start",
				Usage: "start the worker",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "key",
						Usage:    "json file path of private key",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					keyPath := cCtx.String("key")
					return entrypoint.Start(keyPath)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
