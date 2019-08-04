/*
 * Copyright (c) 2019 Leonardo Faoro. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package cloudflare

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	errDomainNotFound  = "cloudflare: domain not found"
	errNoAuthorization = "cloudflare: missing required AuthKey, AuthEmail"
)

// Cloudflare sets up authorization to the API.
type Cloudflare struct {
	API       string
	AuthKey   string
	AuthEmail string
	Client    http.Client
}

// New returns a Cloudflare client
func New(apiKey, apiEmail string) Cloudflare {
	if apiKey == "" || apiEmail == "" {
		panic(errNoAuthorization)
	}
	client := Cloudflare{
		API:       "https://api.cloudflare.com/client/v4",
		AuthKey:   apiKey,
		AuthEmail: apiEmail,
		Client: http.Client{
			Timeout: time.Second * 30,
		},
	}
	return client
}

// TableFor fetches the BIND DNS table for a domain.
func (cf Cloudflare) TableFor(domain string) ([]byte, error) {
	return cf.tableFor(domain)
}

// Zones returns a map(ID:domain) of all the zones available in your
// CloudFlare account.
//
// ref: https://api.cloudflare.com/#zone-list-zones
func (cf Cloudflare) Zones() (map[string]string, error) {
	var result = map[string]string{}

	var count = 1
	for {
		endpoint := cf.API + "/zones"
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, errors.Wrap(err, "cloudflare:")
		}

		v := url.Values{}
		maxPerPageValue := 50
		v.Add("per_page", strconv.Itoa(maxPerPageValue))
		v.Add("page", strconv.Itoa(count))
		u.RawQuery = v.Encode()

		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, errors.Wrap(err, "cloudflare:")
		}

		cf.setAuthHeaders(req)

		res, err := cf.Client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "cloudflare:")
		}
		defer res.Body.Close()

		data := response{}
		if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
			return nil, errors.Wrap(err, "cloudflare:")
		}

		if !data.Success {
			return nil, errors.New(data.Errors[0].Message)
		}

		for _, res := range data.Result {
			result[res.ID] = res.Name
		}

		pages := data.ResultInfo.TotalCount / maxPerPageValue
		if count < pages {
			count++
			continue
		}
		break
	}

	return result, nil
}

func (cf Cloudflare) tableFor(domain string) ([]byte, error) {
	zone, err := cf.zoneIDFor(domain)
	if err != nil {
		return nil, err
	}

	endpoint := cf.API + "/zones/" + zone + "/dns_records/export"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	cf.setAuthHeaders(req)

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

func (cf Cloudflare) zoneIDFor(domain string) (string, error) {
	endpoint := cf.API + "/zones"
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	v := url.Values{}
	v.Add("name", domain)
	u.RawQuery = v.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}

	cf.setAuthHeaders(req)

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
		return "", errors.New(response.Errors[0].Message)
	}
	if len(response.Result) == 0 {
		return "", errors.New(errDomainNotFound)
	}
	return response.Result[0].ID, nil
}

func (cf Cloudflare) setAuthHeaders(req *http.Request) {
	req.Header.Add("X-Auth-Key", cf.AuthKey)
	req.Header.Add("X-Auth-Email", cf.AuthEmail)
}

type response struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
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
