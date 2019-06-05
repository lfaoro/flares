/*
 * Copyright (c) 2019 Leonardo Faoro. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

// ___________.__
// \_   _____/|  | _____ _______   ____   ______
// |    __)  |  | \__  \\_  __ \_/ __ \ /  ___/
// |     \   |  |__/ __ \|  | \/\  ___/ \___ \
// \___  /   |____(____  /__|    \___  >____  >
// \/              \/            \/     \/
//
// Flares is a CloudFlare DNS backup tool.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/lfaoro/flares/internal/cloudflare"
)

var (
	// version is injected during the release process.
	version = "dev"
	// commit is injected during the release process.
	commit = "none"
	// date is injected during the release process.
	date = "unknown"
)

var debugFlag bool

func main() {
	app := cli.NewApp()
	app.Name = "flares"
	app.Usage = "CloudFlare DNS backup tool"
	app.Version = fmt.Sprintf("%s %s %s", version, commit, date)
	app.EnableBashCompletion = true
	app.Authors = []cli.Author{
		{
			Name:  "Leonardo Faoro",
			Email: "lfaoro@gmail.com",
		},
	}

	var flagAPIKey string
	var flagAPIEmail string
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "key",
			Usage:       "CloudFlare API key",
			EnvVar:      "CF_API_KEY",
			Destination: &flagAPIKey,
		},
		cli.StringFlag{
			Name:        "email",
			Usage:       "CloudFlare API email",
			EnvVar:      "CF_API_EMAIL",
			Destination: &flagAPIEmail,
		},
		cli.BoolFlag{
			Name:   "all, a",
			Usage:  "retrieves the records table for all domains",
			EnvVar: "FLARES_ALL",
		},
		cli.BoolFlag{
			Name:   "export, e",
			Usage:  "exports the DNS table into BIND formatted files",
			EnvVar: "FLARES_EXPORT",
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "displays debugFlag information.",
			EnvVar:      "FLARES_DEBUG",
			Hidden:      true,
			Destination: &debugFlag,
		},
	}

	app.Action = func(c *cli.Context) error {
		if flagAPIKey == "" || flagAPIEmail == "" {
			return errors.New(
				"provide --key & --email flags\n" +
					"or $CF_API_EMAIL & $CF_API_EMAIL" + " ENV variables\n" +
					"in order to access CloudFlare.\n\n" +
					"GOTO: https://dash.cloudflare.com/profile -> Global API Key -> View")
		}

		dns := cloudflare.New(flagAPIKey, flagAPIEmail)

		if c.Bool("all") {
			zones, err := dns.Zones()
			fatalIfErr(err)

			wg := sync.WaitGroup{}
			for id, domain := range zones {
				if debugFlag {
					fmt.Printf("ID: %s: domain: %s\n", id, domain)
				}

				wg.Add(1)
				go func(domain string) {
					if debugFlag {
						fmt.Println("closure domain:", domain)
						wg.Done()
						return
					}

					table, err := dns.TableFor(domain)
					fatalIfErr(err)

					if c.Bool("export") {
						export(domain, table)
					} else {
						fmt.Println(string(table))
					}

					wg.Done()
				}(domain)

				wg.Wait()
			}

			return nil
		}

		if c.NArg() < 1 {
			cli.ShowAppHelpAndExit(c, 2)
		}

		for _, domain := range c.Args() {
			if debugFlag {
				log.Println("domain", domain)
				continue
			}

			table, err := dns.TableFor(domain)
			fatalIfErr(err)

			if c.Bool("export") {
				export(domain, table)
				continue
			}

			fmt.Println(string(table))
		}

		return nil
	}

	err := app.Run(os.Args)
	fatalIfErr(err)
}

func export(domain string, data []byte) {
	dir, err := os.Getwd()
	fatalIfErr(err)

	filename := path.Join(dir, domain)
	err = ioutil.WriteFile(filename, data, 0600)
	fatalIfErr(err)

	fmt.Printf("BIND data for %s successfully exported\n", domain)
}

func fatalIfErr(err error) {
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
}
