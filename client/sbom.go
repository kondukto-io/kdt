package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/kondukto-io/kdt/klog"
	"github.com/spf13/viper"
)

func (c *Client) ImportSBOM(file string, repo string, form ImportForm) error {
	klog.Debugf("importing sbom content using file:%s", file)

	projectId := form["project"]

	if projectId != "" {
		projects, err := c.ListProjects(form["project"], repo)
		if err != nil {
			return err
		}

		if len(projects) == 1 {
			projectId = projects[0].ID
			form["project"] = projects[0].Name
		}

		if len(projects) > 1 {
			return errors.New("multiple projects found for given parameters")
		}
	} else if repo != "" {
		projects, err := c.ListProjects(form["project"], repo)
		if err != nil {
			return err
		}

		if len(projects) == 1 {
			projectId = projects[0].ID
			form["project"] = projects[0].Name
		}

		if len(projects) > 1 {
			return errors.New("multiple projects found for given parameters")
		}
	}

	path := fmt.Sprintf("/api/v2/%s/sbom/upload", projectId)
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return err
	}
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	part, err := writer.CreateFormFile("file", filepath.Base(f.Name()))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f)
	if err != nil {
		return err
	}

	for k := range form {
		if err = writer.WriteField(k, form[k]); err != nil {
			return fmt.Errorf("failed to write form field [%s]: %w", k, err)
		}
	}

	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("X-Cookie", viper.GetString("token"))

	type importSBOMResponse struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	var importResponse importSBOMResponse
	resp, err := c.do(req, &importResponse)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to import sbom: %s", importResponse.Error)
	}

	return nil
}
