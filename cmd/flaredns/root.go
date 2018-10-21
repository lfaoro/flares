package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/lfaoro/flares/internal/export"
)

var RootCmd = &cobra.Command{
	Use:   "flaredns",
	Short: "TODO",
	Long:  `flaredns is a tool for reading and exporting DNS tables`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
	PreRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {

	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints program semantic version",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
	PreRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("flaredns | %s\n", Version)
	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
}

var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "outputs DNS table for given domains",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
	PreRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {
		cf := export.Cloudflare{
			API:       "https://api.cloudflare.com/client/v4",
			AuthKey:   cfAuthKey,
			AuthEmail: cfAuthEmail,
			Client: http.Client{
				Timeout: time.Second * 10,
			},
		}

		dir, err := cmd.Flags().GetString("dir")
		log.FatalIfErr(err)
		if dir == "" {
			for _, domain := range args {
				b, err := cf.ExportDNS(domain)
				log.LogIfErr(err)
				fmt.Fprintf(os.Stdout, string(b))
			}
		} else {
			for _, domain := range args {
				table, err := cf.ExportDNS(domain)
				log.LogIfErr(err)
				domain = strings.Replace(domain, ".", "_", -1)
				// TODO: Supressed the error for now, should add proper error checks at some point.
				fullDir, err := filepath.Abs(dir)
				log.FatalIfErr(err)
				filePath := filepath.Join(fullDir, domain+".bind")
				writeFile(table, filePath)
				fmt.Println("Exported:", filePath)
			}
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.FatalIfErr(err)
		os.Exit(1)
	}
}

func init() {
	ExportCmd.Flags().StringP("dir", "d", "", "Path where to export the DNS table files.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
}
