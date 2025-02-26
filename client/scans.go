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

	"github.com/kondukto-io/kdt/klog"

	"github.com/google/go-querystring/query"
	"github.com/spf13/viper"
)

type (
	ImageScanParams struct {
		Project     string `json:"project"`
		Tool        string `json:"tool"`
		Branch      string `json:"branch"`
		Image       string `json:"image"`
		MetaData    string `json:"meta_data"`
		Environment string `json:"environment"`
	}

	ScanDetail struct {
		ID                   string     `json:"id"`
		Name                 string     `json:"name"`
		Branch               string     `json:"branch"`
		ScanType             string     `json:"scan_type"`
		MetaData             string     `json:"meta_data"`
		Tool                 string     `json:"tool"`
		ScannerType          string     `json:"scanner_type"`
		Date                 *time.Time `json:"date"`
		Project              string     `json:"project"`
		Score                int        `json:"score"`
		Summary              Summary    `json:"summary"`
		InfraSourceProjectID string     `json:"infra_source_project_id"`
		Links                struct {
			HTML string `json:"html"`
		} `json:"links"`
	}

	ScanSearchParams struct {
		Branch           string `url:"branch,omitempty"`
		Tool             string `url:"tool,omitempty"`
		MetaData         string `url:"meta_data"`
		PR               bool   `url:"pr"`
		Manual           bool   `url:"manual"`
		AgentID          string `url:"agent_id"`
		Environment      string `url:"environment"`
		ForkScan         bool   `url:"fork_scan"`
		ForkSourceBranch string `url:"fork_source_branch"`
		Limit            int    `url:"limit,omitempty"`
	}

	ScanRestartOptions struct {
		// MergeSourceBranch is source branch of the PR. It is required when PR is true
		MergeSourceBranch string `json:"from"`
		// MergeTargetBranch is target branch of the PR. It is required when PR is true
		MergeTargetBranch        string `json:"to"`
		OverrideOldAnalyze       bool   `json:"override_old_analyze"`
		PRNumber                 string `json:"pr_number"`
		NoDecoration             bool   `json:"no_decoration"`
		PRDecorationScannerTypes string `json:"pr_decoration_scanner_types"`
		Custom                   Custom `json:"custom"`
		Environment              string `url:"environment"`
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
		// MetaData is holding value of scanparam meta-data
		MetaData string `json:"meta_data"`
		// ForkScan is holding value of baseline scan
		ForkScan bool `json:"fork_scan"`
		// ForkSourceBranch is holding value of baseline scan branch
		ForkSourceBranch string `json:"fork_source_branch"`
		// OverrideForkSourceBranch is holding value of baseline scan branch
		OverrideForkSourceBranch bool `json:"override_fork_source_branch"`
		// Environment is holding value of application environment
		Environment string `json:"environment"`
	}

	PRInfo struct {
		// OK means that the merge target is a valid branch to merge the source branch changes into.
		OK                       bool   `json:"ok" json:"ok"`
		MergeTarget              string `json:"target" bson:"target" valid:"Branch"`
		PRNumber                 string `json:"pr_number"`
		NoDecoration             bool   `json:"no_decoration"`
		PRDecorationScannerTypes string `json:"pr_decoration_scanner_types"`
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
		return "", createHTTPRequestError(err)
	}

	type scanResponse struct {
		EventID string `json:"event_id"`
		Message string `json:"message"`
	}
	var rsr scanResponse
	resp, err := c.do(req, &rsr)
	if err != nil {
		return "", fmt.Errorf("failed to create new scan: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", httpStatusError(http.StatusCreated, resp.StatusCode)
	}

	if rsr.EventID == "" {
		return "", errors.New("event id not found in new scan response")
	}

	return rsr.EventID, nil
}

func (c *Client) RestartScanByScanID(id string) (string, error) {
	if id == "" {
		return "", scanIDEmptyError()
	}

	klog.Debugf("restarting scan with id: %s", id)

	path := fmt.Sprintf("/api/v2/scans/%s/restart", id)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", createHTTPRequestError(err)
	}

	type restartScanResponse struct {
		Event   string `json:"event"`
		Message string `json:"message"`
	}
	var rsr restartScanResponse
	resp, err := c.do(req, &rsr)
	if err != nil {
		return "", fmt.Errorf("failed to restart scan: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", httpStatusError(http.StatusCreated, resp.StatusCode)
	}

	if rsr.Event == "" {
		return "", errors.New("event not found in restart scan response")
	}

	return rsr.Event, nil
}

func (c *Client) RestartScanWithOption(id string, opt *ScanRestartOptions) (string, error) {
	if id == "" {
		return "", scanIDEmptyError()
	}

	if opt == nil {
		return "", errors.New("missing scan options")
	}

	klog.Debugf("restarting scan with id and options: %s", id)

	path := fmt.Sprintf("/api/v2/scans/%s/restart_with_option", id)
	req, err := c.newRequest(http.MethodPost, path, opt)
	if err != nil {
		return "", createHTTPRequestError(err)
	}

	type restartScanResponse struct {
		Event   string `json:"event"`
		Message string `json:"message"`
	}
	var rsr restartScanResponse
	resp, err := c.do(req, &rsr)
	if err != nil {
		return "", fmt.Errorf("failed to restart scan: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", httpStatusError(http.StatusCreated, resp.StatusCode)
	}

	if rsr.Event == "" {
		return "", errors.New("event not found")
	}

	return rsr.Event, nil
}

type ScanByImageInput struct {
	Project     string
	Tool        string
	Branch      string
	Image       string
	MetaData    string
	Environment string
}

func (i *ScanByImageInput) prepareRequestQueryParameters() ImageScanParams {
	return ImageScanParams{
		Project:     i.Project,
		Tool:        i.Tool,
		Branch:      i.Branch,
		Image:       i.Image,
		MetaData:    i.MetaData,
		Environment: i.Environment,
	}
}

func (c *Client) ScanByImage(pr *ScanByImageInput) (string, error) {
	if pr == nil {
		return "", errors.New("missing scan fields")
	}

	klog.Debug("scanning image with given parameters")

	path := "/api/v2/scans/image"
	req, err := c.newRequest(http.MethodPost, path, pr.prepareRequestQueryParameters())
	if err != nil {
		return "", createHTTPRequestError(err)
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
		return "", httpStatusError(http.StatusCreated, resp.StatusCode)
	}

	return respBody.EventID, nil
}

type ImportForm map[string]string

func (c *Client) ImportScanResult(file string, form ImportForm) (string, error) {
	if file == "" {
		return "", errors.New("file path cannot be empty")
	}

	klog.Debugf("importing scan results from file: %s", file)

	path := "/api/v2/scans/import"
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", file)
	}

	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = f.Close() }()

	part, err := writer.CreateFormFile("file", filepath.Base(f.Name()))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, f)
	if err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	for k := range form {
		if err = writer.WriteField(k, form[k]); err != nil {
			return "", fmt.Errorf("failed to write form field [%s]: %w", k, err)
		}
	}

	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return "", createHTTPRequestError(err)
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
		return "", fmt.Errorf("failed to import scan results: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", httpStatusError(http.StatusOK, resp.StatusCode)
	}

	if importResponse.EventID == "" {
		return "", errors.New("event id not found in import response")
	}

	return importResponse.EventID, nil
}

func (c *Client) ListScans(project string, params *ScanSearchParams) ([]ScanDetail, error) {
	if project == "" {
		return nil, errors.New("project name cannot be empty")
	}

	klog.Debugf("retrieving scans of the project: %s", project)

	scans := make([]ScanDetail, 0)
	path := fmt.Sprintf("/api/v2/projects/%s/scans", project)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return scans, createHTTPRequestError(err)
	}

	v, err := query.Values(params)
	if err != nil {
		return nil, fmt.Errorf("failed to encode query params: %w", err)
	}
	req.URL.RawQuery = v.Encode()

	type getProjectScansResponse struct {
		Scans []ScanDetail `json:"scans"`
		Total int          `json:"total"`
	}
	var ps getProjectScansResponse

	resp, err := c.do(req, &ps)
	if err != nil {
		return scans, fmt.Errorf("failed to get scans: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpStatusError(http.StatusOK, resp.StatusCode)
	}

	return ps.Scans, nil
}

func (c *Client) FindScan(project string, params *ScanSearchParams) (*ScanDetail, error) {
	if params == nil {
		return nil, errors.New("scan query params cannot be empty")
	}

	klog.Debugf("retrieving scan of the project: %s", project)

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
	if id == "" {
		return nil, scanIDEmptyError()
	}

	klog.Debugf("retrieving scan by id: %s", id)

	path := fmt.Sprintf("/api/v2/scans/%s", id)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, createHTTPRequestError(err)
	}

	var scan ScanDetail
	resp, err := c.do(req, &scan)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpStatusError(http.StatusOK, resp.StatusCode)
	}

	return &scan, nil
}

func (c *Client) GetScanStatus(eventId string) (*Event, error) {
	if eventId == "" {
		return nil, eventIDEmptyError()
	}

	klog.Debugf("retrieving scan status of the event: %s", eventId)

	path := fmt.Sprintf("/api/v2/events/%s/status", eventId)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, createHTTPRequestError(err)
	}

	var e Event
	resp, err := c.do(req, &e)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan status: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpStatusError(http.StatusOK, resp.StatusCode)
	}

	return &e, nil
}

func (c *Client) GetLastResults(id string) (map[string]*ResultSet, error) {
	if id == "" {
		return nil, scanIDEmptyError()
	}

	klog.Debugf("retrieving last results of the scan: %s", id)

	path := fmt.Sprintf("/api/v2/scans/%s/last_results", id)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, createHTTPRequestError(err)
	}

	m := make(map[string]*ResultSet)
	resp, err := c.do(req, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to get last results: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpStatusError(http.StatusOK, resp.StatusCode)
	}

	return m, err
}

func createHTTPRequestError(err error) error {
	return fmt.Errorf("failed to create HTTP request: %w", err)
}

func httpStatusError(expected, status int) error {
	return fmt.Errorf("HTTP response status not expected: %d/%d", expected, status)
}

func scanIDEmptyError() error {
	return errors.New("scan id cannot be empty")
}

func eventIDEmptyError() error {
	return errors.New("event id cannot be empty")
}
