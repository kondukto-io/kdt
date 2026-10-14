/*
Copyright © 2019 Invicti Security
https://www.invicti.com/
*/

package client

import (
	"crypto/fips140"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

// The tests below only make sense when built against the FIPS 140-3 module.
// Run them with `make test-fips`.
func requireFIPSBuild(t *testing.T) {
	t.Helper()
	if os.Getenv("GOFIPS140") == "" {
		t.Skip("GOFIPS140 is not set, run `make test-fips`")
	}
}

func TestFIPSModeEnabled(t *testing.T) {
	requireFIPSBuild(t)

	if !fips140.Enabled() {
		t.Fatal("FIPS 140-3 mode is not enabled")
	}
}

func TestFIPSTLSRoundTrip(t *testing.T) {
	requireFIPSBuild(t)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	c := &Client{httpClient: srv.Client(), BaseURL: u}

	req, err := c.newRequest(http.MethodGet, "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	var body struct {
		Status string `json:"status"`
	}

	resp, err := c.do(req, &body)
	if err != nil {
		t.Fatalf("TLS request failed in FIPS mode: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("unexpected response body: %+v", body)
	}

	if resp.TLS == nil || resp.TLS.Version < tls.VersionTLS12 {
		t.Fatalf("expected TLS 1.2 or newer, got connection state: %+v", resp.TLS)
	}
}
