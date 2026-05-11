// Command flares is a Cloudflare DNS backup tool.
// It exports DNS records as BIND-formatted zone files to stdout or disk.
//
// SPDX-License-Identifier: MIT
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/lfaoro/flares/internal/cloudflare"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := newCmd()
	if err := cmd.Run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func newCmd() *cli.Command {
	return &cli.Command{
		Name:                  "flares",
		Usage:                 "Cloudflare DNS backup tool",
		Version:               fmt.Sprintf("%s (commit=%s, built=%s)", version, commit, date),
		EnableShellCompletion: true,
		Authors: []any{
			"Leonardo Faoro <flares@leonardofaoro.com>",
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "Cloudflare API `TOKEN` (create at https://dash.cloudflare.com/profile/api-tokens with Zone.DNS read permission)",
				Sources: cli.NewValueSourceChain(
					cli.EnvVar("CLOUDFLARE_API_TOKEN"),
					cli.EnvVar("CF_API_TOKEN"),
				),
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "enable debug output",
				Sources: cli.EnvVars("FLARES_DEBUG"),
			},
			&cli.IntFlag{
				Name:    "threads",
				Aliases: []string{"T"},
				Usage:   "max concurrent API requests for --all operations",
				Value:   10,
				Sources: cli.EnvVars("FLARES_THREADS"),
			},
			&cli.StringFlag{
				Name:   "api-url",
				Usage:  "Cloudflare API base `URL` (for testing)",
				Value:  "https://api.cloudflare.com/client/v4",
				Hidden: true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "show",
				Aliases:   []string{"s"},
				Usage:     "Display DNS records for one or more domains",
				ArgsUsage: "[<domain>...]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "retrieve records for all domains",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "output `FORMAT` (text, json)",
						Value:   "text",
					},
				},
				Action: showAction,
			},
			{
				Name:      "export",
				Aliases:   []string{"e"},
				Usage:     "Export DNS records to BIND-formatted zone files",
				ArgsUsage: "[<domain>...]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "export records for all domains",
					},
				},
				Action: exportAction,
			},
			{
				Name:    "zones",
				Aliases: []string{"z"},
				Usage:   "List all zones in the account",
				Action:  zonesAction,
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			return cli.ShowRootCommandHelp(cmd)
		},
	}
}

func clientFromCmd(cmd *cli.Command) (*cloudflare.Client, error) {
	token := cmd.String("token")
	if token == "" {
		return nil, fmt.Errorf("provide --token flag or $CLOUDFLARE_API_TOKEN\n" +
			"  Create at: https://dash.cloudflare.com/profile/api-tokens\n" +
			"  Required scope: Zone.DNS -> Read")
	}
	cf, err := cloudflare.New(token)
	if err != nil {
		return nil, err
	}
	if apiURL := cmd.String("api-url"); apiURL != "" {
		cf.SetBaseURL(apiURL)
	}
	return cf, nil
}

func isDebug(cmd *cli.Command) bool {
	return cmd.Bool("debug")
}

func showAction(ctx context.Context, cmd *cli.Command) error {
	cf, err := clientFromCmd(cmd)
	if err != nil {
		return err
	}

	jsonOut := cmd.String("output") == "json"

	domains, err := resolveDomains(ctx, cmd, cf)
	if err != nil {
		return err
	}

	if jsonOut {
		records := map[string]string{}
		for _, d := range domains {
			if isDebug(cmd) {
				fmt.Println("domain:", d)
				continue
			}
			table, err := cf.Export(ctx, d)
			if err != nil {
				return fmt.Errorf("%s: %w", d, err)
			}
			records[d] = string(table)
		}
		return json.NewEncoder(os.Stdout).Encode(records)
	}

	for _, domain := range domains {
		if isDebug(cmd) {
			fmt.Println("domain:", domain)
			continue
		}
		table, err := cf.Export(ctx, domain)
		if err != nil {
			return fmt.Errorf("%s: %w", domain, err)
		}
		fmt.Println(string(table))
	}

	return nil
}

func exportAction(ctx context.Context, cmd *cli.Command) error {
	cf, err := clientFromCmd(cmd)
	if err != nil {
		return err
	}

	domains, err := resolveDomains(ctx, cmd, cf)
	if err != nil {
		return err
	}

	sem := make(chan struct{}, cmd.Int("threads"))
	var wg sync.WaitGroup
	errs := make(chan error, len(domains))

	for _, domain := range domains {
		wg.Add(1)
		sem <- struct{}{}
		go func(domain string) {
			defer wg.Done()
			defer func() { <-sem }()

			if isDebug(cmd) {
				fmt.Println("domain:", domain)
				return
			}

			table, err := cf.Export(ctx, domain)
			if err != nil {
				errs <- fmt.Errorf("%s: %w", domain, err)
				return
			}
			if err := writeFile(domain, table); err != nil {
				errs <- fmt.Errorf("%s: %w", domain, err)
				return
			}
			fmt.Printf("BIND data for %s successfully exported\n", domain)
		}(domain)
	}

	wg.Wait()
	close(errs)

	for e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

func zonesAction(ctx context.Context, cmd *cli.Command) error {
	cf, err := clientFromCmd(cmd)
	if err != nil {
		return err
	}

	zones, err := cf.Zones(ctx)
	if err != nil {
		return fmt.Errorf("list zones: %w", err)
	}
	for id, name := range zones {
		fmt.Printf("%s  %s\n", id, name)
	}
	return nil
}

func resolveDomains(ctx context.Context, cmd *cli.Command, cf *cloudflare.Client) ([]string, error) {
	if cmd.Bool("all") {
		zones, err := cf.Zones(ctx)
		if err != nil {
			return nil, fmt.Errorf("list zones: %w", err)
		}
		domains := make([]string, 0, len(zones))
		for _, name := range zones {
			domains = append(domains, name)
		}
		return domains, nil
	}

	if cmd.NArg() < 1 {
		return nil, fmt.Errorf("at least one domain required (or use --all)")
	}
	return cmd.Args().Slice(), nil
}

func writeFile(domain string, data []byte) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, domain), data, 0600)
}
