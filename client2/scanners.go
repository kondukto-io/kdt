/*
Copyright © 2021 Kondukto
*/

package client

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/kondukto-io/kdt/klog"

	"github.com/google/go-querystring/query"
)

type (
	ScannersSearchParams struct {
		Types  string `url:"types"`
		Labels string `url:"labels"`
		Name   string `url:"name"`
		Limit  int    `url:"limit"`
	}
	ScannersResponse struct {
		ActiveScanners ActiveScanners `json:"active_scanners"`
		Total          int            `json:"total"`
	}

	ActiveScanners []ScannerInfo
	ScannerInfo    struct {
		ID          string        `json:"id"`
		Type        string        `json:"type"`
		Slug        string        `json:"slug"`
		DisplayName string        `json:"display_name"`
		Labels      []string      `json:"labels"`
		CustomType  int           `json:"custom_type"`
		Disabled    bool          `json:"disabled"`
		Params      ScannerParams `json:"params"`
	}

	// ScannerParams holds the custom parameters for a scanner
	ScannerParams map[string]ScannerCustomParams

	// ScannerCustomParams holds the details of a custom parameter
	ScannerCustomParams struct {
		Examples     string                  `json:"examples,omitempty"`
		Description  string                  `json:"description"`
		DefaultValue string                  `json:"default_value"`
		Optional     bool                    `json:"optional"`
		Type         scannerCustomParamsType `json:"type"`
	}
)

type scannerCustomParamsType string

const (
	scannerCustomParamsTypeString      scannerCustomParamsType = "string"
	scannerCustomParamsTypeStringSlice scannerCustomParamsType = "string_slice"
	scannerCustomParamsTypeInt         scannerCustomParamsType = "int"
	scannerCustomParamsTypeUInt        scannerCustomParamsType = "uint"
	scannerCustomParamsTypeBoolean     scannerCustomParamsType = "bool"
)

// Find returns the given key detail when present, otherwise nil.
func (s ScannerParams) Find(k string) *ScannerCustomParams {
	if v, ok := s[k]; ok {
		return &v
	}
	return nil
}

// RequiredParamsLen returns the required params length.
func (s ScannerParams) RequiredParamsLen() int {
	var count int
	for _, v := range s {
		if !v.Optional {
			count++
		}
	}

	return count
}

// Parse parses the given string into expected type
func (s ScannerCustomParams) Parse(k string) (interface{}, error) {
	switch s.Type {
	case scannerCustomParamsTypeStringSlice:
		return strings.Replace(k, ";", ",", -1), nil
	case scannerCustomParamsTypeString:
		return k, nil
	case scannerCustomParamsTypeInt:
		i, err := strconv.Atoi(k)
		if err != nil {
			return nil, err
		}
		return i, nil
	case scannerCustomParamsTypeUInt:
		i, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	case scannerCustomParamsTypeBoolean:
		return strconv.ParseBool(k)
	}
	return nil, fmt.Errorf("unknown scanner custom param type: %s", s.Type)
}

const (
	ScannerLabelKDT             = "kdt"
	ScannerLabelBind            = "bind"
	ScannerLabelAgent           = "agent"
	ScannerLabelDocker          = "docker"
	ScannerLabelImport          = "import"
	ScannerLabelTemplate        = "template"
	ScannerLabelCreatableOnTool = "creatable-on-tool"
)

// ListActiveScanners returns a list of active scanners
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

// IsValidTool returns true if the given tool name is a valid tool
func (c *Client) IsValidTool(tool string) bool {
	klog.Debugf("validating given tool name [%s]", tool)

	scanners, err := c.ListActiveScanners(&ScannersSearchParams{Name: tool})
	if err != nil {
		klog.Debugf("failed to get active tools: %v", err)
		return false
	}

	if scanners.Total == 0 {
		klog.Debugf("no tool found by given tool name. invalid or inactive tool name: %s", tool)
		return false
	}

	var scanner = scanners.ActiveScanners[0]
	if scanner.Disabled {
		klog.Printf("the scanner [%s] is disabled on the Kondukto", tool)
		return false
	}

	return true
}

// IsRescanOnlyLabel returns true if the given label is a rescan only label
// If fork scan is true, then the ScannerLabelTemplate label is not a rescan only label, it can be used for fork scan
func IsRescanOnlyLabel(label string, isForkScan bool) bool {
	if isForkScan {
		return false
	}
	if label == ScannerLabelBind || label == ScannerLabelAgent || label == ScannerLabelTemplate {
		return true
	}
	return false
}

// HasLabel returns true if the given label is present in the receiver's labels
func (s ScannerInfo) HasLabel(l string) bool {
	for _, label := range s.Labels {
		if label == l {
			return true
		}
	}
	return false
}

// First returns the first element in the list.
func (s ActiveScanners) First() *ScannerInfo {
	if len(s) == 0 {
		return nil
	}
	return &s[0]
}
