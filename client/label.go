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
	var newLabel = Label{
		Name:  label.Name,
		Color: label.Color,
	}

	req, err := c.newRequest(http.MethodPost, "/api/v3/labels", newLabel)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}
