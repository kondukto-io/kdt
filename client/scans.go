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

	"github.com/kondukto-io/kdt/klog"
)

type (
	ScanDetail struct {
		ID       string     `json:"id"`
		Name     string     `json:"name"`
		Branch   string     `json:"branch"`
		ScanType string     `json:"scan_type"`
		MetaData string     `json:"meta_data"`
		Tool     string     `json:"tool"`
		Date     *time.Time `json:"date"`
		Project  string     `json:"project"`
		Score    int        `json:"score"`
		Summary  Summary    `json:"summary"`
		Links    struct {
			HTML string `json:"html"`
		} `json:"links"`
	}

	ScanSearchParams struct {
		Branch   string `url:"branch,omitempty"`
		Tool     string `url:"tool,omitempty"`
		MetaData string `url:"meta_data,omitempty"`
		PR       bool   `url:"pr"`
		Manual   bool   `url:"manual"`
		AgentID  string `url:"agent_id"`
		ForkScan bool   `url:"fork_scan"`
		Limit    int    `url:"limit,omitempty"`
	}

	ScanPROptions struct {
		From               string `json:"from"`
		To                 string `json:"to"`
		OverrideOldAnalyze bool   `json:"override_old_analyze"`
		PRNumber           string `json:"pr_number"`
		NoDecoration       bool   `json:"no_decoration"`
		Custom             Custom `json:"custom"`
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
		ID         string `json:"id"`
		Status     int    `json:"status"`
		Active     int    `json:"active"`
		ScanID     string `json:"scan_id"`
		StatusText string `json:"status_text"`
		Message    string `json:"message"`
		Links      struct {
			HTML string `json:"html"`
		} `json:"links"`
	}

	Scan struct {
		// ScanparamsID is holding identifier of scanparams, when given, it will override other fields
		ScanparamsID string `json:"scanparams_id,omitempty"`
		// Branch is holding current branch value of scan
		Branch string `json:"branch"`
		// Project is holding ID or Name value of project
		Project string `json:"project"`
		// ToolID is holding ID value of selected scanner
		ToolID string `json:"tool_id,omitempty"`
		// AgentID is holding ID value of selected agent
		AgentID string `json:"agent_id,omitempty"`
		// PR is holding detail of pull requests branches to be scanned
		PR PRInfo `json:"pr"`
		// Custom is holding custom type of scanners that specified on the Kondukto side
		Custom Custom `json:"custom"`
		// ForkScan is holding value of baseline scan
		ForkScan bool `json:"fork_scan"`
		// MetaData is holding value of scanparam meta-data
		MetaData string `json:"meta_data"`
	}

	PRInfo struct {
		OK           bool   `json:"ok" json:"ok"`
		Target       string `json:"target" bson:"target" valid:"Branch"`
		PRNumber     string `json:"pr_number"`
		NoDecoration bool   `json:"no_decoration"`
	}

	Custom struct {
		Type   int                    `json:"type" bson:"type"`
		Params map[string]interface{} `json:"params" bson:"params"`
	}
)

func (c *Client) CreateNewScan(scan *Scan) (string, error) {
	klog.Debug("creating new scan with given parameters")
	if scan == nil {
		return "", errors.New("missing scan fields")
	}

	path := "/api/v2/scans/create"
	req, err := c.newRequest(http.MethodPost, path, scan)
	if err != nil {
		return "", err
	}

	type scanResponse struct {
		EventID string `json:"event_id"`
		Message string `json:"message"`
	}
	var rsr scanResponse
	_, err = c.do(req, &rsr)
	if err != nil {
		return "", err
	}

	return rsr.EventID, nil
}

func (c *Client) RestartScanByScanID(id string) (string, error) {
	klog.Debug("starting scan by scan_id")
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
	_, err = c.do(req, &rsr)
	if err != nil {
		return "", err
	}

	return rsr.Event, nil
}

func (c *Client) RestartScanWithOption(id string, opt *ScanPROptions) (string, error) {
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
		return "", fmt.Errorf("HTTP response not OK: %d", resp.StatusCode)
	}

	if rsr.Event == "" {
		return "", errors.New("event not found")
	}

	return rsr.Event, nil
}

type ImageScanParams struct {
	Project  string `json:"project"`
	Tool     string `json:"tool"`
	Branch   string `json:"branch"`
	Image    string `json:"image"`
	MetaData string `json:"meta_data"`
}

func (c *Client) ScanByImage(pr *ImageScanParams) (string, error) {
	path := "/api/v1/scans/image"

	req, err := c.newRequest(http.MethodPost, path, pr)
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
		return "", fmt.Errorf("HTTP response not OK: %s", respBody.Error)
	}

	return respBody.EventID, nil
}

type ImportForm map[string]string

func (c *Client) ImportScanResult(file string, form ImportForm) (string, error) {
	klog.Debugf("importing scan results using the file:%s", file)

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

	for k := range form {
		if err = writer.WriteField(k, form[k]); err != nil {
			return "", fmt.Errorf("failed to write form field [%s]: %w", k, err)
		}
	}

	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
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
		return "", fmt.Errorf("failed to import scan results: %s", importResponse.Error)
	}

	return importResponse.EventID, nil
}

func (c *Client) ListScans(project string, params *ScanSearchParams) ([]ScanDetail, error) {
	klog.Debugf("retrieving scans of the project: %s", project)

	scans := make([]ScanDetail, 0)
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
		Scans []ScanDetail `json:"data"`
		Total int          `json:"total"`
	}
	var ps getProjectScansResponse

	resp, err := c.do(req, &ps)
	if err != nil {
		return scans, err
	}

	if resp.StatusCode != http.StatusOK {
		return scans, fmt.Errorf("HTTP response not OK: %d", resp.StatusCode)
	}

	return ps.Scans, nil
}

func (c *Client) FindScan(project string, params *ScanSearchParams) (*ScanDetail, error) {
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

func (c *Client) FindScanByID(id string) (*ScanDetail, error) {
	path := fmt.Sprintf("/api/v1/scans/%s", id)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var scan ScanDetail
	resp, err := c.do(req, &scan)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP response not OK: %d", resp.StatusCode)
	}

	return &scan, nil
}

func (c *Client) GetScanStatus(eventId string) (*Event, error) {
	path := fmt.Sprintf("/api/v2/events/%s/status", eventId)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var e Event
	resp, err := c.do(req, &e)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP response not OK: %d", resp.StatusCode)
	}

	return &e, nil
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
		return nil, fmt.Errorf("HTTP response not OK: %d", resp.StatusCode)
	}

	return m, err
}
