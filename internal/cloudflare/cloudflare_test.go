// SPDX-License-Identifier: MIT

package cloudflare

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_EmptyToken(t *testing.T) {
	t.Parallel()
	_, err := New("")
	assert.ErrorIs(t, err, ErrNoToken)
}

func TestNew_ValidToken(t *testing.T) {
	t.Parallel()
	c, err := New("valid-token")
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, DefaultBaseURL, c.api)
	assert.Equal(t, "valid-token", c.token)
}

func TestZones_SinglePage(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		resp := zoneListResponse{
			Success: true,
			Result: []zone{
				{ID: "zone1", Name: "example.com"},
				{ID: "zone2", Name: "test.org"},
			},
			Info: resultInfo{Page: 1, TotalPages: 1},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &Client{api: srv.URL, token: "test-token", http: http.DefaultClient}
	zones, err := c.Zones(t.Context())
	require.NoError(t, err)
	assert.Len(t, zones, 2)
	assert.Equal(t, "example.com", zones["zone1"])
	assert.Equal(t, "test.org", zones["zone2"])
}

func TestZones_MultiPage(t *testing.T) {
	t.Parallel()
	pageCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		pageCalls++
		page := r.URL.Query().Get("page")

		resp := zoneListResponse{Success: true, Info: resultInfo{TotalPages: 2}}
		if page == "1" {
			resp.Result = []zone{{ID: "z1", Name: "alpha.com"}}
			resp.Info.Page = 1
		}
		if page == "2" {
			resp.Result = []zone{{ID: "z2", Name: "beta.com"}}
			resp.Info.Page = 2
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &Client{api: srv.URL, token: "t", http: http.DefaultClient}
	zones, err := c.Zones(t.Context())
	require.NoError(t, err)
	assert.Len(t, zones, 2)
	assert.Equal(t, 2, pageCalls)
}

func TestZones_APError(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
		resp := zoneListResponse{
			Success: false,
			Errors:  []apiError{{Code: 1000, Message: "access denied"}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &Client{api: srv.URL, token: "t", http: http.DefaultClient}
	_, err := c.Zones(t.Context())
	assert.ErrorContains(t, err, "access denied")
}

func TestExport(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
		resp := zoneListResponse{
			Success: true,
			Result:  []zone{{ID: "zone-abc", Name: "example.com"}},
			Info:    resultInfo{Page: 1, TotalPages: 1},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/zones/zone-abc/dns_records/export", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("example.com. 300 IN A 1.2.3.4\n"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &Client{api: srv.URL, token: "test-token", http: http.DefaultClient}
	data, err := c.Export(t.Context(), "example.com")
	require.NoError(t, err)
	assert.Contains(t, string(data), "1.2.3.4")
}

func TestExport_DomainNotFound(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
		resp := zoneListResponse{Success: true, Result: []zone{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &Client{api: srv.URL, token: "t", http: http.DefaultClient}
	_, err := c.Export(t.Context(), "nonexistent.com")
	assert.ErrorIs(t, err, ErrDomainNF)
}

func TestZones_HTTP403(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"success":false,"errors":[{"code":9109,"message":"Unknown API Token"}]}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &Client{api: srv.URL, token: "bad-token", http: http.DefaultClient}
	_, err := c.Zones(t.Context())
	require.Error(t, err)
	require.ErrorContains(t, err, "token rejected")
	require.ErrorContains(t, err, "Zone.DNS")
}

func TestExport_HTTP403(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
		resp := zoneListResponse{
			Success: true,
			Result:  []zone{{ID: "zone-abc", Name: "example.com"}},
			Info:    resultInfo{Page: 1, TotalPages: 1},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/zones/zone-abc/dns_records/export", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`forbidden`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := &Client{api: srv.URL, token: "bad-token", http: http.DefaultClient}
	_, err := c.Export(t.Context(), "example.com")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
