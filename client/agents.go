/*
Copyright © 2021 Kondukto
*/

package client

import (
	"fmt"
	"net/http"

	"github.com/google/go-querystring/query"
	"github.com/kondukto-io/kdt/klog"
)

type (
	AgentSearchParams struct {
		Limit int `url:"limit"`
	}
	AgentsResponse struct {
		ActiveAgents []Agent `json:"active_agents"`
		Total        int     `json:"total"`
	}
	Agent struct {
		ID       string `json:"id"`
		Label    string `json:"label"`
		Url      string `json:"url"`
		AgentId  string `json:"agent_id"`
		Password string `json:"password"`
		Insecure bool   `json:"insecure"`
		IsActive int    `json:"isActive"`
	}
)

func (c *Client) ListActiveAgents(params *AgentSearchParams) (*AgentsResponse, error) {
	klog.Debugf("retrieving active agents")

	path := fmt.Sprintf("/api/v2/agents")
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	v, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = v.Encode()

	var agentsResponse AgentsResponse
	res, err := c.do(req, &agentsResponse)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP response not OK: %d", res.StatusCode)
	}

	return &agentsResponse, nil
}
