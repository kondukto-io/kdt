/*
Copyright © 2023 Kondukto

*/

package client

import "net/http"

type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (c *Client) CreateLabel(label Label) error {
	req, err := c.newRequest(http.MethodPost, "/api/v2/labels", label)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}
