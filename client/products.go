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

func (c *Client) FindProductByName(name string) (*Product, error) {
	products, err := c.ListProducts(name)
	if err != nil {
		return nil, err
	}

	for _, p := range products {
		if p.Name == name {
			return &p, nil
		}
	}

	return nil, ProductNotFound
}

func (c *Client) GetProductDetail(id string) (*ProductDetail, error) {
	klog.Debug("retrieving product list...")

	path := fmt.Sprintf("/api/v2/products/%s", id)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	type getProductsResponse struct {
		Product ProductDetail `json:"product"`
		Error   string        `json:"error"`
	}

	var ps getProductsResponse
	resp, err := c.do(req, &ps)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP response not OK : %s", ps.Error)
	}

	return &ps.Product, nil
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
		Products []Product `json:"products"`
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
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Projects         []Project `json:"projects"`
	BusinessUnitTags []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Color    string `json:"color"`
		IsActive int    `json:"isActive"`
		Required bool   `json:"required"`
	} `json:"business_unit_tags"`
	AccessibleFor struct {
		OwnerIDs []string `json:"owner_ids"`
		TeamIDs  []string `json:"team_ids"`
	} `json:"accessible_for"`
	DefaultTeam struct {
		ID string `json:"id"`
	} `json:"default_team"`
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

func (c *Client) UpdateProduct(id string, pd ProductDetail) (*Product, error) {
	klog.Debug("updating a product")

	path := fmt.Sprintf("/api/v2/products/%s", id)
	req, err := c.newRequest(http.MethodPatch, path, pd)
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
