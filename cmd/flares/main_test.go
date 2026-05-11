// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureOutput(fn func()) (string, string) {
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = stdoutW
	os.Stderr = stderrW

	fn()

	os.Stdout = oldStdout
	os.Stderr = oldStderr
	stdoutW.Close()
	stderrW.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, stdoutR)
	_, _ = io.Copy(&stderrBuf, stderrR)

	return stdoutBuf.String(), stderrBuf.String()
}

func runApp(args ...string) error {
	cmd := newCmd()
	return cmd.Run(context.Background(), args)
}

func TestCLI_Help(t *testing.T) {
	stdout, stderr := captureOutput(func() {
		err := runApp("flares", "--help")
		assert.NoError(t, err)
	})
	assert.Contains(t, stdout, "Cloudflare DNS backup tool")
	assert.Contains(t, stdout, "show")
	assert.Contains(t, stdout, "export")
	assert.Contains(t, stdout, "zones")
	assert.Empty(t, stderr)
}

func TestCLI_Version(t *testing.T) {
	stdout, stderr := captureOutput(func() {
		err := runApp("flares", "--version")
		assert.NoError(t, err)
	})
	assert.Contains(t, stdout, "flares version")
	assert.Empty(t, stderr)
}

func TestCLI_ShowHelp(t *testing.T) {
	stdout, stderr := captureOutput(func() {
		err := runApp("flares", "show", "--help")
		assert.NoError(t, err)
	})
	assert.Contains(t, stdout, "all")
	assert.Contains(t, stdout, "output")
	assert.Contains(t, stdout, "text")
	assert.Contains(t, stdout, "json")
	assert.Empty(t, stderr)
}

func TestCLI_ExportHelp(t *testing.T) {
	stdout, stderr := captureOutput(func() {
		err := runApp("flares", "export", "--help")
		assert.NoError(t, err)
	})
	assert.Contains(t, stdout, "all")
	assert.Empty(t, stderr)
}

func TestCLI_ZonesHelp(t *testing.T) {
	stdout, stderr := captureOutput(func() {
		err := runApp("flares", "zones", "--help")
		assert.NoError(t, err)
	})
	assert.Contains(t, stdout, "List all zones")
	assert.Empty(t, stderr)
}

func TestCLI_NoSubcommand(t *testing.T) {
	stdout, stderr := captureOutput(func() {
		err := runApp("flares")
		assert.NoError(t, err)
	})
	assert.Contains(t, stdout, "Cloudflare DNS backup tool")
	assert.Contains(t, stdout, "COMMANDS")
	assert.Empty(t, stderr)
}

func TestCLI_NoToken(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_TOKEN", "")

	err := runApp("flares", "show", "example.com")
	require.Error(t, err)
	require.ErrorContains(t, err, "provide --token flag")
	require.ErrorContains(t, err, "CLOUDFLARE_API_TOKEN")
	require.ErrorContains(t, err, "dash.cloudflare.com")
}

func TestCLI_NoDomains(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_TOKEN", "test-token")

	err := runApp("flares", "show")
	require.Error(t, err)
	require.ErrorContains(t, err, "at least one domain required")
}

func TestCLI_ShowJSONFlag_NoDomains(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_TOKEN", "test-token")

	err := runApp("flares", "show", "--output", "json")
	require.Error(t, err)
	require.ErrorContains(t, err, "at least one domain required")
}

func TestCLI_UnknownCommand(t *testing.T) {
	stdout, _ := captureOutput(func() {
		err := runApp("flares", "unknown")
		assert.NoError(t, err)
	})
	assert.Contains(t, stdout, "Cloudflare DNS backup tool")
}

func TestCLI_ShowDomain(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") == "example.com" {
			resp := map[string]any{"success": true, "result": []map[string]any{{"id": "z1", "name": "example.com"}}, "result_info": map[string]int{"page": 1, "total_pages": 1}}
			json.NewEncoder(w).Encode(resp)
		}
	})
	mux.HandleFunc("/zones/z1/dns_records/export", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("example.com. 300 IN A 1.2.3.4\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	stdout, _ := captureOutput(func() {
		err := runApp("flares", "--token", "test", "--api-url", srv.URL, "show", "example.com")
		assert.NoError(t, err)
	})
	assert.Contains(t, stdout, "1.2.3.4")
}

func TestCLI_ShowDomainJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") == "example.com" {
			resp := map[string]any{"success": true, "result": []map[string]any{{"id": "z1", "name": "example.com"}}, "result_info": map[string]int{"page": 1, "total_pages": 1}}
			json.NewEncoder(w).Encode(resp)
		}
	})
	mux.HandleFunc("/zones/z1/dns_records/export", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("example.com. 300 IN A 1.2.3.4\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	stdout, _ := captureOutput(func() {
		err := runApp("flares", "--token", "test", "--api-url", srv.URL, "show", "--output", "json", "example.com")
		assert.NoError(t, err)
	})
	assert.Contains(t, stdout, "example.com. 300 IN A 1.2.3.4")
}

func TestCLI_ExportDomain(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") == "example.com" {
			resp := map[string]any{"success": true, "result": []map[string]any{{"id": "z1", "name": "example.com"}}, "result_info": map[string]int{"page": 1, "total_pages": 1}}
			json.NewEncoder(w).Encode(resp)
		}
	})
	mux.HandleFunc("/zones/z1/dns_records/export", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("example.com. 300 IN A 1.2.3.4\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	stdout, _ := captureOutput(func() {
		err := runApp("flares", "--token", "test", "--api-url", srv.URL, "export", "example.com")
		assert.NoError(t, err)
	})
	assert.Contains(t, stdout, "BIND data for example.com successfully exported")

	data, err := os.ReadFile(filepath.Join(tmpDir, "example.com"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "1.2.3.4")
}

func TestCLI_Zones(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/zones", func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"success": true,
			"result": []map[string]any{
				{"id": "z1", "name": "example.com"},
				{"id": "z2", "name": "test.org"},
			},
			"result_info": map[string]int{"page": 1, "total_pages": 1},
		}
		json.NewEncoder(w).Encode(resp)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	stdout, _ := captureOutput(func() {
		err := runApp("flares", "--token", "test", "--api-url", srv.URL, "zones")
		assert.NoError(t, err)
	})
	assert.Contains(t, stdout, "z1")
	assert.Contains(t, stdout, "example.com")
	assert.Contains(t, stdout, "z2")
	assert.Contains(t, stdout, "test.org")
}
