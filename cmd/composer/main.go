package main

import (
	"fmt"
	"log"
	"os"

	"github.com/unkaktus/composer"
	"github.com/urfave/cli/v2"
)

func run() error {
	app := &cli.App{
		Name:     "composer",
		HelpName: "composer",
		Usage:    "Command-line tool for CompOSE",
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Ivan Markin",
				Email: "git@unkaktus.art",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "download",
				Usage: "Download EOS data archive from CompOSE",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "id",
						Value: "",
						Usage: "EOS ID",
					},
				},
				Action: func(cCtx *cli.Context) error {
					id := cCtx.String("id")
					if id == "" {
						return fmt.Errorf("EOS ID is not specified")
					}
					err := composer.DownloadEOS(id)
					if err != nil {
						return fmt.Errorf("download EOS: %w", err)
					}
					return nil
				},
			},
		},
	}
	return app.Run(os.Args)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
