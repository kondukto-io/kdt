/*
Copyright © 2019 Kondukto

*/

package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"

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
	modeByFileImport = iota
	modeByScanID
	modeByProjectTool
	modeByProjectToolAndPR
	modeByProjectToolAndForkScan
	modeByProjectToolAndMetadata
	modeByImage
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "base command for starting scans",
	Run:   scanRootCommand,
	PreRun: func(cmd *cobra.Command, args []string) {
		// Initialize Kondukto client
		c, err := client.New()
		if err != nil {
			qwe(ExitCodeError, err, "could not initialize Kondukto client")
		}
		t, _ := cmd.Flags().GetString("tool")
		s, _ := cmd.Flags().GetString("scan-id")
		if s == "" && !c.IsValidTool(t) {
			qwm(ExitCodeError, "unknown or inactive tool name. Run `kdt list scanners` to see the supported active scanner's list.")
		}
	},
}

func scanRootCommand(cmd *cobra.Command, _ []string) {
	// Initialize Kondukto client
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}

	eventID, err := startScan(cmd, c)
	if err != nil {
		qwe(ExitCodeError, err, "failed to start scan")
	}

	async, err := cmd.Flags().GetBool("async")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse async flag")
	}

	// Do not wait for scan to finish if async set to true
	if async {
		eventRows := []Row{
			{Columns: []string{"EVENT ID"}},
			{Columns: []string{"--------"}},
			{Columns: []string{eventID}},
		}
		TableWriter(eventRows...)
		qwm(ExitCodeSuccess, "scan has been started with async parameter, exiting.")
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
	scanCmd.Flags().StringP("merge-target", "M", "", "source branch name for pull request")
	scanCmd.Flags().BoolP("fork-scan", "B", false, "enables a fork scan that based on project's default branch")
	scanCmd.Flags().Bool("override", false, "overrides old analysis results for the source branch")
	scanCmd.Flags().String("image", "", "image to scan with container security products")
	scanCmd.Flags().StringP("agent", "a", "", "specify the agent name for agent type scanners")

	scanCmd.Flags().Bool("threshold-risk", false, "set risk score of last scan as threshold")
	scanCmd.Flags().Int("threshold-crit", 0, "threshold for number of vulnerabilities with critical severity")
	scanCmd.Flags().Int("threshold-high", 0, "threshold for number of vulnerabilities with high severity")
	scanCmd.Flags().Int("threshold-med", 0, "threshold for number of vulnerabilities with medium severity")
	scanCmd.Flags().Int("threshold-low", 0, "threshold for number of vulnerabilities with low severity")

	scanCmd.Flags().Int("timeout", 0, "minutes to wait for scan to finish. scan will continue async if duration exceeds limit")
}

func startScan(cmd *cobra.Command, c *client.Client) (string, error) {
	switch getScanMode(cmd) {
	case modeByFileImport:
		// scan mode to start a scan by importing a file
		eventID, err := scanByFileImport(cmd, c)
		if err != nil {
			return "", err
		}
		return eventID, nil
	case modeByScanID:
		// scan mode to restart a scan with a known scan ID
		scanID, err := cmd.Flags().GetString("scan-id")
		if err != nil {
			return "", err
		}
		eventID, err := c.RestartScanByScanID(scanID)
		if err != nil {
			return "", err
		}
		return eventID, nil
	case modeByProjectTool:
		// scan mode to restart a scan with the given project and tool parameters
		eventID, err := startScanByProjectTool(cmd, c)
		if err != nil {
			return "", err
		}
		return eventID, nil
	case modeByProjectToolAndPR:
		// scan mode to restart a scan with the given project, tool and pr params
		eventID, err := startScanByProjectToolAndPR(cmd, c)
		if err != nil {
			return "", err
		}
		return eventID, nil
	case modeByProjectToolAndForkScan:
		// scan mode to restart a scan with the given project, tool and pr params
		eventID, err := findScanIDByProjectToolAndForkScan(cmd, c)
		if err != nil {
			return "", err
		}
		return eventID, nil
	case modeByProjectToolAndMetadata:
		// scan mode to restart a scan with the given project, tool and meta params
		scanID, err := getScanIDByProjectToolAndMeta(cmd, c)
		if err != nil {
			return "", err
		}
		eventID, err := c.RestartScanByScanID(scanID)
		if err != nil {
			qwe(ExitCodeError, err, "could not start scan")
		}
		return eventID, nil
	case modeByImage:
		eventID, err := scanByImage(cmd, c)
		if err != nil {
			qwe(ExitCodeError, err, "could not start scan")
		}
		return eventID, nil
	default:
		return "", errors.New("invalid scan mode")
	}
}

func checkForRescanOnlyTool(cmd *cobra.Command, c *client.Client) (bool, *client.ScannerInfo, error) {
	klog.Debugf("checking for rescan only tools")
	name, err := cmd.Flags().GetString("tool")
	if err != nil || name == "" {
		return false, nil, errors.New("missing require tool flag")
	}
	scanners, err := c.ListActiveScanners(&client.ScannersSearchParams{Name: name, Limit: 1})
	if err != nil {
		return false, nil, fmt.Errorf("failed to get active scanners: %w", err)
	}
	if scanners.Total == 0 {
		return false, nil, fmt.Errorf("invalid or inactive scanner tool name: %s", name)
	}
	scanner := scanners.ActiveScanners[0]
	for _, label := range scanner.Labels {
		if label == client.ScannerLabelBind ||
			label == client.ScannerLabelAgent ||
			label == client.ScannerLabelTemplate {
			return true, &scanner, nil
		}
	}

	return false, &scanner, nil
}

func getScanMode(cmd *cobra.Command) uint {
	// Check scan method
	byFile := cmd.Flag("file").Changed
	byTool := cmd.Flag("tool").Changed
	byMetaData := cmd.Flag("meta").Changed
	byScanId := cmd.Flag("scan-id").Changed
	byProject := cmd.Flag("project").Changed
	byBranch := cmd.Flag("merge-target").Changed
	byForkScan := cmd.Flag("fork-scan").Changed
	byMerge := cmd.Flag("branch").Changed
	byImage := cmd.Flag("image").Changed
	byPR := byBranch && byMerge

	byProjectAndTool := byProject && byTool
	byProjectAndToolAndMeta := byProjectAndTool && byMetaData && !byPR
	byProjectAndToolAndPullRequest := byProjectAndTool && byPR
	byProjectAndToolAndFile := byProjectAndTool && byFile
	byProjectAndToolAndForkScan := byProjectAndTool && byForkScan && !byPR

	mode := func() uint {
		// sorted by priority
		switch true {
		case byImage:
			return modeByImage
		case byProjectAndToolAndFile:
			return modeByFileImport
		case byScanId:
			return modeByScanID
		case byProjectAndToolAndPullRequest:
			return modeByProjectToolAndPR
		case byProjectAndToolAndForkScan:
			return modeByProjectToolAndForkScan
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
		qwe(ExitCodeError, err, "failed to parse timeout flag")
	}
	duration := time.Duration(timeoutFlag) * time.Minute

	lastStatus := -1
	for {
		event, err := c.GetScanStatus(eventID)
		if err != nil {
			qwe(ExitCodeError, err, "failed to get scan status")
		}

		switch event.Active {
		case eventFailed:
			eventRows := []Row{
				{Columns: []string{"EventID", "Event Status", "UI Link"}},
				{Columns: []string{"-------", "------------", "-------"}},
				{Columns: []string{event.ID, "Failed", event.Links.HTML}},
			}
			TableWriter(eventRows...)
			qwm(ExitCodeError, fmt.Sprintf("Scan failed. Reason: %s", event.Message))
		case eventInactive:
			if event.Status == jobFinished {
				klog.Println("scan finished successfully")
				scan, err := c.FindScanByID(event.ScanId)
				if err != nil {
					qwe(ExitCodeError, err, "failed to fetch scan summary")
				}

				// Printing scan results
				printScanSummary(scan)

				if err = passTests(scan, cmd); err != nil {
					qwe(ExitCodeError, err, "scan could not pass security tests")
				} else if err = checkRelease(scan); err != nil {
					qwe(ExitCodeError, err, "scan failed to pass release criteria")
				}
				qwm(ExitCodeSuccess, "scan passed security tests successfully")
			}
		case eventActive:
			if duration != 0 && time.Now().Sub(start) > duration {
				qwm(ExitCodeSuccess, "scan duration exceeds timeout, it will continue running async in the background")
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
			qwm(ExitCodeError, "invalid event status")
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

func passTests(scan *client.ScanDetail, cmd *cobra.Command) error {
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

func scanByFileImport(cmd *cobra.Command, c *client.Client) (string, error) {
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
	target, err := cmd.Flags().GetString("merge-target")
	if err != nil {
		return "", fmt.Errorf("failed to parse merge target flag: %w", err)
	}
	override, err := cmd.Flags().GetBool("override")
	if err != nil {
		return "", fmt.Errorf("failed to parse override flag: %w", err)
	}
	if override && target == "" {
		return "", errors.New("overriding PR analysis requires a merge target")
	}
	forkScan, err := cmd.Flags().GetBool("fork-scan")
	if err != nil {
		return "", fmt.Errorf("failed to parse fork-scan flag: %w", err)
	}
	if forkScan && target != "" {
		return "", errors.New("the fork-scan and pr-merge commands cannot be used together")
	}
	meta, err := cmd.Flags().GetString("meta")
	if err != nil {
		return "", fmt.Errorf("failed to parse meta flag: %w", err)
	}

	var form = client.ImportForm{
		"project":              project,
		"branch":               branch,
		"tool":                 tool,
		"target":               target,
		"fork-scan":            strconv.FormatBool(forkScan),
		"override-old-analyze": strconv.FormatBool(override),
		"meta":                 meta,
	}

	eventID, err := c.ImportScanResult(absoluteFilePath, form)
	if err != nil {
		return "", fmt.Errorf("failed to import scan results: %w", err)
	}

	return eventID, nil
}

func startScanByProjectTool(cmd *cobra.Command, c *client.Client) (string, error) {
	rescanOnly, scanner, err := checkForRescanOnlyTool(cmd, c)
	if err != nil {
		return "", err
	}

	// Parse command line flags
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}

	agent, err := cmd.Flags().GetString("agent")
	if err != nil {
		return "", fmt.Errorf("failed to parse agent flag: %w", err)
	}

	var agentID string
	if len(agent) > 0 {
		agentDetail, err := c.FindAgentByLabel(agent)
		if err != nil {
			return "", fmt.Errorf("failed to get agent: %w", err)
		}
		agentID = agentDetail.ID
	}

	params := &client.ScanSearchParams{
		Tool:    tool,
		Branch:  branch,
		PR:      false,
		Manual:  false,
		AgentID: agentID,
		Limit:   1,
	}

	scan, err := c.FindScan(project, params)
	if err == nil {
		klog.Print("a completed scan found with the same parameters, restarting")
		eventID, err := c.RestartScanByScanID(scan.ID)
		if err != nil {
			return "", err
		}
		return eventID, nil
	} else {
		klog.Debugf("failed to get completed scans: %v, trying to get scanparams", err)
	}

	sp, err := c.FindScanparams(project, &client.ScanparamSearchParams{
		ToolID: scanner.ID,
		Branch: branch,
		Manual: false,
		PR:     false,
		Agent:  agent,
		Limit:  1,
	})
	if err != nil {
		klog.Debugf("failed to get scanparams: %v, trying to create new scan", err)
	}
	scanData := &client.Scan{
		Branch:  branch,
		Project: project,
		ToolID:  scanner.ID,
		Custom:  client.Custom{Type: scanner.CustomType},
	}

	if sp != nil {
		klog.Debug("a scanparams found with the same parameters")
		scanData.ScanparamsID = sp.Id
		return c.CreateNewScan(scanData)
	}

	if rescanOnly && !scanner.HasLabel(client.ScannerLabelAgent) {
		klog.Debugf("scanner tool %s is only allowing rescans", tool)
		qwm(ExitCodeError, "no scans found for given project and tool configuration")
	}

	klog.Debug("no scanparams found with the same parameters, creating a new scan")
	if rescanOnly && scanner.HasLabel(client.ScannerLabelAgent) {
		agents, err := c.ListActiveAgents(&client.AgentSearchParams{Label: agent})
		if err != nil {
			klog.Debugf("failed to get active agents: %v")
			qwm(ExitCodeError, "failed to get active agents")
		}
		if agents.Total == 0 {
			klog.Debugf("no found agent to start scan: %v")
			qwm(ExitCodeError, "no found agent to start scan")
		}
		if agents.Total > 1 {
			klog.Debugf("[%d] agents found. Please specify it which one should be selected", agents.Total)
			qwm(ExitCodeError, "multiple agents found, please select one")
		}

		agent := agents.ActiveAgents.First()
		klog.Debugf("agent [%s] found. Setting scan with agent", agent.Label)
		scanData.AgentID = agent.ID
	}

	klog.Printf("creating a new scan")
	return c.CreateNewScan(scanData)
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
	meta, err := cmd.Flags().GetString("meta")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}

	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}

	params := &client.ScanSearchParams{
		Tool:     tool,
		MetaData: meta,
		Branch:   branch,
		PR:       false,
		Limit:    1,
	}

	scan, err := c.FindScan(project, params)
	if err != nil {
		klog.Debug("no scans found for given project, tool and metadata configuration")
		qwe(ExitCodeError, err, "could not get scans of the project")
	}

	return scan.ID, nil
}

func startScanByProjectToolAndPR(cmd *cobra.Command, c *client.Client) (string, error) {
	rescanOnly, scanner, err := checkForRescanOnlyTool(cmd, c)
	if err != nil {
		return "", err
	}
	// Parse command line flags
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}
	if branch == "" {
		return "", errors.New("missing branch field")
	}
	override, err := cmd.Flags().GetBool("override")
	if err != nil {
		return "", fmt.Errorf("failed to parse override flag: %w", err)
	}

	mergeTarget, err := cmd.Flags().GetString("merge-target")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	if mergeTarget == "" {
		return "", errors.New("missing merge-target field")
	}

	meta, err := cmd.Flags().GetString("meta")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}

	agent, err := cmd.Flags().GetString("agent")
	if err != nil {
		return "", fmt.Errorf("failed to parse agent flag: %w", err)
	}

	var agentID string
	if len(agent) > 0 {
		agentDetail, err := c.FindAgentByLabel(agent)
		if err != nil {
			return "", fmt.Errorf("failed to get agent: %w", err)
		}
		agentID = agentDetail.ID
	}

	params := &client.ScanSearchParams{
		Tool:     tool,
		MetaData: meta,
		AgentID:  agentID,
		Limit:    1,
	}

	scan, err := c.FindScan(project, params)
	if err == nil {
		opt := &client.ScanPROptions{
			From:               branch,
			To:                 mergeTarget,
			OverrideOldAnalyze: override,
		}
		eventID, err := c.RestartScanWithOption(scan.ID, opt)
		if err != nil {
			qwe(ExitCodeError, err, "could not start scan")
		}
		return eventID, nil
	} else {
		klog.Debugf("failed to get completed scans: %v, trying to get scanparams", err)
	}

	sp, err := c.FindScanparams(project, &client.ScanparamSearchParams{
		Branch: branch,
		ToolID: scanner.ID,
		Agent:  agent,
		Target: mergeTarget,
		PR:     true,
		Limit:  1,
	})
	if err != nil {
		klog.Debugf("failed to get scanparams: %v, trying to create a new scan", err)
	}

	var scanData = func() *client.Scan {
		if sp != nil {
			return &client.Scan{ScanparamsID: sp.Id}
		}

		if rescanOnly && !scanner.HasLabel(client.ScannerLabelAgent) {
			klog.Debugf("scanner tool %s is only allowing rescans", tool)
			qwm(ExitCodeError, "no scans found for given project, tool and PR configuration")
		}

		var scan = &client.Scan{
			Branch:  branch,
			Project: project,
			ToolID:  scanner.ID,
			Custom: client.Custom{
				Type: scanner.CustomType,
			},
			PR: client.PRInfo{
				OK:     true,
				Target: mergeTarget,
			},
		}

		if rescanOnly && scanner.HasLabel(client.ScannerLabelAgent) {
			agents, err := c.ListActiveAgents(&client.AgentSearchParams{Label: agent})
			if err != nil {
				klog.Debugf("failed to get active agents: %v")
				qwm(ExitCodeError, "failed to get active agents")
			}
			if agents.Total == 0 {
				klog.Debugf("no found agent to start scan: %v")
				qwm(ExitCodeError, "no found agent to start scan")
			}
			if agents.Total > 1 {
				klog.Debugf("[%d] agents found. Please specify it which one should be selected", agents.Total)
				qwm(ExitCodeError, "multiple agents found, please select one")
			}

			agent := agents.ActiveAgents.First()
			klog.Debugf("agent [%s] found. Setting scan with agent", agent.Label)
			scan.AgentID = agent.ID
		}

		return scan

	}()

	return c.CreateNewScan(scanData)
}

func findScanIDByProjectToolAndForkScan(cmd *cobra.Command, c *client.Client) (string, error) {
	rescanOnly, scanner, err := checkForRescanOnlyTool(cmd, c)
	if err != nil {
		return "", err
	}
	// Parse command line flags
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	branch, err := cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}
	if branch == "" {
		return "", errors.New("missing branch field")
	}

	meta, err := cmd.Flags().GetString("meta")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}

	forkScan, err := cmd.Flags().GetBool("fork-scan")
	if err != nil {
		return "", fmt.Errorf("failed to parse fork-scan flag: %w", err)
	}

	params := &client.ScanSearchParams{
		Tool:     tool,
		Branch:   branch,
		MetaData: meta,
		ForkScan: forkScan,
		Limit:    1,
	}

	scan, err := c.FindScan(project, params)
	if err == nil {
		eventID, err := c.RestartScanByScanID(scan.ID)
		if err != nil {
			qwe(1, err, "could not start scan")
		}
		return eventID, nil
	} else {
		klog.Debugf("failed to get completed scans: %v, trying to get scanparams", err)
	}

	sp, err := c.FindScanparams(project, &client.ScanparamSearchParams{
		ToolID:   scanner.ID,
		Branch:   branch,
		ForkScan: forkScan,
		MetaData: meta,
		Limit:    1,
	})
	if err != nil {
		klog.Debugf("failed to get scanparams: %v, trying to create a new scan", err)
	}

	var scanData = func() *client.Scan {
		if sp != nil {
			return &client.Scan{ScanparamsID: sp.Id}
		}

		if rescanOnly {
			klog.Debugf("scanner tool %s is only allowing rescans", tool)
			klog.Fatal("no scans found for given project, tool and PR configuration")
		}
		return &client.Scan{
			Branch:   branch,
			MetaData: meta,
			Project:  project,
			ForkScan: forkScan,
			ToolID:   scanner.ID,
			Custom:   client.Custom{Type: scanner.CustomType},
		}
	}()

	return c.CreateNewScan(scanData)
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

func checkRelease(scan *client.ScanDetail) error {
	c, err := client.New()
	if err != nil {
		return err
	}

	rs, err := c.ReleaseStatus(scan.Project)
	if err != nil {
		return fmt.Errorf("failed to get release status: %w", err)
	}

	const statusFail = "fail"

	if rs.Status == statusFail {
		return errors.New("project does not pass release criteria")
	}

	return nil
}

func printScanSummary(scan *client.ScanDetail) {
	s := scan.Summary
	name, id, branch, meta, tool, date := scan.Name, scan.ID, scan.Branch, scan.MetaData, scan.Tool, scan.Date.String()
	crit, high, med, low, score := strC(s.Critical), strC(s.High), strC(s.Medium), strC(s.Low), strC(scan.Score)
	scanSummaryRows := []Row{
		{Columns: []string{"NAME", "ID", "BRANCH", "META", "TOOL", "CRIT", "HIGH", "MED", "LOW", "SCORE", "DATE", "UI Link"}},
		{Columns: []string{"----", "--", "------", "----", "----", "----", "----", "---", "---", "-----", "----", "-------"}},
		{Columns: []string{name, id, branch, meta, tool, crit, high, med, low, score, date, scan.Links.HTML}},
	}

	TableWriter(scanSummaryRows...)
}
