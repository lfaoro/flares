// SPDX-License-Identifier: MIT

package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	apiBaseURL   = "https://api.cloudflare.com/client/v4"
	ErrNoToken   = simpleError("cloudflare: missing API token")
	ErrDomainNF  = simpleError("cloudflare: domain not found")
	perPage      = 50
	requestTimut = 30 * time.Second
)

type simpleError string

func (e simpleError) Error() string { return string(e) }

type Client struct {
	api   string
	token string
	http  *http.Client
}

func New(token string) (*Client, error) {
	if token == "" {
		return nil, ErrNoToken
	}
	return &Client{
		api:   apiBaseURL,
		token: token,
		http:  &http.Client{Timeout: requestTimut},
	}, nil
}

func (c *Client) SetBaseURL(url string) {
	c.api = url
}

func (c *Client) Zones(ctx context.Context) (map[string]string, error) {
	zones := map[string]string{}
	page := 1

	for {
		v := url.Values{}
		v.Set("per_page", strconv.Itoa(perPage))
		v.Set("page", strconv.Itoa(page))

		var res zoneListResponse
		if err := c.do(ctx, http.MethodGet, "/zones", v, &res); err != nil {
			return nil, fmt.Errorf("list zones: %w", err)
		}
		if !res.Success {
			return nil, fmt.Errorf("list zones: %s", res.Errors[0].Message)
		}
		for _, z := range res.Result {
			zones[z.ID] = z.Name
		}
		if page >= res.Info.TotalPages {
			break
		}
		page++
	}

	return zones, nil
}

func (c *Client) Export(ctx context.Context, domain string) ([]byte, error) {
	zoneID, err := c.zoneID(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("export %s: %w", domain, err)
	}

	endpoint := fmt.Sprintf("/zones/%s/dns_records/export", zoneID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.api+endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("export request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("export do: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("export: status %d: %s", res.StatusCode, string(body))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("export read: %w", err)
	}
	return body, nil
}

func (c *Client) zoneID(ctx context.Context, domain string) (string, error) {
	v := url.Values{}
	v.Set("name", domain)

	var res zoneListResponse
	if err := c.do(ctx, http.MethodGet, "/zones", v, &res); err != nil {
		return "", fmt.Errorf("zone id: %w", err)
	}
	if !res.Success {
		return "", fmt.Errorf("zone id: %s", res.Errors[0].Message)
	}
	if len(res.Result) == 0 {
		return "", ErrDomainNF
	}
	return res.Result[0].ID, nil
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, dest any) error {
	u, err := url.Parse(c.api + path)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return fmt.Errorf("token rejected: check your API token has Zone.DNS read permission")
	}
	if res.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("rate limited: %s", string(raw))
	}
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("unexpected status %d: %s", res.StatusCode, string(raw))
	}

	if err := json.Unmarshal(raw, dest); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type resultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

type zoneListResponse struct {
	Success bool       `json:"success"`
	Errors  []apiError `json:"errors"`
	Result  []zone     `json:"result"`
	Info    resultInfo `json:"result_info"`
}

type zone struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	DevelopmentMode     int      `json:"development_mode"`
	OriginalNameServers []string `json:"original_name_servers"`
	OriginalRegistrar   string   `json:"original_registrar"`
	OriginalDnshost     string   `json:"original_dnshost"`
	CreatedOn           string   `json:"created_on"`
	ModifiedOn          string   `json:"modified_on"`
	ActivatedOn         string   `json:"activated_on"`
	Owner               struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
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
	Meta        struct {
		CDNOnly                bool `json:"cdn_only"`
		CustomCertificateQuota int  `json:"custom_certificate_quota"`
		DNSOnly                bool `json:"dns_only"`
		FoundationDNS          bool `json:"foundation_dns"`
		PageRuleQuota          int  `json:"page_rule_quota"`
		PhishingDetected       bool `json:"phishing_detected"`
		Step                   int  `json:"step"`
	} `json:"meta"`
	CnameSuffix     string `json:"cname_suffix,omitempty"`
	VerificationKey string `json:"verification_key,omitempty"`
	Tenant          *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"tenant,omitempty"`
	TenantUnit *struct {
		ID string `json:"id"`
	} `json:"tenant_unit,omitempty"`
	VanityNameServers []string `json:"vanity_name_servers,omitempty"`
}
