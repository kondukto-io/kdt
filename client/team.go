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
	Username string `json:"username"`
}

func (c *Client) CreateTeam(teamName, responsible string) error {
	var team = Team{
		Name: teamName,
		IssueResponsible: IssueResponsible{
			Username: responsible,
		},
	}

	req, err := c.newRequest(http.MethodPost, "/api/v3/teams", team)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}
