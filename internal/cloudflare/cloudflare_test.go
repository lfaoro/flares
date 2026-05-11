// SPDX-License-Identifier: MIT

package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestZones(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mux     func(t *testing.T) *http.ServeMux
		wantLen int
		wantErr string
	}{
		{
			name: "single page",
			mux: func(t *testing.T) *http.ServeMux {
				mux := http.NewServeMux()
				mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
					resp := zoneListResponse{
						Success: true,
						Result: []zone{
							{ID: "z1", Name: "alpha.com"},
							{ID: "z2", Name: "beta.com"},
						},
						Info: resultInfo{Page: 1, TotalPages: 1},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(resp)
				})
				return mux
			},
			wantLen: 2,
		},
		{
			name: "multi page",
			mux: func(t *testing.T) *http.ServeMux {
				mux := http.NewServeMux()
				mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
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
				return mux
			},
			wantLen: 2,
		},
		{
			name: "API error",
			mux: func(t *testing.T) *http.ServeMux {
				mux := http.NewServeMux()
				mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
					resp := zoneListResponse{
						Success: false,
						Errors:  []apiError{{Code: 1000, Message: "access denied"}},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(resp)
				})
				return mux
			},
			wantErr: "access denied",
		},
		{
			name: "403 forbidden",
			mux: func(t *testing.T) *http.ServeMux {
				mux := http.NewServeMux()
				mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(`{"success":false,"errors":[{"code":9109,"message":"Unknown API Token"}]}`))
				})
				return mux
			},
			wantErr: "token rejected",
		},
		{
			name: "401 unauthorized",
			mux: func(t *testing.T) *http.ServeMux {
				mux := http.NewServeMux()
				mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`unauthorized`))
				})
				return mux
			},
			wantErr: "token rejected",
		},
		{
			name: "429 rate limited",
			mux: func(t *testing.T) *http.ServeMux {
				mux := http.NewServeMux()
				mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusTooManyRequests)
					w.Write([]byte(`rate limit hit`))
				})
				return mux
			},
			wantErr: "rate limited",
		},
		{
			name: "500 server error",
			mux: func(t *testing.T) *http.ServeMux {
				mux := http.NewServeMux()
				mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`internal error`))
				})
				return mux
			},
			wantErr: "unexpected status 500",
		},
		{
			name: "bad JSON response",
			mux: func(t *testing.T) *http.ServeMux {
				mux := http.NewServeMux()
				mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`not valid json`))
				})
				return mux
			},
			wantErr: "decode",
		},
		{
			name: "network error",
			mux: func(t *testing.T) *http.ServeMux {
				return http.NewServeMux()
			},
			wantErr: "do:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(tt.mux(t))
			if tt.wantErr == "do:" {
				srv.Close()
			}
			defer srv.Close()

			c := &Client{api: srv.URL, token: "test-token", http: http.DefaultClient}
			zones, err := c.Zones(t.Context())

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, zones, tt.wantLen)
		})
	}
}

func TestExport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mux      func(t *testing.T) *http.ServeMux
		domain   string
		wantData string
		wantErr  string
		wantIs   error
	}{
		{
			name: "success",
			mux: func(t *testing.T) *http.ServeMux {
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
				return mux
			},
			domain:   "example.com",
			wantData: "1.2.3.4",
		},
		{
			name: "domain not found",
			mux: func(t *testing.T) *http.ServeMux {
				mux := http.NewServeMux()
				mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
					resp := zoneListResponse{Success: true, Result: []zone{}}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(resp)
				})
				return mux
			},
			domain:  "nonexistent.com",
			wantIs:  ErrDomainNF,
			wantErr: "domain not found",
		},
		{
			name: "zone lookup 403",
			mux: func(t *testing.T) *http.ServeMux {
				mux := http.NewServeMux()
				mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(`forbidden`))
				})
				return mux
			},
			domain:  "example.com",
			wantErr: "token rejected",
		},
		{
			name: "export endpoint 403",
			mux: func(t *testing.T) *http.ServeMux {
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
				return mux
			},
			domain:  "example.com",
			wantErr: "403",
		},
		{
			name: "export endpoint 500",
			mux: func(t *testing.T) *http.ServeMux {
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
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`internal error`))
				})
				return mux
			},
			domain:  "example.com",
			wantErr: "unexpected status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(tt.mux(t))
			defer srv.Close()

			c := &Client{api: srv.URL, token: "test-token", http: http.DefaultClient}
			data, err := c.Export(t.Context(), tt.domain)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			if tt.wantIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantIs)
				return
			}
			require.NoError(t, err)
			assert.Contains(t, string(data), tt.wantData)
		})
	}
}

func ExampleClient_Export() {
	c, err := New("test-token")
	if err != nil {
		return
	}
	c.SetBaseURL("https://api.cloudflare.com/client/v4")
	_, err = c.Export(context.Background(), "example.com")
	if err != nil {
		fmt.Println("export failed:", err)
	}
}
