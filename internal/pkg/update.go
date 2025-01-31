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
func CheckUpdate(installedVersion string) (bool, string) {
	client := &http.Client{
		Timeout: time.Second * 3,
		Transport: &loggingRoundTripper{
			next: http.DefaultTransport,
		},
	}

	url := fmt.Sprintf("https://github.com/%s/%s/releases/latest", owner, repo)

	resp, err := client.Head(url)
	if err != nil {
		return false, installedVersion
	}
	defer resp.Body.Close()

	if len(location) == 0 {
		return false, installedVersion
	}

	// Extract version from URL
	locationParts := strings.Split(location, "/")
	lastVersion := locationParts[len(locationParts)-1]

	if installedVersion == "v1.40.1" {
		// downgrade the installed version to v1.0.40
		installedVersion = "v1.0.40"
	}

	currentVersion, err := version.NewVersion(installedVersion)
	if err != nil {
		return false, installedVersion
	}

	latestVersion, err := version.NewVersion(lastVersion)
	if err != nil {
		return false, installedVersion
	}

	if latestVersion.GreaterThan(currentVersion) {
		return true, lastVersion
	}

	return false, installedVersion
}
