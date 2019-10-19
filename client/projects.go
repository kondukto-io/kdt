/*
Copyright © 2019 Kondukto

*/
package client

import (
	"errors"
	"net/http"
)

type Project struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (c *Client) ListProjects() ([]Project, error) {
	projects := make([]Project, 0)

	req, err := c.newRequest("GET", "/api/v1/projects", nil)
	if err != nil {
		return projects, err
	}

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
