/*
Copyright © 2021 Kondukto
*/

package client

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-querystring/query"

	"github.com/kondukto-io/kdt/klog"
)

type (
	ScanparamSearchParams struct {
		ToolID           string `url:"tool_id"`
		Branch           string `url:"branch"`
		Limit            int    `url:"limit"`
		MetaData         string `url:"meta_data"`
		Target           string `url:"target"`
		Manual           bool   `url:"manual"`
		Agent            string `url:"agent"`
		Environment      string `url:"environment"`
		ForkScan         bool   `url:"fork_scan"`
		ForkSourceBranch string `url:"fork_source_branch"`
		PR               bool   `url:"pr"`
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
		Tool        *ScanparamsItem `json:"tool"`
		Project     *ScanparamsItem `json:"project"`
		Agent       *ScanparamsItem `json:"agent"`
		BindName    string          `json:"bind_name"`
		Branch      string          `json:"branch"`
		ScanType    string          `json:"scan_type"`
		MetaData    string          `json:"meta_data"`
		ForkScan    bool            `json:"fork_scan"`
		PR          PRInfo          `json:"pr"`
		Manual      bool            `json:"manual"`
		Custom      Custom          `json:"custom"`
		Environment string          `json:"environment"`
	}

	ScanparamsItem struct {
		ID string `json:"id,omitempty"`
	}

	ScanParamsDeleteParams struct {
		ScanParamsID    string `json:"id"`
		ToolName        string `json:"tool_name"`
		Branch          string `json:"branch"`
		MetaData        string `json:"meta_data"`
		MetaDataIsEmpty bool   `json:"meta_data_is_empty"`
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

	path := fmt.Sprintf("/api/v2/projects/%s/scanparams", pID)
	req, err := c.newRequest(http.MethodPost, path, sp)
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

func (c *Client) DeleteScanparamsBy(pID string, deleteParams ScanParamsDeleteParams) error {
	klog.Debug("deleting scanparams")

	path := fmt.Sprintf("/api/v2/projects/%s/scanparams/delete", pID)
	req, err := c.newRequest(http.MethodPost, path, deleteParams)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}
