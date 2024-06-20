/*
Copyright © 2023 Kondukto

*/

package client

import "net/http"

type Team struct {
	Name             string           `json:"name"`
	IssueResponsible IssueResponsible `json:"issue_responsible"`
}

type IssueResponsible struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
}

func (c *Client) CreateTeam(teamName string, issueResponsible IssueResponsible) error {
	var team = Team{
		Name:             teamName,
		IssueResponsible: issueResponsible,
	}

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
