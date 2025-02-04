package pkg

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/kondukto-io/kdt/klog"
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
	if installedVersion == "" {
		klog.Debug("installed version is not defined, the update checking is skipped")
		return false, installedVersion
	}

	client := &http.Client{
		Timeout: time.Second * 3,
		Transport: &loggingRoundTripper{
			next: http.DefaultTransport,
		},
	}

	url := fmt.Sprintf("https://github.com/%s/%s/releases/latest", owner, repo)

	resp, err := client.Head(url)
	if err != nil {
		klog.Debugf("failed to get latest version from %s: %v", url, err)
		return false, installedVersion

	}
	defer resp.Body.Close()

	if len(location) == 0 {
		klog.Debug("could not get the location of the version request")
		return false, installedVersion
	}

	locationParts := strings.Split(location, "/")
	lastVersion := locationParts[len(locationParts)-1]

	currentVersion, err := version.NewVersion(installedVersion)
	if err != nil {
		klog.Debugf("failed to parsing version [%s]: %v", installedVersion, err)
		return false, installedVersion
	}

	latestVersion, err := version.NewVersion(lastVersion)
	if err != nil {
		klog.Debugf("failed to parsing version [%s]: %v", lastVersion, err)
		return false, installedVersion
	}

	if latestVersion.GreaterThan(currentVersion) {
		return true, lastVersion
	}

	return false, installedVersion
}
