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

type Event struct {
	ID      string `json:"id"`
	Status  int    `json:"status"`
	Active  int    `json:"active"`
	ScanId  string `json:"scan_id"`
	Message string `json:"message"`
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

func (c *Client) GetScanStatus(eventId string) (int, int, error) {
	path := fmt.Sprintf("/api/v1/events/%s/status", eventId)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return -1, -1, err
	}

	type eventStatusResponse struct {
		Status  int    `json:"status"`
		Active  int    `json:"active"`
		Message string `json:"message"`
	}

	var esr eventStatusResponse
	resp, err := c.do(req, &esr)
	if err != nil {
		return -1, -1, err
	}

	if resp.StatusCode != http.StatusOK {
		return -1, -1, errors.New("response not ok")
	}

	return esr.Status, esr.Active, nil
}

func (c *Client) ScanByProjectAndTool(project string, tool string) (string, error) {
	panic("not implemented")
}
