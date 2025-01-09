package client

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/kondukto-io/kdt/klog"
)

func (c *Client) ImportEndpoint(filePath string, projectName string) error {
	klog.Debugf("importing endpoint using file:%s", filePath)

	if filePath == "" {
		return errors.New("file parameter is required")
	}

	projectDoc, err := c.FindProjectByName(projectName)
	if err != nil || projectDoc == nil {
		return fmt.Errorf("no projects found for name [%s]", projectName)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	path := fmt.Sprintf("/api/v2/projects/%s/openapispec", projectDoc.ID)
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Cookie", viper.GetString("token"))

	resp, err := c.do(req, nil)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to import endpoint: %s", resp.Status)
	}

	return nil
}
