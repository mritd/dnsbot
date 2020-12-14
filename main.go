package main

import (
	"fmt"
	"os"

	"github.com/mritd/logger"

	"github.com/urfave/cli/v2"
)

var (
	version   string
	buildDate string
	commitID  string
)

func main() {
	app := &cli.App{
		Name:    "dnsbot",
		Usage:   "Telegram Bot for etcdhosts",
		Version: fmt.Sprintf("%s %s %s", version, buildDate, commitID),
		Authors: []*cli.Author{
			{
				Name:  "mritd",
				Email: "mritd@linux.com",
			},
		},
		Copyright: "Copyright (c) 2020 mritd, All rights reserved.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "./dnsbot.yaml",
				Usage:   "dns bot config",
				EnvVars: []string{"DNSBOT_CONFIG"},
			},
			&cli.BoolFlag{
				Name:    "debug",
				Value:   false,
				Usage:   "debug mode",
				EnvVars: []string{"DNSBOT_DEBUG"},
			},
		},
		Action: func(c *cli.Context) error {
			if c.Bool("debug") {
				logger.SetDevelopment()
			}
			cfg := &Config{}
			if err := cfg.LoadFrom(c.String("config")); err != nil {
				return err
			}
			serve(cfg)
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
