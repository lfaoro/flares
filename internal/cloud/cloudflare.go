package cloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Cloudflare sets up authorization to the API.
type Cloudflare struct {
	API       string
	AuthKey   string
	AuthEmail string
	Client    http.Client
}

func NewCloudflare(apiKey, apiEmail string) Cloudflare {
	if apiKey == "" || apiEmail == "" {
		apiKey = os.Getenv("CF_API_KEY")
		apiEmail = os.Getenv("CF_API_EMAIL")
	}
	client := Cloudflare{
		API:       "https://api.cloudflare.com/client/v4",
		AuthKey:   apiKey,
		AuthEmail: apiEmail,
		Client: http.Client{
			Timeout: time.Second * 10,
		},
	}
	return client
}

const errDomainNotFound = "cloudflare: domain not found"

// guarantee interface compliance on build.
var _ Service = Cloudflare{}

// DNSTableFor fetches the BIND DNS table for a domain.
func (cf Cloudflare) DNSTableFor(domain string) ([]byte, error) {
	if cf.AuthKey == "" || cf.AuthEmail == "" {
		return nil, errors.New("missing required AuthKey || AuthEmail")
	}
	return cf.exportFor(domain)
}

func (cf Cloudflare) AllZones() ([]string, error) {
	endpoint := cf.API + "/zones" //+ "/?per_page=50"
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	response := response{}

	var result []string

	page := 1
	for {

		req.Header.Add("x-auth-key", cf.AuthKey)
		req.Header.Add("x-auth-email", cf.AuthEmail)

		q := req.URL.Query()
		q.Set("per_page", "50")
		req.URL.RawQuery = q.Encode()

		res, err := cf.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			return nil, err
		}
		if !response.Success {
			fmt.Println("DEBUG:", response.Errors)
			return nil, errors.New(errDomainNotFound)
		}
		if len(response.Result) == 0 {
			return nil, errors.New(errDomainNotFound)
		}

		// populate result
		for _, res := range response.Result {
			result = append(result, strings.Join([]string{res.ID, res.Name}, ","))
		}

		totalPages := response.ResultInfo.TotalCount / response.ResultInfo.PerPage
		if page < totalPages {
			page++
			q := req.URL.Query()
			q.Add("page", string(page))
			req.URL.RawQuery = q.Encode()
			continue
		}
		break
	}

	return result, nil
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
	endpoint := cf.API + "/zones" + fmt.Sprintf("?name=%v", domain)
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("GET", parsed.String(), nil)
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
	if !response.Success {
		fmt.Println("DEBUG:", response.Errors)
		return "", errors.New(errDomainNotFound)
	}
	if len(response.Result) == 0 {
		return "", errors.New(errDomainNotFound)
	}
	return response.Result[0].ID, nil
}

type response struct {
	Success  bool          `json:"success"`
	Errors   []interface{} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   []struct {
		ID                  string    `json:"id"`
		Name                string    `json:"name"`
		DevelopmentMode     int       `json:"development_mode"`
		OriginalNameServers []string  `json:"original_name_servers"`
		OriginalRegistrar   string    `json:"original_registrar"`
		OriginalDnshost     string    `json:"original_dnshost"`
		CreatedOn           time.Time `json:"created_on"`
		ModifiedOn          time.Time `json:"modified_on"`
		ActivatedOn         time.Time `json:"activated_on"`
		Owner               struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Type  string `json:"type"`
		} `json:"owner"`
		Account struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"account"`
		Permissions []string `json:"permissions"`
		Plan        struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Price        int    `json:"price"`
			Currency     string `json:"currency"`
			Frequency    string `json:"frequency"`
			LegacyID     string `json:"legacy_id"`
			IsSubscribed bool   `json:"is_subscribed"`
			CanSubscribe bool   `json:"can_subscribe"`
		} `json:"plan"`
		PlanPending struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Price        int    `json:"price"`
			Currency     string `json:"currency"`
			Frequency    string `json:"frequency"`
			LegacyID     string `json:"legacy_id"`
			IsSubscribed bool   `json:"is_subscribed"`
			CanSubscribe bool   `json:"can_subscribe"`
		} `json:"plan_pending"`
		Status      string   `json:"status"`
		Paused      bool     `json:"paused"`
		Type        string   `json:"type"`
		NameServers []string `json:"name_servers"`
	} `json:"result"`
	ResultInfo struct {
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
}
