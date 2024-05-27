/*
Copyright © 2019 Kondukto

*/

package client

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/kondukto-io/kdt/klog"
)

type Project struct {
	ID            string         `json:"id,omitempty"`
	Name          string         `json:"name,omitempty"`
	DefaultBranch string         `json:"default_branch"`
	Labels        []ProjectLabel `json:"labels"`
	Team          ProjectTeam    `json:"team"`
	Links         struct {
		HTML string `json:"html"`
	} `json:"links"`
}

func (p *Project) FieldsAsRow() []string {
	return []string{p.Name, p.ID, p.DefaultBranch, p.Team.Name, p.LabelsAsString(), p.Links.HTML}
}

func (p *Project) LabelsAsString() string {
	var l string
	for i, label := range p.Labels {
		if i == 0 {
			l = label.Name
			continue
		}
		l += fmt.Sprintf(",%s", label.Name)
	}
	return l
}

func (c *Client) ListProjects(name, repo string) ([]Project, error) {
	projects := make([]Project, 0)

	klog.Debug("retrieving project list...")

	req, err := c.newRequest("GET", "/api/v2/projects", nil)
	if err != nil {
		return projects, err
	}

	queryParams := req.URL.Query()
	queryParams.Add("name", name)
	queryParams.Add("alm", repo)
	req.URL.RawQuery = queryParams.Encode()

	type getProjectsResponse struct {
		Projects []Project `json:"data"`
		Total    int       `json:"total"`
		Error    string    `json:"error"`
	}
	var ps getProjectsResponse

	resp, err := c.do(req, &ps)
	if err != nil {
		return projects, err
	}

	if resp.StatusCode != http.StatusOK {
		return projects, fmt.Errorf("HTTP response not OK : %s", ps.Error)
	}

	return ps.Projects, nil
}

func (c *Client) FindProjectByName(name string) (*Project, error) {
	projects, err := c.ListProjects(name, "")
	if err != nil {
		return nil, err
	}

	for _, p := range projects {
		if p.Name == name {
			return &p, nil
		}
	}

	return nil, ProjectNotFound
}

type ProjectDetail struct {
	Name      string         `json:"name"`
	Source    ProjectSource  `json:"source"`
	Team      ProjectTeam    `json:"team"`
	Labels    []ProjectLabel `json:"labels"`
	Override  bool           `json:"override"`  // That means, if the project already exists, create a new one with suffix "-"
	Overwrite bool           `json:"overwrite"` // That means, if the project already exists, overwrite it
	// ForkSourceBranch holds the name of the branch to be used as the source for the fork scan.
	// It is only used for [feature] environemnt
	ForkSourceBranch string `json:"fork_source_branch"`
	// FeatureBranchRetention holds the number of days to delete the feature branch after the latest scan.
	FeatureBranchRetention uint `json:"feature_branch_retention"`
	// FeatureBranchInfiniteRetention holds a value that disables the feature branch retention period.
	FeatureBranchInfiniteRetention bool `json:"feature_branch_no_retention"`
}

type ProjectSource struct {
	Tool string `json:"tool"`
	ID   string `json:"id"`
	URL  string `json:"url"`
}
type ProjectTeam struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}
type ProjectLabel struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

func (c *Client) CreateProject(pd ProjectDetail) (*Project, error) {
	klog.Debug("creating a project")

	req, err := c.newRequest(http.MethodPost, "/api/v2/projects", pd)
	if err != nil {
		return nil, err
	}

	type projectResponse struct {
		Project Project `json:"project"`
		Message string  `json:"message"`
	}
	var pr projectResponse
	_, err = c.do(req, &pr)
	if err != nil {
		return nil, err
	}

	return &pr.Project, nil
}

type ReleaseStatus struct {
	Status  string             `json:"status"`
	SAST    PlaybookTypeDetail `json:"sast"`
	DAST    PlaybookTypeDetail `json:"dast"`
	PENTEST PlaybookTypeDetail `json:"pentest"`
	IAST    PlaybookTypeDetail `json:"iast"`
	SCA     PlaybookTypeDetail `json:"sca"`
	CS      PlaybookTypeDetail `json:"cs"`
	IAC     PlaybookTypeDetail `json:"iac"`
}

type PlaybookTypeDetail struct {
	Status string `json:"status" bson:"status"`
	ScanID string `json:"scan_id,omitempty" bson:"scan_id"`
}

func (c *Client) ReleaseStatus(project, branch string) (*ReleaseStatus, error) {
	if project == "" {
		return nil, errors.New("missing project id or name")
	}

	path := fmt.Sprintf("/api/v2/projects/%s/release", project)

	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	if branch != "" {
		queryParams := req.URL.Query()
		queryParams.Add("branch", branch)
		req.URL.RawQuery = queryParams.Encode()
	}

	rs := new(ReleaseStatus)

	resp, err := c.do(req, rs)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP response not OK: %d", resp.StatusCode)
	}

	return rs, nil
}

func (c *Client) IsAvailable(project, almTool string) (bool, error) {
	path := fmt.Sprintf("/api/v2/projects/check/%s/%s", almTool, project)

	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return false, err
	}

	var checkProjectResponse struct {
		Exist bool `json:"exist"`
	}

	resp, err := c.do(req, &checkProjectResponse)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP response not OK: %d", resp.StatusCode)
	}

	return checkProjectResponse.Exist, nil
}
