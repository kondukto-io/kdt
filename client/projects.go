package client

type Project struct {
}

func (c *Client) ListProjects() ([]Project, error) {
	projects := make([]Project, 0)

	return projects, nil
}
