/*
Copyright © 2019 Kondukto

*/

package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kondukto-io/kdt/client"
	"github.com/kondukto-io/kdt/klog"
)

const (
	eventStatusWaiting           = 0
	eventStatusStarting          = 1
	eventStatusRunning           = 2
	eventStatusRetrievingResults = 3
	eventStatusAnalyzing         = 4
	eventStatusNotifying         = 5
	eventStatusFinished          = 6
	eventStatusFailed            = -1
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
	modeByProjectToolAndPRNumber
	modeByProjectToolAndForkScan
	modeByProjectToolAndMetadata
	modeByImage
)

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().Bool("async", false, "does not block build process")
	scanCmd.Flags().StringP("project", "p", "", "kondukto project id or name")
	scanCmd.Flags().StringP("tool", "t", "", "tool name")
	scanCmd.Flags().StringP("scan-id", "s", "", "scan id")
	scanCmd.Flags().StringP("meta", "m", "", "meta data")
	scanCmd.Flags().StringP("file", "f", "", "scan result file")
	scanCmd.Flags().StringP("branch", "b", "", "branch")
	scanCmd.Flags().StringP("merge-target", "M", "", "source branch name for pull-request")
	scanCmd.Flags().StringP("pr-number", "", "", "pull-request number. supported alms[github, gitlab, azure, bitbucket]")
	scanCmd.Flags().Bool("no-decoration", false, "no decoration for pr number")
	scanCmd.Flags().StringP("image", "I", "", "image to scan with container security products")
	scanCmd.Flags().StringP("agent", "a", "", "agent name for agent type scanners")
	scanCmd.Flags().BoolP("fork-scan", "B", false, "enables a fork scan that based on project's default branch")
	scanCmd.Flags().BoolP("break-by-scanner-type", "", false, "breaks pipeline if only scanner type matches with the given scanner's type")
	scanCmd.Flags().Bool("override", false, "overrides old analysis results for the source branch")
	scanCmd.Flags().Bool("create-project", false, "creates a new project when no project is found with the given parameters")
	scanCmd.Flags().StringSlice("params", nil, "parameters for the scan")

	scanCmd.Flags().StringP("labels", "l", "", "comma separated label names [create-project]")
	scanCmd.Flags().StringP("team", "T", "", "project team name [create-project]")
	scanCmd.Flags().StringP("repo-id", "r", "", "URL or ID of ALM repository [create-project]")
	scanCmd.Flags().String("alm-tool", "A", "ALM tool name [create-project]")
	scanCmd.Flags().StringP("product-name", "P", "", "name for product")

	scanCmd.Flags().Bool("threshold-risk", false, "set risk score of last scan as threshold")
	scanCmd.Flags().Int("threshold-crit", 0, "threshold for number of vulnerabilities with critical severity")
	scanCmd.Flags().Int("threshold-high", 0, "threshold for number of vulnerabilities with high severity")
	scanCmd.Flags().Int("threshold-med", 0, "threshold for number of vulnerabilities with medium severity")
	scanCmd.Flags().Int("threshold-low", 0, "threshold for number of vulnerabilities with low severity")

	scanCmd.Flags().Int("timeout", 0, "minutes to wait for scan to finish. scan will continue async if duration exceeds limit")
}

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
			qwm(ExitCodeError, "unknown, disabled or inactive tool name. Run `kdt list scanners` to see the supported active scanner's list.")
		}
	},
}

func scanRootCommand(cmd *cobra.Command, _ []string) {
	// Initialize Kondukto client
	c, err := client.New()
	if err != nil {
		qwe(ExitCodeError, err, "could not initialize Kondukto client")
	}
	scan := Scan{
		cmd:    cmd,
		client: c,
	}
	eventID, err := scan.startScan()
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

type Scan struct {
	cmd    *cobra.Command
	client *client.Client
}

func (s *Scan) startScan() (string, error) {
	switch getScanMode(s.cmd) {
	case modeByFileImport:
		// scan mode to start a scan by importing a file
		eventID, err := s.scanByFileImport()
		if err != nil {
			return "", err
		}
		return eventID, nil
	case modeByScanID:
		// scan mode to restart a scan with a known scan ID
		scanID, err := s.cmd.Flags().GetString("scan-id")
		if err != nil {
			return "", err
		}
		if !primitive.IsValidObjectID(scanID) {
			return "", errors.New("invalid object id")
		}
		eventID, err := s.client.RestartScanByScanID(scanID)
		if err != nil {
			return "", err
		}
		return eventID, nil
	case modeByProjectTool:
		// scan mode to restart a scan with the given project and tool parameters
		eventID, err := s.startScanByProjectTool()
		if err != nil {
			return "", err
		}
		return eventID, nil
	case modeByProjectToolAndPR:
		// scan mode to restart a scan with the given project, tool and pr params
		eventID, err := s.startScanByProjectToolAndPR()
		if err != nil {
			return "", err
		}
		return eventID, nil
	case modeByProjectToolAndPRNumber:
		// scan mode to restart a scan with the given project, tool and pr number
		eventID, err := s.startScanByProjectToolAndPRNumber()
		if err != nil {
			return "", err
		}
		return eventID, nil
	case modeByProjectToolAndForkScan:
		// scan mode to restart a scan with the given project, tool and pr params
		eventID, err := s.findScanIDByProjectToolAndForkScan()
		if err != nil {
			return "", err
		}
		return eventID, nil
	case modeByImage:
		eventID, err := s.scanByImage()
		if err != nil {
			qwe(ExitCodeError, err, "could not start scan")
		}
		return eventID, nil
	default:
		return "", errors.New("invalid scan mode")
	}
}

func (s *Scan) scanByImage() (string, error) {
	project, err := s.findORCreateProject()
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := s.cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	branch, err := s.cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}
	meta, err := s.cmd.Flags().GetString("meta")
	if err != nil {
		return "", fmt.Errorf("failed to parse meta flag: %w", err)
	}
	image, err := s.cmd.Flags().GetString("image")
	if err != nil {
		return "", fmt.Errorf("failed to parse image flag: %w", err)
	}
	if image == "" {
		return "", errors.New("image name is required")
	}
	var pr = &client.ImageScanParams{
		Project:  project.ID,
		Tool:     tool,
		Branch:   branch,
		Image:    image,
		MetaData: meta,
	}
	eventID, err := s.client.ScanByImage(pr)
	if err != nil {
		return "", err
	}

	return eventID, nil
}

func (s *Scan) scanByFileImport() (string, error) {
	// Parse command line flags needed for file uploads
	project, err := s.findORCreateProject()
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}

	tool, err := s.cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	if !s.cmd.Flag("branch").Changed {
		return "", errors.New("branch parameter is required to import scan results")
	}
	pathToFile, err := s.cmd.Flags().GetString("file")
	if err != nil {
		return "", fmt.Errorf("failed to parse file path: %w", err)
	}
	absoluteFilePath, err := filepath.Abs(pathToFile)
	if err != nil {
		return "", fmt.Errorf("failed to parse absolute path: %w", err)
	}
	branch, err := s.cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}
	meta, err := s.cmd.Flags().GetString("meta")
	if err != nil {
		return "", fmt.Errorf("failed to parse meta flag: %w", err)
	}
	target, err := s.cmd.Flags().GetString("merge-target")
	if err != nil {
		return "", fmt.Errorf("failed to parse merge target flag: %w", err)
	}
	override, err := s.cmd.Flags().GetBool("override")
	if err != nil {
		return "", fmt.Errorf("failed to parse override flag: %w", err)
	}
	if override && target == "" {
		return "", errors.New("overriding PR analysis requires a merge target")
	}
	forkScan, err := s.cmd.Flags().GetBool("fork-scan")
	if err != nil {
		return "", fmt.Errorf("failed to parse fork-scan flag: %w", err)
	}
	if forkScan && target != "" {
		return "", errors.New("the fork-scan and pr-merge commands cannot be used together")
	}

	var form = client.ImportForm{
		"project":              project.Name,
		"branch":               branch,
		"tool":                 tool,
		"meta_data":            meta,
		"target":               target,
		"fork-scan":            strconv.FormatBool(forkScan),
		"override_old_analyze": strconv.FormatBool(override),
	}

	eventID, err := s.client.ImportScanResult(absoluteFilePath, form)
	if err != nil {
		return "", fmt.Errorf("failed to import scan results: %w", err)
	}

	return eventID, nil
}

func (s *Scan) startScanByProjectTool() (string, error) {
	rescanOnly, scanner, err := s.checkForRescanOnlyTool()
	if err != nil {
		return "", err
	}
	// Parse command line flags
	project, err := s.findORCreateProject()
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := s.cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	branch, err := s.cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}
	agent, err := s.cmd.Flags().GetString("agent")
	if err != nil {
		return "", fmt.Errorf("failed to parse agent flag: %w", err)
	}
	meta, err := s.cmd.Flags().GetString("meta")
	if err != nil {
		return "", fmt.Errorf("failed to parse meta flag: %w", err)
	}

	var agentID string
	if len(agent) > 0 {
		agentDetail, err := s.client.FindAgentByLabel(agent)
		if err != nil {
			return "", fmt.Errorf("failed to get agent: %w", err)
		}
		agentID = agentDetail.ID
	}

	params := &client.ScanSearchParams{
		Tool:     tool,
		Branch:   branch,
		PR:       false,
		Manual:   false,
		AgentID:  agentID,
		MetaData: meta,
		Limit:    1,
	}

	scan, err := s.client.FindScan(project.Name, params)
	if err != nil {
		klog.Debugf("failed to get completed scans: %v, trying to get scanparams", err)

	} else if !s.cmd.Flags().Changed("params") {
		klog.Printf("a completed scan [%s] found with the same parameters, restarting", scan.ID)

		eventID, err := s.client.RestartScanByScanID(scan.ID)
		if err != nil {
			return "", err
		}
		return eventID, nil
	}

	sp, err := s.client.FindScanparams(project.Name, &client.ScanparamSearchParams{
		ToolID:   scanner.ID,
		Branch:   branch,
		Manual:   false,
		PR:       false,
		Agent:    agent,
		MetaData: meta,
		Limit:    1,
	})
	if err != nil {
		klog.Debugf("failed to get scanparams: %v, trying to create new scan", err)
	}

	var custom = client.Custom{Type: scanner.CustomType}
	if s.cmd.Flags().Changed("params") {
		custom = s.parseCustomParams(custom, *scanner, sp)
	}

	scanData := &client.Scan{
		MetaData: meta,
		Branch:   branch,
		Project:  project.Name,
		ToolID:   scanner.ID,
		Custom:   custom,
	}

	if sp != nil {
		klog.Debugf("a scanparams [%s] found with the same parameters", sp.ID)

		scanData.ScanparamsID = sp.ID
		return s.client.CreateNewScan(scanData)
	}

	if rescanOnly && !scanner.HasLabel(client.ScannerLabelAgent) && !s.cmd.Flags().Changed("params") {
		klog.Debugf("scanner tool [%s] is only allowing rescans", tool)

		qwm(ExitCodeError, "no scans found for given project and tool configuration")
	}

	scanparamsData := client.ScanparamsDetail{
		Branch:   branch,
		MetaData: meta,
		Custom:   custom,
		ScanType: "kdt",
		Tool: &client.ScanparamsItem{
			ID: scanner.ID,
		},
		Agent: &client.ScanparamsItem{
			ID: scanData.AgentID,
		},
		Project: &client.ScanparamsItem{
			ID: project.ID,
		},
	}

	klog.Debug("no scanparams found with the same parameters, creating a new scan")
	if rescanOnly && scanner.HasLabel(client.ScannerLabelAgent) {
		agents, err := s.client.ListActiveAgents(&client.AgentSearchParams{Label: agent})
		if err != nil {
			klog.Debugf("failed to get active agents: %v", err)
			qwm(ExitCodeError, "failed to get active agents")
		}
		if agents.Total == 0 {
			klog.Debug("no found agent to start scan")
			qwm(ExitCodeError, "no found agent to start scan")
		}
		if agents.Total > 1 {
			klog.Debugf("[%d] agents found. Please specify it which one should be selected", agents.Total)
			qwm(ExitCodeError, "multiple agents found, please select one")
		}

		activeAgent := agents.ActiveAgents.First()
		klog.Debugf("agent [%s] found. Setting scan with agent", activeAgent.Label)
		scanparamsData.Agent = &client.ScanparamsItem{ID: activeAgent.ID}
	}

	klog.Printf("creating a new scanparams")
	scanparams, err := s.client.CreateScanparams(project.ID, scanparamsData)
	if err != nil {
		qwe(ExitCodeError, err, "failed to create scanparams")
	}
	scanData.ScanparamsID = scanparams.ID
	scanData.Custom = *scanparams.Custom

	klog.Printf("creating a new scan")
	return s.client.CreateNewScan(scanData)
}

func (s *Scan) parseCustomParams(custom client.Custom, scanner client.ScannerInfo, existParams *client.Scanparams) client.Custom {
	if len(scanner.Params) == 0 {
		klog.Debugf("the scanner tool [%s] does not allow custom parameter", scanner.DisplayName)
		qwm(ExitCodeError, "the scanner tool does not allow custom parameter")
	}

	params, err := s.cmd.Flags().GetStringSlice("params")
	if err != nil {
		klog.Debugf("failed to parse param flag: %v", err)
		qwm(ExitCodeError, "failed to parse params flag")
	}

	var requiredParamsLen = scanner.Params.RequiredParamsLen()

	if requiredParamsLen > len(params) {
		klog.Debugf("missing parameters for the scanner tool [%s]", scanner.DisplayName)
		qwm(ExitCodeError, "missing parameters for the scanner tool")
	}

	custom.Params = map[string]interface{}{}
	for _, v := range params {
		var keyValuePair = strings.SplitN(v, ":", 2)
		if len(keyValuePair) != 2 {
			klog.Debugf("invalid params flag: it should be key:value pairs: [%s]", keyValuePair)
			qwm(ExitCodeError, "invalid params flag, the flag is should be a pair of [key:value]")
		}

		var key = keyValuePair[0]
		var value = keyValuePair[1]

		// validate the given key parameter
		var fieldDetail = scanner.Params.Find(key)
		if fieldDetail == nil {
			klog.Debugf("params [%s] is not allowed by the scanner tool [%s], run `list scanners` command to display allowed params", key, scanner.DisplayName)
			qwm(ExitCodeError, "params key is not allowed by the scanner tool")
		}

		parsedValue, err := fieldDetail.Parse(value)
		if err != nil {
			klog.Debugf("failed to parse params key [%s] value [%s]: %v", key, value, err)
			qwm(ExitCodeError, "invalid value for custom params key")
		}

		custom = appendKeyToParamsMap(key, custom, parsedValue)
	}

	if existParams == nil || existParams.Custom == nil || existParams.Custom.Params == nil {
		return s.updateCustomParamsWithDefaultValue(scanner, custom)
	}

	for i, v := range existParams.Custom.Params {
		_, ok := custom.Params[i]
		if !ok {
			custom.Params[i] = v
		}
	}

	return s.updateCustomParamsWithDefaultValue(scanner, custom)
}

func (*Scan) updateCustomParamsWithDefaultValue(scanner client.ScannerInfo, custom client.Custom) client.Custom {
	for key := range scanner.Params {
		_, ok := custom.Params[key]
		if ok {
			continue
		}

		var fieldDetail = scanner.Params.Find(key)
		if fieldDetail == nil {
			klog.Debugf("params [%s] is not allowed by the scanner tool [%s], run `list scanners` command to display allowed params", key, scanner.DisplayName)
			qwm(ExitCodeError, "params key is not allowed by the scanner tool")
		}

		if fieldDetail.DefaultValue == "" {
			continue
		}

		parsedValue, err := fieldDetail.Parse(fieldDetail.DefaultValue)
		if err != nil {
			klog.Debugf("failed to parse default params key [%s] value [%s]: %v", key, fieldDetail.DefaultValue, err)
			qwm(ExitCodeError, "invalid value for custom params key")
		}

		klog.Debugf("the field [%s] is using a default value: [%v]", key, parsedValue)

		custom = appendKeyToParamsMap(key, custom, parsedValue)
	}
	return custom
}

func (s *Scan) startScanByProjectToolAndPR() (string, error) {
	rescanOnly, scanner, err := s.checkForRescanOnlyTool()
	if err != nil {
		return "", err
	}
	// Parse command line flags
	project, err := s.findORCreateProject()
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := s.cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	branch, err := s.cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}
	if branch == "" {
		return "", errors.New("missing branch field")
	}
	override, err := s.cmd.Flags().GetBool("override")
	if err != nil {
		return "", fmt.Errorf("failed to parse override flag: %w", err)
	}
	mergeTarget, err := s.cmd.Flags().GetString("merge-target")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	if mergeTarget == "" {
		return "", errors.New("missing merge-target field")
	}
	meta, err := s.cmd.Flags().GetString("meta")
	if err != nil {
		return "", fmt.Errorf("failed to parse meta flag: %w", err)
	}
	agent, err := s.cmd.Flags().GetString("agent")
	if err != nil {
		return "", fmt.Errorf("failed to parse agent flag: %w", err)
	}

	var agentID string
	if len(agent) > 0 {
		agentDetail, err := s.client.FindAgentByLabel(agent)
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

	var custom = client.Custom{Type: scanner.CustomType}
	if s.cmd.Flags().Changed("params") {
		custom = s.parseCustomParams(custom, *scanner, nil)
	}

	scan, err := s.client.FindScan(project.Name, params)
	if err == nil {
		opt := &client.ScanPROptions{
			From:               branch,
			To:                 mergeTarget,
			OverrideOldAnalyze: override,
			Custom:             custom,
		}
		eventID, err := s.client.RestartScanWithOption(scan.ID, opt)
		if err != nil {
			qwe(ExitCodeError, err, "could not start scan")
		}
		return eventID, nil
	}
	klog.Debugf("failed to get completed scans: %v, trying to get scanparams", err)

	sp, err := s.client.FindScanparams(project.Name, &client.ScanparamSearchParams{
		MetaData: meta,
		Branch:   branch,
		ToolID:   scanner.ID,
		Agent:    agent,
		Target:   mergeTarget,
		PR:       true,
		Limit:    1,
	})
	if err != nil {
		klog.Debugf("failed to get scanparams: %v, trying to create a new scan", err)
	}

	var scanData = func() *client.Scan {
		if sp != nil {
			return &client.Scan{
				Project:      project.Name,
				ToolID:       scanner.ID,
				ScanparamsID: sp.ID,
				Custom:       custom,
			}
		}

		if rescanOnly && !scanner.HasLabel(client.ScannerLabelAgent) && !s.cmd.Flags().Changed("params") {
			klog.Debugf("scanner tool %s is only allowing rescans", tool)
			qwm(ExitCodeError, "no scans found for given project, tool and PR configuration")
		}

		var scan = &client.Scan{
			MetaData: meta,
			Branch:   branch,
			Custom:   custom,
			Project:  project.Name,
			ToolID:   scanner.ID,
			PR: client.PRInfo{
				OK:     true,
				Target: mergeTarget,
			},
		}

		if rescanOnly && scanner.HasLabel(client.ScannerLabelAgent) {
			agents, err := s.client.ListActiveAgents(&client.AgentSearchParams{Label: agent})
			if err != nil {
				klog.Debugf("failed to get active agents: %v", err)
				qwm(ExitCodeError, "failed to get active agents")
			}
			if agents.Total == 0 {
				klog.Debug("no found agent to start scan")
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

	return s.client.CreateNewScan(scanData)
}

func (s *Scan) startScanByProjectToolAndPRNumber() (string, error) {
	rescanOnly, scanner, err := s.checkForRescanOnlyTool()
	if err != nil {
		return "", err
	}
	// Parse command line flags
	project, err := s.findORCreateProject()
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := s.cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	noDecoration, err := s.cmd.Flags().GetBool("no-decoration")
	if err != nil {
		return "", fmt.Errorf("failed to parse no-decoration flag: %w", err)
	}
	prNumber, err := s.cmd.Flags().GetString("pr-number")
	if err != nil {
		return "", fmt.Errorf("failed to get request number: %w", err)
	}
	override, err := s.cmd.Flags().GetBool("override")
	if err != nil {
		return "", fmt.Errorf("failed to parse override flag: %w", err)
	}
	meta, err := s.cmd.Flags().GetString("meta")
	if err != nil {
		return "", fmt.Errorf("failed to parse meta flag: %w", err)
	}
	agent, err := s.cmd.Flags().GetString("agent")
	if err != nil {
		return "", fmt.Errorf("failed to parse agent flag: %w", err)
	}

	var agentID string
	if len(agent) > 0 {
		agentDetail, err := s.client.FindAgentByLabel(agent)
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

	var custom = client.Custom{Type: scanner.CustomType}
	if s.cmd.Flags().Changed("params") {
		custom = s.parseCustomParams(custom, *scanner, nil)
	}
	scan, err := s.client.FindScan(project.Name, params)
	if err == nil {
		opt := &client.ScanPROptions{
			OverrideOldAnalyze: override,
			PRNumber:           prNumber,
			NoDecoration:       noDecoration,
			Custom:             custom,
		}

		eventID, err := s.client.RestartScanWithOption(scan.ID, opt)
		if err != nil {
			qwe(ExitCodeError, err, "could not start scan")
		}
		return eventID, nil
	}

	klog.Debugf("failed to get completed scans: %v, trying to get scanparams", err)

	sp, err := s.client.FindScanparams(project.Name, &client.ScanparamSearchParams{
		MetaData: meta,
		ToolID:   scanner.ID,
		Agent:    agent,
		PR:       true,
		Limit:    1,
	})
	if err != nil {
		klog.Debugf("failed to get scanparams: %v, trying to create a new scan", err)
	}

	var scanData = func() *client.Scan {
		if sp != nil {
			return &client.Scan{
				ScanparamsID: sp.ID,
				ToolID:       scanner.ID,
				Project:      project.Name,
				PR: client.PRInfo{
					OK:           true,
					PRNumber:     prNumber,
					NoDecoration: noDecoration,
				},
				Custom: custom,
			}
		}

		if rescanOnly && !scanner.HasLabel(client.ScannerLabelAgent) && !s.cmd.Flags().Changed("params") {
			klog.Debugf("scanner tool %s is only allowing rescans", tool)
			qwm(ExitCodeError, "no scans found for given project, tool and PR configuration")
		}

		var scan = &client.Scan{
			Project:  project.Name,
			ToolID:   scanner.ID,
			Custom:   custom,
			MetaData: meta,
			PR: client.PRInfo{
				OK:           true,
				PRNumber:     prNumber,
				NoDecoration: noDecoration,
			},
		}

		if rescanOnly && scanner.HasLabel(client.ScannerLabelAgent) {
			agents, err := s.client.ListActiveAgents(&client.AgentSearchParams{Label: agent})
			if err != nil {
				klog.Debugf("failed to get active agents: %v", err)
				qwm(ExitCodeError, "failed to get active agents")
			}
			if agents.Total == 0 {
				klog.Debug("no found agent to start scan")
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

	return s.client.CreateNewScan(scanData)
}

func (s *Scan) findScanIDByProjectToolAndForkScan() (string, error) {
	rescanOnly, scanner, err := s.checkForRescanOnlyTool()
	if err != nil {
		return "", err
	}
	// Parse command line flags
	project, err := s.findORCreateProject()
	if err != nil {
		return "", fmt.Errorf("failed to parse project flag: %w", err)
	}
	tool, err := s.cmd.Flags().GetString("tool")
	if err != nil {
		return "", fmt.Errorf("failed to parse tool flag: %w", err)
	}
	branch, err := s.cmd.Flags().GetString("branch")
	if err != nil {
		return "", fmt.Errorf("failed to parse branch flag: %w", err)
	}
	if branch == "" {
		return "", errors.New("missing branch field")
	}
	meta, err := s.cmd.Flags().GetString("meta")
	if err != nil {
		return "", fmt.Errorf("failed to parse meta flag: %w", err)
	}
	forkScan, err := s.cmd.Flags().GetBool("fork-scan")
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

	scan, err := s.client.FindScan(project.Name, params)
	if err == nil {
		eventID, err := s.client.RestartScanByScanID(scan.ID)
		if err != nil {
			qwe(1, err, "could not start scan")
		}
		return eventID, nil
	} else {
		klog.Debugf("failed to get completed scans: %v, trying to get scanparams", err)
	}

	sp, err := s.client.FindScanparams(project.Name, &client.ScanparamSearchParams{
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
		var custom = client.Custom{Type: scanner.CustomType}
		if s.cmd.Flags().Changed("params") {
			custom = s.parseCustomParams(custom, *scanner, sp)
		}

		var scanPayload = &client.Scan{
			Branch:   branch,
			MetaData: meta,
			Project:  project.Name,
			ForkScan: forkScan,
			ToolID:   scanner.ID,
			Custom:   custom,
		}

		if sp != nil {
			scanPayload.ScanparamsID = sp.ID
			return scanPayload
		}

		if rescanOnly {
			klog.Debugf("scanner tool %s is only allowing rescans", tool)
			klog.Fatal("no scans found for given project, tool and PR configuration")
		}

		return scanPayload
	}()

	return s.client.CreateNewScan(scanData)
}

func (s *Scan) checkForRescanOnlyTool() (bool, *client.ScannerInfo, error) {
	klog.Debug("checking for rescan only tools")
	name, err := s.cmd.Flags().GetString("tool")
	if err != nil || name == "" {
		return false, nil, errors.New("missing require tool flag")
	}
	scanners, err := s.client.ListActiveScanners(&client.ScannersSearchParams{Name: name, Limit: 1})
	if err != nil {
		return false, nil, fmt.Errorf("failed to get active scanners: %w", err)
	}
	if scanners.Total == 0 {
		return false, nil, fmt.Errorf("invalid or inactive scanner tool name: %s", name)
	}
	if scanners.Total > 1 {
		return false, nil, fmt.Errorf("multiple scanners found for tool: %s", name)
	}

	scanner := scanners.ActiveScanners.First()
	if scanner.HasLabel(client.ScannerLabelCreatableOnTool) {
		return false, scanner, nil
	}

	isForkScan, err := s.cmd.Flags().GetBool("fork-scan")
	if err != nil {
		return false, nil, fmt.Errorf("failed to get fork-scan flag: %w", err)
	}

	for _, label := range scanner.Labels {
		if client.IsRescanOnlyLabel(label, isForkScan) {
			return true, scanner, nil
		}
	}

	return false, scanner, nil
}

func (s *Scan) findORCreateProject() (*client.Project, error) {
	if !s.cmd.Flags().Changed("repo-id") && !s.cmd.Flags().Changed("project") {
		return nil, errors.New("missing a required flag(repo or project) to get project detail")
	}

	repo, err := s.cmd.Flags().GetString("repo-id")
	if err != nil {
		return nil, fmt.Errorf("failed to get repo flag: %w", err)
	}
	var name string
	if repo == "" {
		project, err := getSanitizedFlagStr(s.cmd, "project")
		if err != nil {
			return nil, fmt.Errorf("failed to get project flag: %w", err)
		}
		name = project
	}

	projects, err := s.client.ListProjects(name, repo)
	if err != nil {
		return nil, err
	}

	if len(projects) == 1 {
		return &projects[0], nil
	}

	if len(projects) > 1 {
		return nil, errors.New("multiple projects found for given parameters")
	}

	createProject, err := s.cmd.Flags().GetBool("create-project")
	if err != nil {
		return nil, fmt.Errorf("failed to get create-project flag: %w", err)
	}

	if !createProject {
		return nil, errors.New("no projects were found according to the given parameters")
	}

	if !s.cmd.Flags().Changed("repo-id") {
		return nil, errors.New("missing a required repo flag to create project")
	}

	var p = Project{
		cmd:    s.cmd,
		client: s.client,
	}

	var project = p.createProject(repo, false)

	if !p.cmd.Flags().Changed("product-name") {
		return project, nil
	}
	var pr = Product{
		cmd:    s.cmd,
		client: s.client,
	}

	productName, err := p.cmd.Flags().GetString("product-name")
	if err != nil {
		qwe(ExitCodeError, err, "failed to parse the product-name flag: %v")
	}
	var parsedProjects = []client.Project{*project}
	product, created := pr.createProduct(productName, parsedProjects)
	if created {
		klog.Println("product created successfully")
		return project, nil
	}

	pr.updateProduct(product, parsedProjects)
	qwm(ExitCodeSuccess, "the project assigned to the product")
	return project, nil
}

func getScanMode(cmd *cobra.Command) uint {
	// Check scan method
	byImportFile := cmd.Flag("file").Changed
	byTool := cmd.Flag("tool").Changed
	byScanID := cmd.Flag("scan-id").Changed
	byProject := cmd.Flag("project").Changed
	byBranch := cmd.Flag("merge-target").Changed
	byForkScan := cmd.Flag("fork-scan").Changed
	byMerge := cmd.Flag("branch").Changed
	byPRNumber := cmd.Flag("pr-number").Changed
	byImage := cmd.Flag("image").Changed
	byRepo := cmd.Flag("repo-id").Changed
	byProjectORRepo := byProject || byRepo
	byPR := byBranch && byMerge

	byProjectAndTool := byProjectORRepo && byTool && !byPR //
	byProjectAndToolAndFile := byProjectAndTool && byImportFile
	byProjectAndToolAndForkScan := byProjectORRepo && byTool && byForkScan && !byPR
	byProjectAndToolAndPullRequestNumber := byProjectORRepo && byTool && byPRNumber && !byImportFile
	byProjectAndToolAndPullRequest := byProjectORRepo && byTool && byPR && !byImportFile //

	mode := func() uint {
		// sorted by priority
		switch true {
		case byImage:
			return modeByImage
		case byProjectAndToolAndFile:
			return modeByFileImport
		case byScanID:
			return modeByScanID
		case byProjectAndToolAndPullRequest:
			return modeByProjectToolAndPR
		case byProjectAndToolAndPullRequestNumber:
			return modeByProjectToolAndPRNumber
		case byProjectAndToolAndForkScan:
			return modeByProjectToolAndForkScan
		case byProjectAndTool:
			return modeByProjectTool
		default:
			return modeByScanID
		}
	}()
	return mode
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

func checkRelease(scan *client.ScanDetail, cmd *cobra.Command) error {
	c, err := client.New()
	if err != nil {
		return err
	}

	rs, err := c.ReleaseStatus(scan.Project, scan.Branch)
	if err != nil {
		return fmt.Errorf("failed to get release status: %w", err)
	}

	return isScanReleaseFailed(scan, rs, cmd)
}

func isScanReleaseFailed(scan *client.ScanDetail, release *client.ReleaseStatus, cmd *cobra.Command) error {
	breakByScannerType, err := cmd.Flags().GetBool("break-by-scanner-type")
	if err != nil {
		return err
	}

	const statusFail = "fail"

	if release.Status != statusFail {
		return nil
	}

	var failedScans = make(map[string]string, 0)

	if release.SAST.Status == statusFail {
		failedScans["SAST"] = scan.ID
	}
	if release.DAST.Status == statusFail {
		failedScans["DAST"] = scan.ID
	}
	if release.PENTEST.Status == statusFail {
		failedScans["PENTEST"] = scan.ID
	}
	if release.IAST.Status == statusFail {
		failedScans["IAST"] = scan.ID
	}
	if release.SCA.Status == statusFail {
		failedScans["SCA"] = scan.ID
	}
	if release.CS.Status == statusFail {
		failedScans["CS"] = scan.ID
	}
	if release.IAC.Status == statusFail {
		failedScans["IAC"] = scan.ID
	}

	if breakByScannerType {
		scannerType := strings.ToUpper(scan.ScannerType)

		if _, ok := failedScans[scannerType]; !ok {
			// This means, this scanner type isn't failed. So, we can ignore it because we are breaking by scanner type
			return nil
		}
		return fmt.Errorf("project does not pass release criteria due to [%s] failure", scannerType)
	}

	if verbose {
		c, err := client.New()
		if err != nil {
			qwe(ExitCodeError, err, "could not initialize Kondukto client")
		}

		for toolType, scanID := range failedScans {
			fmt.Println()
			fmt.Println("-----------------------------------------------------------------")
			fmt.Printf("[!] project does not pass release criteria due to [%s] failure\n", toolType)
			scan, err := c.FindScanByID(scanID)
			if err != nil {
				qwe(ExitCodeError, err, "failed to fetch scan summary")
			}

			printScanSummary(scan)
			fmt.Println("-----------------------------------------------------------------")
		}
	}

	var failedToolTypes []string
	for toolType := range failedScans {
		failedToolTypes = append(failedToolTypes, toolType)
	}

	returnMSG := fmt.Sprintf("project does not pass release criteria due to [%s] failure", strings.Join(failedToolTypes, ", "))

	return errors.New(returnMSG)
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
			if event.Status == eventStatusFinished {
				klog.Println("scan finished successfully")
				scan, err := c.FindScanByID(event.ScanID)
				if err != nil {
					qwe(ExitCodeError, err, "failed to fetch scan summary")
				}

				// Printing scan results
				printScanSummary(scan)

				if err = passTests(scan, cmd); err != nil {
					qwe(ExitCodeError, err, "scan could not pass security tests")
				} else if err = checkRelease(scan, cmd); err != nil {
					qwe(ExitCodeError, err, "scan failed to pass release criteria")
				}
				qwm(ExitCodeSuccess, "scan passed security tests successfully")
			}
		case eventActive:
			if duration != 0 && time.Now().Sub(start) > duration {
				qwm(ExitCodeSuccess, "scan duration exceeds timeout, it will continue running async in the background")
			}
			if event.Status != lastStatus {
				klog.Printf("scan status is [%s]", event.StatusText)
				lastStatus = event.Status
				// Get new scans scan id
			} else {
				klog.Debugf("event status is [%s]", event.StatusText)
			}
			time.Sleep(10 * time.Second)
		default:
			qwm(ExitCodeError, fmt.Sprintf("unknown event status: %d", event.Active))
		}
	}
}

// appendKeyToParamsMap appends the key to the custom params map
// generates a nested map object if the key is contians a dot
// for example: if key:"image.tag" and value:"latest" will generate a map object {"image": {"tag": "value"}}
func appendKeyToParamsMap(key string, custom client.Custom, parsedValue interface{}) client.Custom {
	var splitted = strings.Split(key, ".")
	switch len(splitted) {
	case 1:
		custom.Params[key] = parsedValue
	case 2:
		key0 := splitted[0]
		key1 := splitted[1]
		if _, ok := custom.Params[key0]; !ok {
			custom.Params[key0] = map[string]interface{}{}
		}

		key0map := custom.Params[key0].(map[string]interface{})
		if _, ok := key0map[key1]; ok {
			klog.Debugf("params keys are not unique [%s]", key)
			qwm(ExitCodeError, "params keys are not unique")
		}

		key0map[key1] = parsedValue
		custom.Params[key0] = key0map

	case 3:
		key0 := splitted[0]
		key1 := splitted[1]
		key2 := splitted[2]
		if _, ok := custom.Params[key0]; !ok {
			custom.Params[key0] = map[string]interface{}{}
		}

		key0map := custom.Params[key0].(map[string]interface{})
		if _, ok := key0map[key1]; !ok {
			key0map[key1] = map[string]interface{}{}
		}

		key1map := key0map[key1].(map[string]interface{})
		if _, ok := key1map[key2]; ok {
			klog.Debugf("params keys are not unique [%s]", key)
			qwm(ExitCodeError, "params keys are not unique")
		}
		key1map[key2] = parsedValue
		key0map[key1] = key1map
		custom.Params[key0] = key0map

	default:
		klog.Debugf("unsupportted key: [%s]", key)
		qwm(ExitCodeError, "unsupportted key, key can only contain one or two dots")
	}
	return custom
}
