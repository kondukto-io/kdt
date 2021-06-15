/*
Copyright © 2019 Kondukto

*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/kondukto-io/kdt/klog"

	"github.com/kondukto-io/kdt/client"
	"github.com/spf13/cobra"
)

const (
	jobStarting = iota
	jobRunning
	jobAnalyzing
	jobNotifying
	jobFinished
)

const (
	eventFailed = iota - 1
	eventInactive
	eventActive
)

const (
	modeByFile = iota
	modeByScanID
	modeByProjectTool
	modeByProjectToolAndPR
	modeByProjectToolAndMetadata
	modeByImage
)

var byTool string

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "base command for starting scans",
	Run:   scanRootCommand,
	PreRun: func(cmd *cobra.Command, args []string) {
		t, _ := cmd.Flags().GetString("tool")
		if !validTool(t) {
			klog.Fatal("a valid tool must be given. Run `kdt list scanners` to see the supported scanner's list.")
		}
	},
}

func scanRootCommand(cmd *cobra.Command, args []string) {
	// Initialize Kondukto client
	c, err := client.New()
	if err != nil {
		qwe(1, err, "could not initialize Kondukto client")
	}

	eventID, err := startScan(cmd, c)
	if err != nil {
		klog.Fatalf("failed to start scan: %v", err)
	}

	async, err := cmd.Flags().GetBool("async")
	if err != nil {
		klog.Fatalf("failed to parse async flag: %v", err)
	}

	// Do not wait for scan to finish if async set to true
	if async {
		qwm(0, "scan has been started with async parameter, exiting.")
	}

	waitTillScanEnded(cmd, c, eventID)
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().Bool("async", false, "does not block build process")

	scanCmd.Flags().StringP("project", "p", "", "project name or id")
	scanCmd.Flags().StringP("tool", "t", "", "tool name")
	scanCmd.Flags().StringP("scan-id", "s", "", "scan id")
	scanCmd.Flags().StringP("meta", "m", "", "meta data")
	scanCmd.Flags().StringP("file", "f", "", "scan file")
	scanCmd.Flags().StringP("branch", "b", "", "branch")
	scanCmd.Flags().StringP("merge-target", "M", "", "target branch name for pull request")
	scanCmd.Flags().String("image", "", "image to scan with container security products")

	scanCmd.Flags().Bool("threshold-risk", false, "set risk score of last scan as threshold")
	scanCmd.Flags().Int("threshold-crit", 0, "threshold for number of vulnerabilities with critical severity")
	scanCmd.Flags().Int("threshold-high", 0, "threshold for number of vulnerabilities with high severity")
	scanCmd.Flags().Int("threshold-med", 0, "threshold for number of vulnerabilities with medium severity")
	scanCmd.Flags().Int("threshold-low", 0, "threshold for number of vulnerabilities with low severity")

	scanCmd.Flags().Int("timeout", 0, "minutes to wait for scan to finish. scan will continue async if duration exceeds limit")
}

func startScan(cmd *cobra.Command, c *client.Client) (string, error) {
	var err error
	var scanID string
	switch getScanMode(cmd) {
	case modeByScanID:
		// scan mode to restart a scan with a known scan ID
		scanID, err = cmd.Flags().GetString("scan-id")
		if err != nil {
			return "", err
		}
	case modeByFile:
		// scan mode to start a scan by importing a file
		eventID, err := scanByFile(cmd, c)
		if err != nil {
			return "", err
		}
		return eventID, nil

	case modeByProjectTool:
		// scan mode to restart a scan with the given project and tool params
		scanID, err = getScanIDByProjectTool(cmd, c)
		if err != nil {
			return "", err
		}
	case modeByProjectToolAndMetadata:
		// scan mode to restart a scan with the given project, tool and meta params
		scanID, err = getScanIDByProjectToolAndMeta(cmd, c)
		if err != nil {
			return "", err
		}
	case modeByProjectToolAndPR:
		// scan mode to restart a scan with the given project, tool and pr params
		scanID, opt, err := getScanIDByProjectToolAndPR(cmd, c)
		if err != nil {
			return "", err
		}

		eventID, err := c.StartScanByOption(scanID, opt)
		if err != nil {
			qwe(1, err, "could not start scan")
		}

		return eventID, nil
	case modeByImage:
		eventID, err := scanByImage(cmd, c)
		if err != nil {
			qwe(1, err, "could not start scan")
		}
		return eventID, nil
	default:
		return "", errors.New("invalid scan mode")
	}

	eventID, err := c.StartScanByScanId(scanID)
	if err != nil {
		qwe(1, err, "could not start scan")
	}

	return eventID, nil
}

func getScanMode(cmd *cobra.Command) uint {
	// Check scan method
	byFile := cmd.Flag("file").Changed
	byTool := cmd.Flag("tool").Changed
	byMetaData := cmd.Flag("meta").Changed
	byScanId := cmd.Flag("scan-id").Changed
	byProject := cmd.Flag("project").Changed
	byBranch := cmd.Flag("merge-target").Changed
	byMerge := cmd.Flag("branch").Changed
	byPR := byBranch && byMerge
	byImage := cmd.Flag("image").Changed

	byProjectAndTool := byProject && byTool && !byMetaData
	byProjectAndToolAndMeta := byProjectAndTool && byMetaData && !byPR
	byProjectAndToolAndPullRequest := byProjectAndTool && byPR
	byProjectAndToolAndFile := byProjectAndTool && byFile && !byMetaData

	mode := func() uint {
		// sorted by priority
		switch true {
		case byImage:
			return modeByImage
		case byProjectAndToolAndFile:
			return modeByFile
		case byScanId:
			return modeByScanID
		case byProjectAndToolAndPullRequest:
			return modeByProjectToolAndPR
		case byProjectAndTool:
			return modeByProjectTool
		case byProjectAndToolAndMeta:
			return modeByProjectToolAndMetadata
		default:
			return modeByScanID
		}
	}()
	return mode
}

func waitTillScanEnded(cmd *cobra.Command, c *client.Client, eventID string) {
	start := time.Now()
	timeoutFlag, err := cmd.Flags().GetInt("timeout")
	if err != nil {
		qwe(1, err, "failed to parse timeout flag")
	}
	duration := time.Duration(timeoutFlag) * time.Minute

	lastStatus := -1
	for {
		event, err := c.GetScanStatus(eventID)
		if err != nil {
			klog.Fatalf("failed to get scan status: %v", err)
		}

		switch event.Active {
		case eventFailed:
			qwm(1, "scan failed")
		case eventInactive:
			if event.Status == jobFinished {
				klog.Println("scan finished successfully")
				scan, err := c.GetScanSummary(event.ScanId)
				if err != nil {
					qwe(1, err, "failed to fetch scan summary")
				}

				// Printing scan results
				printScanSummary(scan)

				if err := passTests(scan, cmd); err != nil {
					qwe(1, err, "scan could not pass security tests")
				} else if err := checkRelease(cmd); err != nil {
					qwe(1, err, "scan failed to pass release criteria")
				}
				qwm(0, "scan passed security tests successfully")
			}
		case eventActive:
			if duration != 0 && time.Now().Sub(start) > duration {
				qwm(0, "scan duration exceeds timeout, it will continue running async in the background")
			}
			if event.Status != lastStatus {
				klog.Println(statusMsg(event.Status))
				lastStatus = event.Status
				// Get new scans scan id
			} else {
				klog.Debugf("event status [%s]", statusMsg(event.Status))
			}
			time.Sleep(10 * time.Second)
		default:
			qwm(1, "invalid event status")
		}
	}
}

func statusMsg(s int) string {
	switch s {
	case jobStarting:
		return "starting scan"
	case jobRunning:
		return "scan running"
	case jobAnalyzing:
		return "analyzing scan results"
	case jobNotifying:
		return "setting notifications"
	case jobFinished:
		return "scan finished"
	default:
		return "unknown scan status"
	}
}

func passTests(scan *client.Scan, cmd *cobra.Command) error {
	c, err := client.New()
	if err != nil {
		return err
	}

	if cmd.Flag("threshold-risk").Changed {
		m, err := c.GetLastResults(scan.ID)
		if err != nil {
			return err
		}

		if m["last"] == nil || m["previous"] == nil {
			return errors.New("missing score records")
		}

		if m["last"].Score > m["previous"].Score {
			return errors.New("risk score of the scan is higher than last scan's")
		}
	}

	if cmd.Flag("threshold-crit").Changed {
		crit, err := cmd.Flags().GetInt("threshold-crit")
		if err != nil {
			return err
		}
		if scan.Summary.Critical > crit {
			return errors.New("number of vulnerabilities with critical severity is higher than threshold")
		}
	}

	if cmd.Flag("threshold-high").Changed {
		high, err := cmd.Flags().GetInt("threshold-high")
		if err != nil {
			return err
		}
		if scan.Summary.High > high {
			return errors.New("number of vulnerabilities with high severity is higher than threshold")
		}
	}

	if cmd.Flag("threshold-med").Changed {
		med, err := cmd.Flags().GetInt("threshold-med")
		if err != nil {
			return err
		}
		if scan.Summary.Medium > med {
			return errors.New("number of vulnerabilities with medium severity is higher than threshold")
		}
	}

	if cmd.Flag("threshold-low").Changed {
		low, err := cmd.Flags().GetInt("threshold-low")
		if err != nil {
			return err
		}
		if scan.Summary.Low > low {
			return errors.New("number of vulnerabilities with low severity is higher than threshold")
		}
	}

	return nil
}

func scanByFile(cmd *cobra.Command, c *client.Client) (string, error) {
	// Parse command line flags needed for file uploads
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	if !cmd.Flag("branch").Changed {
		return "", errors.New("branch parameter is required to import scan results")
	}

	pathToFile, err := cmd.Flags().GetString("file")
	if err != nil {
		return "", fmt.Errorf("failed to parse file path: %w", err)
	}
	absoluteFilePath, err := filepath.Abs(pathToFile)
	if err != nil {
		return "", fmt.Errorf("failed to parse absolute path: %w", err)
	}
	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}

	eventID, err := c.ImportScanResult(project, branch, tool, absoluteFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to import scan results: %w", err)
	}

	return eventID, nil
}

func getScanIDByProjectTool(cmd *cobra.Command, c *client.Client) (string, error) {
	// Parse command line flags
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}

	if !validTool(tool) {
		return "", errors.New("invalid tool name")
	}

	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}

	params := &client.ScanSearchParams{
		Tool:   tool,
		Branch: branch,
		Limit:  1,
	}

	scan, err := c.FindScan(project, params)
	if err != nil {
		klog.Fatal("no scans found for given project and tool configuration")
	}

	return scan.ID, nil
}

func getScanIDByProjectToolAndMeta(cmd *cobra.Command, c *client.Client) (string, error) {
	// Parse command line flags
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}

	if !validTool(tool) {
		return "", errors.New("invalid tool name")
	}

	meta, err := cmd.Flags().GetString("meta")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}

	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}

	params := &client.ScanSearchParams{
		Tool:   tool,
		Meta:   meta,
		Branch: branch,
		Limit:  1,
	}

	scan, err := c.FindScan(project, params)
	if err != nil {
		klog.Fatal("no scans found for given project, tool and metadata configuration")
		qwe(1, err, "could not get scans of the project")
	}

	return scan.ID, nil
}

func getScanIDByProjectToolAndPR(cmd *cobra.Command, c *client.Client) (string, *client.ScanPROptions, error) {
	// Parse command line flags
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := cmd.Flags().GetString("tool")
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse tool flag: %w", err)
	}

	if !validTool(tool) {
		return "", nil, errors.New("invalid tool name")
	}

	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse tool flag: %w", err)
	}
	if branch == "" {
		return "", nil, errors.New("missing branch field")
	}

	mergeTarget, err := cmd.Flags().GetString("merge-target")
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse tool flag: %w", err)
	}
	if mergeTarget == "" {
		return "", nil, errors.New("missing merge-target field")
	}

	meta, err := cmd.Flags().GetString("meta")
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse tool flag: %w", err)
	}

	params := &client.ScanSearchParams{
		Tool:  tool,
		Meta:  meta,
		Limit: 1,
	}

	scan, err := c.FindScan(project, params)
	if err != nil {
		klog.Fatalf("no scans found for given project, tool and PR configuration")
		qwe(1, err, "could not get scans of the project")
	}
	if scan == nil {
		qwm(1, "no found scan by given parameters")
	}
	opt := &client.ScanPROptions{
		From: branch,
		To:   mergeTarget,
	}
	return scan.ID, opt, nil
}

func scanByImage(cmd *cobra.Command, c *client.Client) (string, error) {
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	if !validTool(tool) {
		return "", errors.New("invalid tool name")
	}
	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}
	image, err := cmd.Flags().GetString("image")
	if err != nil {
		return "", fmt.Errorf("failed to parse image flag: %w", err)
	}
	if image == "" {
		return "", errors.New("image name is required")
	}

	eventID, err := c.ScanByImage(project, branch, tool, image)
	if err != nil {
		return "", err
	}

	return eventID, nil
}

func checkRelease(cmd *cobra.Command) error {
	c, err := client.New()
	if err != nil {
		return err
	}
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return fmt.Errorf("project flag parsing error: %v", err)
	}

	rs, err := c.ReleaseStatus(project)
	if err != nil {
		return fmt.Errorf("failed to get release status: %w", err)
	}

	const statusFail = "fail"

	if rs.Status == statusFail {
		return errors.New("project does not pass release criteria")
	}

	return nil
}

func printScanSummary(scan *client.Scan) {
	w := tabwriter.NewWriter(os.Stdout, 8, 8, 4, ' ', 0)
	_, _ = fmt.Fprintf(w, "NAME\tID\tMETA\tTOOL\tCRIT\tHIGH\tMED\tLOW\tINFO\tDATE\n")
	_, _ = fmt.Fprintf(w, "---\t---\t---\t---\t---\t---\t---\t---\t---\t---\n")
	_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\n", scan.Name, scan.ID, scan.MetaData, scan.Tool,
		scan.Summary.Critical, scan.Summary.High, scan.Summary.Medium, scan.Summary.Low, scan.Summary.Info, scan.Date)
	_ = w.Flush()
}
