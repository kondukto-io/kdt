/*
Copyright © 2021 Kondukto
Created by Yusuf Eyisan aka @yeyisan
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
		ToolID string `url:"tool_id"`
		Branch string `url:"branch"`
		Limit  int    `url:"limit"`
	}
	ScanparamResponse struct {
		Data  []ScanparamsDetail `json:"data"`
		Total int                `json:"total"`
	}
	ScanparamsDetail struct {
		Id       string `json:"id"`
		Branch   string `json:"branch"`
		BindName string `json:"bind_name"`
	}
)

func (c *Client) FindScanparams(project string, params *ScanparamSearchParams) (*ScanparamResponse, error) {
	klog.Debugf("retrieving scanparams")
	if project == "" {
		return nil, errors.New("missing project identifier")
	}

	path := fmt.Sprintf("/api/v1/projects%s/scanparams", project)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	v, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = v.Encode()

	var scanparams ScanparamResponse
	_, err = c.do(req, &scanparams)
	if err != nil {
		return nil, err
	}

	return &scanparams, nil
}
