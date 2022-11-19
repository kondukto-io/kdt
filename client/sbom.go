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
	"strconv"

	"github.com/kondukto-io/kdt/klog"
	"github.com/spf13/viper"
)

func (c *Client) ImportSBOM(file string, repo string, form ImportForm) error {
	klog.Debugf("importing sbom content using file:%s", file)

	projectName := form["project"]
	if projectName == "" && repo == "" {
		return errors.New("project and repo parameter values can not be empty same time")
	}

	projects, err := c.ListProjects(projectName, repo)
	if err != nil {
		return fmt.Errorf("no projects found for name [%s] and repo [%s]", projectName, repo)
	}

	if len(projects) == 0 {
		return fmt.Errorf("no projects found for name [%s] and repo [%s]", projectName, repo)
	}

	if len(projects) > 1 {
		return fmt.Errorf("multiple projects found for name [%s] and repo [%s]", projectName, repo)
	}

	allowEmptyParam := form["allow_empty"]
	if allowEmptyParam == "" {
		allowEmptyParam = "false"
	}

	_, err = strconv.ParseBool(allowEmptyParam)
	if err != nil {
		return fmt.Errorf("can not parse allow_empty parameter value [%s]", allowEmptyParam)
	}

	path := fmt.Sprintf("/api/v2/projects/%s/sbom/upload", projects[0].ID)
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
		return fmt.Errorf("failed to import sbom: %s, status code: %d", importResponse.Error, resp.StatusCode)
	}

	return nil
}
