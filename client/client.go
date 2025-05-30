/*
Copyright © 2019 Kondukto

*/

package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
)

const (
	userAgent = "KDT"
)

var (
	ProductNotFound = errors.New("product not found")
	ProjectNotFound = errors.New("project not found")
)

type Client struct {
	httpClient *http.Client

	BaseURL *url.URL
}

type KonduktoError struct {
	Error string `json:"error"`
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
		return nil, fmt.Errorf("failed to do request: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, fmt.Errorf("failed to parse response: %v: %s", err, string(data))
		}

		return resp, nil
	}

	var e KonduktoError
	if err = json.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("failed to parse error message: %v: %s", err, string(data))
	}

	if e.Error != "" {
		return nil, fmt.Errorf("response not OK: response status:%d error message: %s", resp.StatusCode, e.Error)
	}

	return nil, fmt.Errorf("response not OK: %s", string(data))
}
