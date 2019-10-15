package client

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Scan struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	MetaData string     `json:"meta_data"`
	Tool     string     `json:"tool"`
	Date     *time.Time `json:"date"`
	Summary  struct {
		Critical int `json:"critical"`
		High     int `json:"high"`
		Medium   int `json:"medium"`
		Low      int `json:"low"`
		Info     int `json:"info"`
	} `json:"summary"`
}

func (c *Client) ListScans(project string) ([]Scan, error) {
	scans := make([]Scan, 0)

	path := fmt.Sprintf("/api/v1/projects/%s/scans", project)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return scans, err
	}

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

func (c *Client) ScanByScanId(id string) (string, error) {
	path := fmt.Sprintf("/api/v1/scans/%s/restart", id)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return "", err
	}

	panic("not implemented")
}

func (c *Client) ScanByProjectAndTool(project string, tool string) (string, error) {
	panic("not implemented")
}
