// Flaredns backups the DNS records table of all provided Cloudflare domains.
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

	"git.vlct.io/vaulter/vaulter/pkg/logger"

	"github.com/gobuffalo/envy"
	"github.com/lfaoro/flares/internal/export"
	"github.com/lfaoro/flares/internal/vcs"
)

var (
	log           = logger.New()
	start         = time.Now().UTC().Format(time.RFC3339)
	timeFrameFlag = flag.Duration("timeframe", 0, "Every how often to execute the program operations.")
	domainsFlag   = flag.String("domains", "", "Comma(,) separated list of domains from which to export the BIND formatted DNS records table.")
	versionFlag   = flag.Bool("version", false, "Program semantic version.")
)

// Environment variables
var (
	envErr      error
	cfAuthKey   string
	cfAuthEmail string
	gitRepo     string
	gitUsername string
	gitPassword string
)

func init() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	// load environment variables
	log.FatalIfErr(envy.Load())
	cfAuthKey, envErr = envy.MustGet("CF_AUTH_KEY")
	cfAuthEmail, envErr = envy.MustGet("CF_AUTH_EMAIL")
	gitRepo, envErr = envy.MustGet("GIT_REPO")
	gitUsername, envErr = envy.MustGet("GIT_USERNAME")
	gitPassword, envErr = envy.MustGet("GIT_PASSWORD")
	log.FatalIfErr(envErr)
}

func main() {
	if *versionFlag {
		fmt.Println(Version)
		os.Exit(0)
	}
	if *domainsFlag == "" {
		fmt.Println(`
Error: no domains provided
	`)
		flag.Usage()
		os.Exit(1)
	}
	domains := parseDomains(*domainsFlag)

	cf := export.Cloudflare{
		API:       "https://api.cloudflare.com/client/v4",
		AuthKey:   cfAuthKey,
		AuthEmail: cfAuthEmail,
		Client: http.Client{
			Timeout: time.Second * 10,
		},
	}
	files, err := createTables(domains, cf)
	log.FatalIfErr(err)

	gitDir := os.TempDir() + "cf_dns_bck/"
	repo := &vcs.Git{
		Repository: gitRepo,
		Username:   gitUsername,
		Password:   gitPassword,
		Directory:  gitDir,
	}
	log.FatalIfErr(repo.Clone())
	for _, f := range files {
		log.LogIfErr(repo.Add(f))
	}
	log.FatalIfErr(repo.Push())
	// cleanup
	os.RemoveAll(gitDir)
}

func parseDomains(d string) []string {
	return strings.Split(d, ",")
}

func createTables(domains []string, cf export.Cloudflare) (files []string, err error) {
	filesDir := filepath.Join(os.TempDir(), "tables")
	if err := os.MkdirAll(filesDir, 0750); err != nil {
		return nil, err
	}
	for _, domain := range domains {
		filePath := filepath.Join(filesDir, domain+".bind")
		table, err := cf.Export(domain)
		if err != nil {
			return nil, err
		}
		domain = strings.Replace(domain, ".", "_", -1)
		file, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		_, err = file.Write(table)
		if err != nil {
			return nil, err
		}
		files = append(files, filePath)
	}
	return files, nil
}
