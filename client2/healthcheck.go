/*
Copyright © 2023 Kondukto

*/

package client

// HealthCheck is a healthcheck for Kondukto service
// Requires a valid API token
func (c *Client) HealthCheck() error {
	req, err := c.newRequest("GET", "/api/v2/health/check", nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}

// Ping is a healthcheck for Kondukto service
// Does not require a valid API token
func (c *Client) Ping() error {
	req, err := c.newRequest("GET", "/core/version", nil)
	if err != nil {
		return err
	}

	_, err = c.do(req, nil)
	if err != nil {
		return err
	}

	return nil
}
