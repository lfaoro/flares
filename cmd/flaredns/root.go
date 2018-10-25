package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lfaoro/flares/internal/cloud"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	client cloud.Cloudflare

	cfgFileFlag string
	exportFlag  string
	allFlag     bool
)

var rootCmd = &cobra.Command{
	Use:     "flaredns",
	Short:   "flaredns is a CloudFlare DNS backup tool.",
	Version: Version,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.RunE = flaredns
	rootCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "retrieve DNS records table for all domains.")
	rootCmd.Flags().StringVarP(&exportFlag, "export", "e", "", "export the DNS table into BIND formatted files.")
}

func flaredns(cmd *cobra.Command, args []string) error {
	if allFlag {
		zones, err := client.AllZones()
		if err != nil {
			return err
		}
		for _, zone := range zones {
			split := strings.Split(zone, ",")[1]
			table, err := client.DNSTableFor(split)
			if err != nil {
				return err
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
				return err
			}
			domain = strings.Replace(domain, ".", "_", -1)
			fullDir, err := filepath.Abs(dir)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(fullDir, 0755); err != nil {
				return err
			}
			filePath := filepath.Join(fullDir, domain+".bind")
			writeFile(table, filePath)
			fmt.Println("Exported:", filePath)
		}
		return nil
	}
	if len(args) == 0 {
		cmd.Usage()
	}
	for _, domain := range args {
		b, err := client.DNSTableFor(domain)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, string(b))
	}
	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFileFlag != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFileFlag)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Search config in home directory with name ".flaredns" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".flaredns")
	}
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	viper.AutomaticEnv() // read in environment variables that match
	ok := viper.IsSet("CF_API_KEY")
	ok = viper.IsSet("CF_API_EMAIL")
	if !ok {
		fmt.Println("Missing required environment variables.")
		os.Exit(1)
	}
	client = cloud.Cloudflare{
		API:       "https://api.cloudflare.com/client/v4",
		AuthKey:   viper.GetString("CF_API_KEY"),
		AuthEmail: viper.GetString("CF_API_EMAIL"),
		Client: http.Client{
			Timeout: time.Second * 10,
		},
	}

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func writeFile(data []byte, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}
