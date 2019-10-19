/*
Copyright © 2019 Kondukto

*/
package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
)

const (
	userAgent = "KDT"
)

type Client struct {
	httpClient *http.Client

	BaseURL *url.URL
}

func New() (*Client, error) {
	client := new(Client)

	httpClient := http.DefaultClient

	u, err := url.Parse(viper.GetString("host"))
	if err != nil {
		return client, err
	}
	client.BaseURL = u

	if viper.GetBool("insecure") {
		tp := http.DefaultTransport.(*http.Transport)
		tp.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		httpClient.Transport = tp
	}

	client.httpClient = httpClient

	return client, nil
}

func (c *Client) newRequest(method string, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Cookie", viper.GetString("token"))

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}
