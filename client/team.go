/*
Copyright © 2023 Kondukto

*/

package client

import "net/http"

type Team struct {
	Name             string           `json:"name"`
	IssueResponsible IssueResponsible `json:"issue_responsible"`
	TeamAdmin        TeamAdmin        `json:"team_admin"`
}

type IssueResponsible struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
}

type TeamAdmin struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
}

func (c *Client) CreateTeam(team Team) error {
	req, err := c.newRequest(http.MethodPost, "/api/v2/teams", team)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}
