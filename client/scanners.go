/*
Copyright © 2021 Kondukto
*/

package client

import (
	"fmt"
	"net/http"

	"github.com/google/go-querystring/query"
	"github.com/kondukto-io/kdt/klog"
)

type (
	ScannersSearchParams struct {
		Types  string `url:"types"`
		Labels string `url:"labels"`
		Name   string `url:"name"`
		Limit  int    `url:"limit"`
	}
	ScannersResponse struct {
		ActiveScanners []ScannerInfo `json:"active_scanners"`
		Total          int           `json:"total"`
	}
	ScannerInfo struct {
		ID          string   `json:"id"`
		Type        string   `json:"type"`
		Slug        string   `json:"slug"`
		DisplayName string   `json:"display_name"`
		Labels      []string `json:"labels"`
		CustomType  int      `json:"custom_type"`
	}
)

const (
	ScannerLabelKDT      = "kdt"
	ScannerLabelBind     = "bind"
	ScannerLabelAgent    = "agent"
	ScannerLabelDocker   = "docker"
	ScannerLabelImport   = "import"
	ScannerLabelTemplate = "template"
)

func (c *Client) ListActiveScanners(params *ScannersSearchParams) (*ScannersResponse, error) {
	klog.Debugf("retrieving active scanners")

	path := fmt.Sprintf("/api/v1/scanners/active")
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	v, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = v.Encode()

	var scanners ScannersResponse
	res, err := c.do(req, &scanners)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP response not OK: %d", res.StatusCode)
	}

	return &scanners, nil
}

func (c *Client) IsValidTool(tool string) bool {
	klog.Debugf("validating tool name")

	scanners, err := c.ListActiveScanners(&ScannersSearchParams{Name: tool})
	if err != nil {
		klog.Debugf("failed to get active scanners: %v", err)
		return false
	}

	if scanners.Total == 0 {
		klog.Debugf("invalid or inactive tool name: %s", tool)
		return false
	}

	return true
}
