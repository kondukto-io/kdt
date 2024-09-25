/*
Copyright © 2019 Kondukto

*/

package client

import (
	"errors"
	"fmt"
	"net/http"
	"time"

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
		Projects []Project `json:"projects"`
		Total    int       `json:"total"`
	}
	var ps getProjectsResponse

	resp, err := c.do(req, &ps)
	if err != nil {
		return projects, err
	}

	if resp.StatusCode != http.StatusOK {
		return projects, fmt.Errorf("HTTP response not OK : %s", resp.Status)
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
	// It is only used for [feature] environment
	ForkSourceBranch string `json:"fork_source_branch"`
	// FeatureBranchRetention holds the number of days to delete the feature branch after the latest scan.
	FeatureBranchRetention uint `json:"feature_branch_retention"`
	// FeatureBranchInfiniteRetention holds a value that disables the feature branch retention period.
	FeatureBranchInfiniteRetention bool   `json:"feature_branch_no_retention"`
	DefaultBranch                  string `json:"default_branch"`
}

type ProjectSource struct {
	Tool          string    `json:"tool"`
	ID            string    `json:"id"`
	URL           string    `json:"url"`
	CloneDisabled bool      `json:"clone_disabled"`
	PathScope     PathScope `json:"path_scope"`
}

type PathScope struct {
	IncludeEmpty  bool   `json:"include_empty"`
	IncludedPaths string `json:"included_paths"`
	IncludedFiles string `json:"included_files"`
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
	ProgressStatus string             `json:"progress_status"`
	Status         string             `json:"status"`
	SAST           PlaybookTypeDetail `json:"sast"`
	DAST           PlaybookTypeDetail `json:"dast"`
	PENTEST        PlaybookTypeDetail `json:"pentest"`
	IAST           PlaybookTypeDetail `json:"iast"`
	SCA            PlaybookTypeDetail `json:"sca"`
	CS             PlaybookTypeDetail `json:"cs"`
	IAC            PlaybookTypeDetail `json:"iac"`
	MAST           PlaybookTypeDetail `json:"mast"`
}

const ReleaseStatusHistoryInprogress = "in_progress"

type PlaybookTypeDetail struct {
	Status string `json:"status" bson:"status"`
	ScanID string `json:"scan_id,omitempty" bson:"scan_id"`
}

type ReleaseStatusOpts struct {
	WaitTillComplete           bool
	TotalWaitDurationToTimeout time.Duration
	WaitDuration               time.Duration
}

func (c *Client) ReleaseStatus(project, branch string, opts ...ReleaseStatusOpts) (*ReleaseStatus, error) {
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

	return c.waitReleaseProgress(rs, project, branch, opts...)
}

func (c *Client) waitReleaseProgress(rs *ReleaseStatus, project, branch string, opts ...ReleaseStatusOpts) (*ReleaseStatus, error) {
	if len(opts) == 0 {
		return rs, nil
	}

	var opt = opts[0]
	if !opt.WaitTillComplete {
		return rs, nil
	}

	var timeout = time.After(opt.TotalWaitDurationToTimeout)

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout [%s] exceeded while waiting for release status", opt.TotalWaitDurationToTimeout)
		default:
			if rs.ProgressStatus != ReleaseStatusHistoryInprogress {
				return rs, nil
			}

			klog.Debugf("Release status is still in progress for project [%s] on branch [%s]. Waiting for 5 seconds...", project, branch)
			time.Sleep(time.Second * 5)

			var err error
			rs, err = c.ReleaseStatus(project, branch)
			if err != nil {
				return nil, fmt.Errorf("failed to get release status: %w", err)
			}
		}
	}
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
