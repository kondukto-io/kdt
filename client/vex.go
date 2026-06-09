/*
Copyright © 2019 Invicti Security
https://www.invicti.com/
*/

package client

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type VexImport struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	Filename     string    `json:"filename"`
	DocumentID   string    `json:"document_id,omitempty"`
	Author       string    `json:"author,omitempty"`
	Version      int       `json:"version,omitempty"`
	DocumentTime time.Time `json:"document_time,omitempty"`
	StatementCnt int       `json:"statement_count"`
	UploadedBy   string    `json:"uploaded_by"`
	UploadedAt   time.Time `json:"uploaded_at"`
	Type         string    `json:"type"`
	Tooling      string    `json:"tooling,omitempty"`
}

type VexStatement struct {
	ID                   string    `json:"id"`
	ImportID             string    `json:"import_id"`
	VulnerabilityID      string    `json:"vulnerability_id"`
	VulnerabilityAliases []string  `json:"vulnerability_aliases,omitempty"`
	PURL                 string    `json:"purl,omitempty"`
	State                string    `json:"state"`
	Justification        string    `json:"justification,omitempty"`
	ImpactStatement      string    `json:"impact_statement,omitempty"`
	ActionStatement      string    `json:"action_statement,omitempty"`
	StatusNotes          string    `json:"status_notes,omitempty"`
	StatementTimestamp   time.Time `json:"statement_timestamp,omitempty"`
	DocumentTimestamp    time.Time `json:"document_timestamp,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

func (c *Client) ListVexImports(projectID string) ([]VexImport, int, error) {
	path := fmt.Sprintf("/api/v2/projects/%s/vex/imports", projectID)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, 0, err
	}

	q := req.URL.Query()
	q.Set("start", "0")
	q.Set("limit", "20")
	req.URL.RawQuery = q.Encode()

	var res struct {
		Total int         `json:"total"`
		Data  []VexImport `json:"data"`
	}
	if _, err = c.do(req, &res); err != nil {
		return nil, 0, err
	}

	return res.Data, res.Total, nil
}

func (c *Client) GetVexImport(projectID, importID string, includeStatements bool) (*VexImport, []VexStatement, error) {
	path := fmt.Sprintf("/api/v2/projects/%s/vex/imports/%s", projectID, importID)
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	if includeStatements {
		q := req.URL.Query()
		q.Set("include_statements", "true")
		req.URL.RawQuery = q.Encode()
	}

	var res struct {
		Data       VexImport      `json:"data"`
		Statements []VexStatement `json:"statements,omitempty"`
	}
	if _, err = c.do(req, &res); err != nil {
		return nil, nil, err
	}

	return &res.Data, res.Statements, nil
}

func (c *Client) UploadVexImport(projectID, filePath string) (*VexImport, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(f.Name()))
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(part, f); err != nil {
		return nil, err
	}
	_ = writer.Close()

	path := fmt.Sprintf("/api/v2/projects/%s/vex/imports", projectID)
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Cookie", viper.GetString("token"))

	var res struct {
		Data VexImport `json:"data"`
	}
	if _, err = c.do(req, &res); err != nil {
		return nil, err
	}

	return &res.Data, nil
}

func (c *Client) DeleteVexImport(projectID, importID string) error {
	path := fmt.Sprintf("/api/v2/projects/%s/vex/imports/%s", projectID, importID)
	req, err := c.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	var res struct {
		Message string `json:"message"`
	}
	_, err = c.do(req, &res)
	return err
}
