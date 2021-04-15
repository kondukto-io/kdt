/*
Copyright © 2019 Kondukto

*/
package client

import (
	"errors"
	"fmt"
	"net/http"
)

type Project struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (c *Client) ListProjects(arg string) ([]Project, error) {
	projects := make([]Project, 0)

	req, err := c.newRequest("GET", "/api/v1/projects", nil)
	if err != nil {
		return projects, err
	}

	queryParams := req.URL.Query()
	queryParams.Add("search", arg)
	req.URL.RawQuery = queryParams.Encode()

	type getProjectsResponse struct {
		Projects []Project `json:"data"`
		Total    int       `json:"total"`
	}
	var ps getProjectsResponse

	resp, err := c.do(req, &ps)
	if err != nil {
		return projects, err
	}

	if resp.StatusCode != http.StatusOK {
		return projects, errors.New("response not ok")
	}

	return ps.Projects, nil
}

type ReleaseStatus struct {
	Status string `json:"status" bson:"status"`
	SAST   struct {
		Tool   string `json:"tool" bson:"tool"`
		Status string `json:"status" bson:"status"`
		ScanID string `json:"scan_id,omitempty" bson:"scan_id"`
	} `json:"sast" bson:"sast"`
	DAST struct {
		Tool   string `json:"tool" bson:"tool"`
		Status string `json:"status" bson:"status"`
		ScanID string `json:"scan_id,omitempty" bson:"scan_id"`
	} `json:"dast" bson:"dast"`
	SCA struct {
		Tool   string `json:"tool" bson:"tool"`
		Status string `json:"status" bson:"status"`
		ScanID string `json:"scan_id,omitempty" bson:"scan_id"`
	} `json:"sca" bson:"sca"`
}

func (c *Client) ReleaseStatus(project string) (*ReleaseStatus, error) {
	if project == "" {
		return nil, errors.New("invalid project id or name")
	}

	path := fmt.Sprintf("/api/v1/projects/%s/release", project)

	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	rs := new(ReleaseStatus)

	resp, err := c.do(req, rs)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("response not ok")
	}

	return rs, nil
}
