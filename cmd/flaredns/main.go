package main

import (
	"flag"
	"math/rand"
	"net/http"
	"os"
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
	develFlag     = flag.Bool("devel", false, "Activates development mode.")
	timeFrameFlag = flag.Duration("timeframe", time.Second*5, "Every how often to execute the program operations.")
	domainsFlag   = flag.String("domains", "vlct.io", "Comma(,) separated list of domains from which to export the BIND formatted DNS records table.")
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
	cf := export.Cloudflare{
		API:       "https://api.cloudflare.com/client/v4",
		AuthKey:   cfAuthKey,
		AuthEmail: cfAuthEmail,
		Client: http.Client{
			Timeout: time.Second * 10,
		},
	}
	domains := parseDomains(*domainsFlag)
	domain := domains[0]
	table, err := cf.Export(domain)
	log.FatalIfErr(err)
	domain = strings.Replace(domain, ".", "_", -1)

	// create file
	file, _ := os.Create(domain + ".bind")
	defer file.Close()
	file.Write(table)
	filePath := "./" + file.Name()

	repo := &vcs.Git{
		Repository: gitRepo,
		Username:   gitUsername,
		Password:   gitPassword,
		Directory:  os.TempDir() + "cf_dns_bck/",
	}
	log.FatalIfErr(repo.Clone())
	log.FatalIfErr(repo.Add(filePath))
	log.FatalIfErr(repo.Push())
}

func parseDomains(d string) []string {
	return strings.Split(d, ",")
}
