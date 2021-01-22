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
		Tool  string `json:"tool,omitempty"`
		Meta  string `json:"meta,omitempty"`
		Limit int    `json:"limit,omitempty"`
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
	req, err := c.newRequest("GET", path, nil)
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
	req, err := c.newRequest("GET", path, nil)
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
		return "", errors.New("")
	}

	return rsr.Event, nil
}

func (c *Client) GetScanStatus(eventId string) (*Event, error) {
	path := fmt.Sprintf("/api/v1/events/%s/status", eventId)
	req, err := c.newRequest("GET", path, nil)
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
	req, err := c.newRequest("GET", path, nil)
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
	req, err := c.newRequest("GET", path, nil)
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

func (c *Client) ImportScanResult(project, branch, tool string, files []string) error {
	path := "/api/v1/scans/import"
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return err
		}
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		part, err := writer.CreateFormFile("files", filepath.Base(f.Name()))
		if err != nil {
			return err
		}
		_, err = io.Copy(part, f)
		if err != nil {
			return err
		}
	}
	if err := writer.WriteField("project", project); err != nil {
		return err
	}
	if err := writer.WriteField("branch", branch); err != nil {
		return err
	}
	if err := writer.WriteField("tool", tool); err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Cookie", viper.GetString("token"))

	type importScanResultResponse struct {
		Message string `json:"message"`
	}
	var isrr importScanResultResponse
	resp, err := c.do(req, &isrr)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to import scan results")
	}

	return nil
}
