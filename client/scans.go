/*
Copyright © 2019 Kondukto

*/
package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/spf13/viper"
)

type (
	Scan struct {
		ID       string     `json:"id"`
		Name     string     `json:"name"`
		Branch   string     `json:"branch"`
		MetaData string     `json:"meta_data"`
		Tool     string     `json:"tool"`
		Date     *time.Time `json:"date"`
		Score    int        `json:"score"`
		Summary  Summary    `json:"summary"`
	}

	ScanSearchParams struct {
		Branch string `url:"branch,omitempty"`
		Tool   string `url:"tool,omitempty"`
		Meta   string `url:"meta,omitempty"`
		Limit  int    `url:"limit,omitempty"`
	}
	ScanPROptions struct {
		From string `json:"from"`
		To   string `json:"to"`
	}

	ResultSet struct {
		Score   int      `json:"score"`
		Summary *Summary `json:"summary"`
	}

	Summary struct {
		Critical int `json:"critical"`
		High     int `json:"high"`
		Medium   int `json:"medium"`
		Low      int `json:"low"`
		Info     int `json:"info"`
	}

	Event struct {
		ID      string `json:"id"`
		Status  int    `json:"status"`
		Active  int    `json:"active"`
		ScanId  string `json:"scan_id"`
		Message string `json:"message"`
	}
)

func (c *Client) ListScans(project string, params *ScanSearchParams) ([]Scan, error) {
	// TODO: list scans call should be updated to take tool and metadata arguments
	scans := make([]Scan, 0)

	path := fmt.Sprintf("/api/v1/projects/%s/scans", project)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return scans, err
	}

	v, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = v.Encode()

	type getProjectScansResponse struct {
		Scans []Scan `json:"data"`
		Total int    `json:"total"`
	}
	var ps getProjectScansResponse

	resp, err := c.do(req, &ps)
	if err != nil {
		return scans, err
	}

	if resp.StatusCode != http.StatusOK {
		return scans, errors.New("response not ok")
	}

	return ps.Scans, nil
}

func (c *Client) FindScan(project string, params *ScanSearchParams) (*Scan, error) {
	if params == nil {
		return nil, errors.New("scan query params cannot be empty")
	}
	params.Limit = 1
	scans, err := c.ListScans(project, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get scans :%w", err)
	}

	if len(scans) == 0 {
		return nil, errors.New("scan not found")
	}

	return &scans[0], nil
}

func (c *Client) StartScanByScanId(id string) (string, error) {
	path := fmt.Sprintf("/api/v1/scans/%s/restart", id)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}

	type restartScanResponse struct {
		Event   string `json:"event"`
		Message string `json:"message"`
	}
	var rsr restartScanResponse
	resp, err := c.do(req, &rsr)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		return "", errors.New("response not ok")
	}

	if rsr.Event == "" {
		return "", errors.New("event not found")
	}

	return rsr.Event, nil
}

func (c *Client) StartScanByOption(id string, opt *ScanPROptions) (string, error) {
	if opt == nil {
		return "", errors.New("missing scan options")
	}

	path := fmt.Sprintf("/api/v1/scans/%s/restart_with_option", id)
	req, err := c.newRequest(http.MethodPost, path, opt)
	if err != nil {
		return "", err
	}

	type restartScanResponse struct {
		Event   string `json:"event"`
		Message string `json:"message"`
	}
	var rsr restartScanResponse
	resp, err := c.do(req, &rsr)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		return "", errors.New("response not ok")
	}

	if rsr.Event == "" {
		return "", errors.New("event not found")
	}

	return rsr.Event, nil
}

func (c *Client) GetScanStatus(eventId string) (*Event, error) {
	path := fmt.Sprintf("/api/v1/events/%s/status", eventId)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	e := &Event{}
	resp, err := c.do(req, &e)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("response not ok")
	}

	return e, nil
}

func (c *Client) GetScanSummary(id string) (*Scan, error) {
	path := fmt.Sprintf("/api/v1/scans/%s", id)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	scan := &Scan{}
	resp, err := c.do(req, &scan)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("response not ok")
	}

	return scan, nil
}

func (c *Client) GetLastResults(id string) (map[string]*ResultSet, error) {
	path := fmt.Sprintf("/api/v1/scans/%s/last_results", id)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*ResultSet)
	resp, err := c.do(req, &m)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("response not ok")
	}

	return m, err
}

func (c *Client) ImportScanResult(project, branch, tool string, file string) (string, error) {
	path := "/api/v1/scans/import"
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return "", err
	}
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	part, err := writer.CreateFormFile("file", filepath.Base(f.Name()))
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, f)
	if err != nil {
		return "", err
	}

	if err = writer.WriteField("project", project); err != nil {
		return "", err
	}
	if err = writer.WriteField("branch", branch); err != nil {
		return "", err
	}
	if err = writer.WriteField("tool", tool); err != nil {
		return "", err
	}
	_ = writer.Close()

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Cookie", viper.GetString("token"))

	type importScanResultResponse struct {
		EventID string `json:"event_id"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	var importResponse importScanResultResponse
	resp, err := c.do(req, &importResponse)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to import scan results: %v", importResponse.Error)
	}

	return importResponse.EventID, nil
}

func (c *Client) ScanByImage(project, branch, tool, image string) (string, error) {
	path := "/api/v1/scans/image"

	type imageScanBody struct {
		Project string
		Tool    string
		Branch  string
		Image   string
	}
	reqBody := imageScanBody{
		Project: project,
		Tool:    tool,
		Branch:  branch,
		Image:   image,
	}

	req, err := c.newRequest(http.MethodPost, path, &reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	type responseBody struct {
		EventID string `json:"event_id"`
		Error   string `json:"error"`
	}
	respBody := new(responseBody)

	resp, err := c.do(req, respBody)
	if err != nil {
		return "", fmt.Errorf("HTTP response failed: %w", err)

	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("HTTP response not OK: %v", respBody.Error)
	}

	return respBody.EventID, nil
}
