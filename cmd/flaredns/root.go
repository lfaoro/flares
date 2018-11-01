package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lfaoro/flares/internal/cloud"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	exportFlag string
	allFlag    bool
	keyFlag    string
	emailFlag  string

	client cloud.Cloudflare
)

var rootCmd = &cobra.Command{
	Use:          "flaredns [DOMAIN]",
	Short:        "flaredns is a CloudFlare DNS backup tool.",
	RunE:         flaredns,
	SilenceUsage: true,
	Version:      Version,
	Example:      "flaredns google.com --export /tmp/dnsdump",
}

func init() {
	cobra.OnInitialize(initClient)

	rootCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return errors.Wrap(err, "🔥")
	})

	rootCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "retrieve DNS records table for all domains.")
	rootCmd.Flags().StringVarP(&exportFlag, "export", "e", "", "export the DNS table into BIND formatted files.")
	rootCmd.Flags().StringVarP(&keyFlag, "key", "k", "", "CloudFlare API key (defaults to $CF_API_KEY)")
	rootCmd.Flags().StringVarP(&emailFlag, "email", "m", "", "CloudFlare API email (defaults to $CF_API_EMAIL)")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func flaredns(cmd *cobra.Command, args []string) error {
	if allFlag {
		zones, err := client.AllZones()
		if err != nil {
			return errors.Wrap(err, "🔥")
		}
		for _, zone := range zones {
			split := strings.Split(zone, ",")[1]
			table, err := client.DNSTableFor(split)
			if err != nil {
				return errors.Wrap(err, "🔥")
			}
			fmt.Fprintf(os.Stdout, string(table))
		}
		return nil
	}
	if exportFlag != "" {
		dir := exportFlag
		for _, domain := range args {
			table, err := client.DNSTableFor(domain)
			if err != nil {
				return errors.Wrap(err, "🔥")
			}
			domain = strings.Replace(domain, ".", "_", -1)
			fullDir, err := filepath.Abs(dir)
			if err != nil {
				return errors.Wrap(err, "🔥")
			}
			if err := os.MkdirAll(fullDir, 0755); err != nil {
				return errors.Wrap(err, "🔥")
			}
			filePath := filepath.Join(fullDir, domain+".bind")
			writeFile(table, filePath)
			fmt.Println("🔥  Exported:", filePath)
		}
		return nil
	}
	if len(args) == 0 {
		cmd.Usage()
	}
	for _, domain := range args {
		b, err := client.DNSTableFor(domain)
		if err != nil {
			return errors.Wrap(err, "🔥")
		}
		fmt.Fprintf(os.Stdout, string(b))
	}
	return nil
}

func initClient() {
	if keyFlag == "" || emailFlag == "" {
		keyFlag = os.Getenv("CF_API_KEY")
		emailFlag = os.Getenv("CF_API_EMAIL")
	}
	client = cloud.Cloudflare{
		API:       "https://api.cloudflare.com/client/v4",
		AuthKey:   keyFlag,
		AuthEmail: emailFlag,
		Client: http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func writeFile(data []byte, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "🔥")
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return errors.Wrap(err, "🔥")
	}
	return nil
}
