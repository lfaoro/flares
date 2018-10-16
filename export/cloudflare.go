package export

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Cloudflare struct {
	API       string
	AuthKey   string
	AuthEmail string
	Client    http.Client
}

var _ Exporter = Cloudflare{}

// Export fetches the BIND DNS table for a domain and returns
// its contents.
func (cf Cloudflare) Export(domain string) ([]byte, error) {
	return cf.exportFor(domain)
}

func (cf Cloudflare) exportFor(domain string) ([]byte, error) {
	// fetch the zone for the domain
	zone, err := cf.zoneFor(domain)
	if err != nil {
		return nil, err
	}
	endpoint := cf.API + "/zones/" + zone + "/dns_records/export"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-auth-key", cf.AuthKey)
	req.Header.Add("x-auth-email", cf.AuthEmail)
	res, err := cf.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (cf Cloudflare) zoneFor(domain string) (string, error) {
	endpoint := cf.API + "/zones" + fmt.Sprintf("?%v", domain)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("x-auth-key", cf.AuthKey)
	req.Header.Add("x-auth-email", cf.AuthEmail)
	res, err := cf.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	response := response{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return "", err
	}
	return response.Result[0].ID, nil
}

type response struct {
	Result []struct {
		ID string `json:"id"`
	} `json:"result"`
}
