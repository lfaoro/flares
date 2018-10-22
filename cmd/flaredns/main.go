// Flaredns backups the DNS records table of all provided Cloudflare domains.
// Optionally commits all your DNS records to a git repository.
package main

import (
	"github.com/vlct-io/pkg/svc"
	"math/rand"
	"os"
	"time"

	"github.com/vlct-io/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	log   = logger.New()
	start = time.Now().UTC().Format(time.RFC3339)
)

// Environment variables
var (
	exportFlag  string
	cfAuthKey  = svc.MustGetEnv("CF_API_KEY")
	cfAuthEmail = svc.MustGetEnv("CF_API_EMAIL")
)

func init() {
	rand.Seed(time.Now().UnixNano())

	cobra.OnInitialize(initConfig)
	RootCmd.AddCommand(VersionCmd)
	RootCmd.AddCommand(ExportCmd)
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
