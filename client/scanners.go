/*
Copyright © 2021 Kondukto
Created by Yusuf Eyisan aka @yeyisan
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
		ActiveScanners []struct {
			Id          string   `json:"id"`
			Type        string   `json:"type"`
			Slug        string   `json:"slug"`
			DisplayName string   `json:"display_name"`
			Labels      []string `json:"labels"`
			CustomType  int      `json:"custom_type"`
		} `json:"active_scanners"`
		Total int `json:"total"`
	}
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
