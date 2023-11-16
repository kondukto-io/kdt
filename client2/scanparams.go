/*
Copyright © 2021 Kondukto
*/

package client

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/google/go-querystring/query"
	"github.com/kondukto-io/kdt/klog"
)

type (
	ScanparamSearchParams struct {
		ToolID   string `url:"tool_id"`
		Branch   string `url:"branch"`
		Limit    int    `url:"limit"`
		MetaData string `url:"meta_data"`
		Target   string `url:"target"`
		Manual   bool   `url:"manual"`
		Agent    string `url:"agent"`
		ForkScan bool   `url:"fork_scan"`
		PR       bool   `url:"pr"`
	}
	ScanparamResponse struct {
		Scanparams []Scanparams `json:"scanparams"`
		Limit      int          `json:"limit"`
		Start      int          `json:"start"`
		Total      int          `json:"total"`
	}
	Scanparams struct {
		ID       string  `json:"id"`
		Branch   string  `json:"branch"`
		BindName string  `json:"bind_name"`
		Custom   *Custom `json:"custom"`
	}

	ScanparamsDetail struct {
		Tool     *ScanparamsItem `json:"tool"`
		Project  *ScanparamsItem `json:"project"`
		Agent    *ScanparamsItem `json:"agent"`
		BindName string          `json:"bind_name"`
		Branch   string          `json:"branch"`
		ScanType string          `json:"scan_type"`
		MetaData string          `json:"meta_data"`
		ForkScan bool            `json:"fork_scan"`
		PR       PRInfo          `json:"pr"`
		Manual   bool            `json:"manual"`
		Custom   Custom          `json:"custom"`
	}

	ScanparamsItem struct {
		ID string `json:"id,omitempty"`
	}
)

func (c *Client) FindScanparams(project string, params *ScanparamSearchParams) (*Scanparams, error) {
	klog.Debugf("retrieving scanparams")
	if project == "" {
		return nil, errors.New("missing project identifier")
	}

	path := fmt.Sprintf("/api/v2/projects/%s/scanparams", project)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	v, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = v.Encode()

	var scanparamsResponse ScanparamResponse
	_, err = c.do(req, &scanparamsResponse)
	if err != nil {
		return nil, err
	}

	if scanparamsResponse.Total == 0 {
		return nil, errors.New("scanparams not found")
	}

	return &scanparamsResponse.Scanparams[0], nil
}

func (c *Client) CreateScanparams(pID string, sp ScanparamsDetail) (*Scanparams, error) {
	klog.Debug("creating a scanparams")

	req, err := c.newRequest(http.MethodPost, filepath.Join("/api/v2/projects", pID, "scanparams"), sp)
	if err != nil {
		return nil, err
	}

	type scanparamsResponse struct {
		Scanparams Scanparams `json:"scanparams"`
		Message    string     `json:"message"`
	}

	var pr scanparamsResponse
	_, err = c.do(req, &pr)
	if err != nil {
		return nil, err
	}

	return &pr.Scanparams, nil
}
