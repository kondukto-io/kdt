package pkg

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
)

var location = ""

const (
	owner     = "kondukto-io"
	repo      = "kdt"
	delimeter = "/tag/"
)

type loggingRoundTripper struct {
	next http.RoundTripper
}

func (l loggingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if strings.Contains(r.URL.String(), delimeter) {
		location = r.URL.String()
	}
	return l.next.RoundTrip(r)
}

// checks if there is a new version
func CheckUpdate(ver string) (bool, string) {
	client := &http.Client{
		Timeout: time.Second * 3,
		Transport: &loggingRoundTripper{
			next: http.DefaultTransport,
		},
	}

	url := fmt.Sprintf("https://github.com/%s/%s/releases/latest", owner, repo)

	resp, err := client.Head(url)
	if err != nil {
		return false, ver

	}
	defer resp.Body.Close()

	if len(location) > 0 {
		s := strings.Split(location, "/")
		lastVersion := s[len(s)-1]

		c, err := version.NewVersion(ver)
		if err != nil {
			return false, ver
		}

		l, err := version.NewVersion(lastVersion)
		if err != nil {
			return false, ver
		}

		if l.GreaterThan(c) {
			return true, lastVersion
		}
	}

	return false, ver
}
