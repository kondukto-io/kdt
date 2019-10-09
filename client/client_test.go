package client

import (
	"net/url"
	"testing"
)

func TestNewRequestPath(t *testing.T) {
	expected := "http://localhost:8080/test/path"
	client := New()
	u, err := url.Parse("http://localhost:8080")
	if err != nil {
		t.Error(err)
	}
	client.BaseURL = u
	req, err := client.newRequest("GET", "/test/path", nil)
	if err != nil {
		t.Error(err)
	}

	got := req.URL.String()
	if got != expected {
		t.Errorf("wrong request url: expected: %s got: %s", expected, got)
	}
}
