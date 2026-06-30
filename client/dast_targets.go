/*
Copyright © 2019 Invicti Security
https://www.invicti.com/
*/

package client

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/kondukto-io/kdt/klog"

	"github.com/google/go-querystring/query"
)

type DastTarget struct {
	ID     string         `json:"id,omitempty"`
	Name   string         `json:"name,omitempty"`
	URL    string         `json:"url"`
	Labels []ProjectLabel `json:"labels"`
	Team   ProjectTeam    `json:"team"`
	Links  struct {
		HTML string `json:"html"`
	} `json:"links"`
}

type DastTargetSource struct {
	URL string `json:"url"`
}

func (t *DastTarget) FieldsAsRow() []string {
	return []string{t.Name, t.ID, t.URL, t.Team.Name, t.LabelsAsString(), t.Links.HTML}
}

func (t *DastTarget) LabelsAsString() string {
	var l string
	for i, label := range t.Labels {
		if i == 0 {
			l = label.Name
			continue
		}
		l += fmt.Sprintf(",%s", label.Name)
	}
	return l
}

type DastScanResponse struct {
	Message       string `json:"message"`
	DastScanID    string `json:"dast_scan_id"`
	DastTargetID  string `json:"dast_target_id"`
	DastProfileID string `json:"dast_profile_id"`
}

func (c *Client) StartDastTargetScan(project string) (*DastScanResponse, error) {
	if project == "" {
		return nil, errors.New("target id or name cannot be empty")
	}

	klog.Debugf("starting dast target scan for target: %s", project)

	path := fmt.Sprintf("/api/v2/dast_targets/%s/scan", project)
	req, err := c.newRequest(http.MethodPost, path, struct{}{})
	if err != nil {
		return nil, createHTTPRequestError(err)
	}

	var resp DastScanResponse
	httpResp, err := c.do(req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to start dast target scan: %w", err)
	}

	if httpResp.StatusCode != http.StatusCreated {
		return nil, httpStatusError(http.StatusCreated, httpResp.StatusCode)
	}

	return &resp, nil
}

func (c *Client) ListDastTargetScans(project string, params *ScanSearchParams) ([]ScanDetail, error) {
	if project == "" {
		return nil, errors.New("target id or name cannot be empty")
	}

	klog.Debugf("retrieving dast target scans for target: %s", project)

	path := fmt.Sprintf("/api/v2/dast_targets/%s/scans", project)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, createHTTPRequestError(err)
	}

	if params != nil {
		v, err := query.Values(params)
		if err != nil {
			return nil, fmt.Errorf("failed to encode query params: %w", err)
		}
		req.URL.RawQuery = v.Encode()
	}

	type response struct {
		Scans []ScanDetail `json:"scans"`
		Total int          `json:"total"`
	}
	var r response

	httpResp, err := c.do(req, &r)
	if err != nil {
		return nil, fmt.Errorf("failed to get dast target scans: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, httpStatusError(http.StatusOK, httpResp.StatusCode)
	}

	return r.Scans, nil
}

func (c *Client) ListDastTargets(name string) ([]DastTarget, error) {
	klog.Debug("retrieving dast target list...")

	req, err := c.newRequest(http.MethodGet, "/api/v2/dast_targets", nil)
	if err != nil {
		return nil, createHTTPRequestError(err)
	}

	if name != "" {
		q := req.URL.Query()
		q.Add("name", name)
		req.URL.RawQuery = q.Encode()
	}

	type response struct {
		Projects []DastTarget `json:"projects"`
		Total    int          `json:"total"`
	}
	var r response

	resp, err := c.do(req, &r)
	if err != nil {
		return nil, fmt.Errorf("failed to get dast targets: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpStatusError(http.StatusOK, resp.StatusCode)
	}

	return r.Projects, nil
}

func (c *Client) CreateDastTarget(pd ProjectDetail) (*Project, error) {
	klog.Debug("creating a dast target")

	req, err := c.newRequest(http.MethodPost, "/api/v2/dast_targets", pd)
	if err != nil {
		return nil, createHTTPRequestError(err)
	}

	type projectResponse struct {
		Project Project `json:"project"`
		Message string  `json:"message"`
	}
	var pr projectResponse
	_, err = c.do(req, &pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create dast target: %w", err)
	}

	return &pr.Project, nil
}
