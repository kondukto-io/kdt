/*
Copyright © 2019 Kondukto

*/

package client

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/kondukto-io/kdt/klog"
)

type Product struct {
	ID            string `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	ProjectsCount int    `json:"projects_count"`
	Links         struct {
		HTML string `json:"html"`
	} `json:"links"`
}

func (p *Product) FieldsAsRow() []string {
	return []string{p.Name, p.ID, strconv.Itoa(p.ProjectsCount), p.Links.HTML}
}

func (c *Client) ListProducts(name string) ([]Product, error) {
	products := make([]Product, 0)

	klog.Debug("retrieving product list...")

	req, err := c.newRequest("GET", "/api/v2/products", nil)
	if err != nil {
		return products, err
	}

	queryParams := req.URL.Query()
	queryParams.Add("name", name)
	req.URL.RawQuery = queryParams.Encode()

	type getProductsResponse struct {
		Products []Product `json:"data"`
		Total    int       `json:"total"`
		Error    string    `json:"error"`
	}

	var ps getProductsResponse
	resp, err := c.do(req, &ps)
	if err != nil {
		return products, err
	}

	if resp.StatusCode != http.StatusOK {
		return products, fmt.Errorf("HTTP response not OK : %s", ps.Error)
	}

	return ps.Products, nil
}

type ProductDetail struct {
	Name     string    `json:"name"`
	Projects []Project `json:"projects"`
}

func (c *Client) CreateProduct(pd ProductDetail) (*Product, error) {
	klog.Debug("creating a product")

	req, err := c.newRequest(http.MethodPost, "/api/v2/products", pd)
	if err != nil {
		return nil, err
	}

	type productResponse struct {
		Product Product `json:"product"`
		Message string  `json:"message"`
	}
	var pr productResponse
	_, err = c.do(req, &pr)
	if err != nil {
		return nil, err
	}

	return &pr.Product, nil
}
