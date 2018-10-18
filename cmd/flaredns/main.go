// Flaredns backups the DNS records table of all provided Cloudflare domains.
// Optionally commits all your DNS records to a git repository.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vlct-io/pkg/logger"

	"github.com/gobuffalo/envy"
	"github.com/lfaoro/flares/internal/export"
)

var (
	log         = logger.New()
	start       = time.Now().UTC().Format(time.RFC3339)
	exportFlag  = flag.String("export", "", "Path where to export the DNS table files.")
	versionFlag = flag.Bool("version", false, "Program semantic version.")
	// TODO: separate concerns add `-backup` flag
)

// Environment variables
var (
	envErr      error
	cfAuthKey   string
	cfAuthEmail string
)

func init() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	// load environment variables
	envy.Load()
	cfAuthKey, envErr = envy.MustGet("CF_AUTH_KEY")
	cfAuthEmail, envErr = envy.MustGet("CF_AUTH_EMAIL")
}

func main() {
	if *versionFlag {
		fmt.Printf("flaredns | %s\n", Version)
		os.Exit(0)
	}
	domains := flag.Args()
	if len(domains) == 0 {
		fmt.Println(`Error: no domains provided`)
		flag.Usage()
		os.Exit(1)
	}

	cf := export.Cloudflare{
		API:       "https://api.cloudflare.com/client/v4",
		AuthKey:   cfAuthKey,
		AuthEmail: cfAuthEmail,
		Client: http.Client{
			Timeout: time.Second * 10,
		},
	}

	if exportFlag != nil {
		filesDir := filepath.Dir(*exportFlag)
		if err := os.MkdirAll(filesDir, 0750); err != nil {
			log.FatalIfErr(err)
		}
		for _, domain := range domains {
			table, err := cf.ExportDNS(domain)
			log.LogIfErr(err)
			domain = strings.Replace(domain, ".", "_", -1)
			filePath := filepath.Join(filesDir, domain+".bind")
			writeFile(table, filePath)
			fmt.Println("Exported:", filePath)
		}
		return
	}

	for _, domain := range domains {
		b, err := cf.ExportDNS(domain)
		log.LogIfErr(err)
		fmt.Fprintf(os.Stdout, string(b))
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
