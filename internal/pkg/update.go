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

// CheckUpdate checks if there is a new version
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

	if len(location) == 0 {
		return false, ver
	}

	// Extract version from URL
	locationParts := strings.Split(location, "/")
	lastVersion := locationParts[len(locationParts)-1]

	if lastVersion == "v1.40.1" {
		// Set the correct version for "v1.40.0"
		lastVersion = "v1.0.41"
		return true, lastVersion
	}

	currentVersion, err := version.NewVersion(ver)
	if err != nil {
		return false, ver
	}

	latestVersion, err := version.NewVersion(lastVersion)
	if err != nil {
		return false, ver
	}

	if latestVersion.GreaterThan(currentVersion) {
		return true, lastVersion
	}

	return false, ver
}
