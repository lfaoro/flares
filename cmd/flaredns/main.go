// Flaredns backups the DNS records table of all provided Cloudflare domains.
// Optionally commits all your DNS records to a git repository.
package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/vlct-io/pkg/logger"

	"github.com/gobuffalo/envy"
	"github.com/spf13/cobra"
)

var (
	log   = logger.New()
	start = time.Now().UTC().Format(time.RFC3339)
)

// Environment variables
var (
	exportFlag  string
	envErr      error
	cfAuthKey   string
	cfAuthEmail string
)

func init() {
	rand.Seed(time.Now().UnixNano())

	envy.Load()
	cfAuthKey, envErr = envy.MustGet("CF_AUTH_KEY")
	cfAuthEmail, envErr = envy.MustGet("CF_AUTH_EMAIL")

	cobra.OnInitialize(initConfig)
	RootCmd.AddCommand(VersionCmd)
	RootCmd.AddCommand(ExportCmd)
	return
}

func main() {

	// This should be all that's needed after everything gets converted.
	Execute()
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
