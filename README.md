<p align="center"><a href="https://kondukto.io" target="_blank" rel="noopener noreferrer"><img width="200" src="https://kondukto.io/logo.png" alt="Kondukto logo"></a></p>

# KDT
KDT is a command line client for [Kondukto](https://kondukto.io) written in [Go](https://golang.org). It interacts with Kondukto engine through public API. 

With KDT, you can list projects and their scans in **Kondukto**, and restart a scan with a specific application security tool. KDT is also easy to use in CI/CD pipelines to trigger scans and break releases if a scan fails or scan results don't met specified release criteria. 

### What is Kondukto?
[Kondukto](https://kondukto.io) is an Application Security Testing Orchestration platform that helps you centralize and automate your entire AppSec related vulnerability management process. Providing an interface where security health of applications can be continuously monitored, and a command line interface where your AppSec operations can be integrated into DevOps pipelines, Kondukto lets you manage your AppSec processes automatically with ease.

## Installation
You can install the CLI with a `curl` utility script or by downloading the pre-compiled binary from the Github release page.
Once installed youl'll get the `kdt-cli` command and `kdt` alias.

Utility script with `curl`:
```shell
$ curl -sSL https://cli.kondukto.io | sudo sh
```

Non-root with curl:
```shell
$ curl -sSL https://cli.kondukto.io | sh
```

### Windows 
To install the kdt-cli on Windows go to [Releases](https://github.com/kondukto-io/kdt/releases) and download the latest kdt-cli.exe.


Or you can also simply run the following if you have an existing [Go](https://golang.org) environment:
```shell
go get github.com/kondukto-io/kdt
```

If you want to build it yourself, clone the source files using Github, change into the `kdt` directory and run:
```shell
git clone https://github.com/kondukto-io/kdt.git
cd kdt
go install
```

## Configuration
KDT needs Kondukto host and an API token for authentication. API tokens can be created under Integrations/API Tokens menu.

You can provide configuration by:

##### 1) Setting environment variables: 

*(example is for BASH shell)*
```shell
$ export KONDUKTO_HOST=http://localhost:8080
$ export KONDUKTO_TOKEN=WmQ2eHFDRzE3elplN0ZRbUVsRDd3VnpUSHk0TmF6Uko5OGlyQ1JvR2JOOXhoWEFtY2ZrcDJZUGtrb2tV
```
It is always better to set environment variables in shell profile files(`~/.bashrc`, `~/.zshrc`, `~/.profile` etc.)
##### 2) Providing a configuration file.

Default path for config file is `$HOME/.kdt.yaml`. Another file can be provided with `--config` command line flag.
```
// $HOME/.kdt.yaml 
host: http://localhost:8088
token: WmQ2eHFDRzE3elplN0ZRbUVsRDd3VnpUSHk0TmF6Uko5OGlyQ1JvR2JOOXhoWEFtY2ZrcDJZUGtrb2tV
```

##### 3) Using command line flags
```
kdt list projects --host http://localhost:8088 --token WmQ2eHFDRzE3elplN0ZRbUVsRDd3VnpUSHk0TmF6Uko5OGlyQ1JvR2JOOXhoWEFtY2ZrcDJZUGtrb2tV
```

### Health Checks

Regular health checks are critical in ensuring uninterrupted communication between KDT and the Kondukto service. This section provides the necessary commands to perform these checks.

- **Verify KDT Connection to Kondukto Service**

  This command allows you to check whether KDT is successfully connected to the Kondukto service.

    ```shell
    $ kdt ping
    ```

- **Validate API Token**

  This command enables you to confirm that your API token is valid.

    ```shell
    $ kdt ping -a
    ```

## Command Overview

This section provides an overview of key KDT commands, including instructions on how to list projects, list project scans, check ALM project availability, restart a scan, and import scan results.

### Listing Projects

To retrieve a list of all projects, utilize the following command:

```shell
$ kdt list projects
```

### Listing Scans for a Specific Project

To list all scans associated with a specific project, use the command below. Remember to replace "ExampleProject" with the name of your project:

```shell
$ kdt list scans -p ExampleProject
```

### Checking ALM Project Availability

The command below checks the availability of an ALM project. The `$ALM_TOOL` placeholder should be replaced with the name of your ALM tool, and the `$PROJECT_ID` placeholder with your project ID.

```shell
$ kdt project available -a $ALM_TOOL -r $PROJECT_ID
```

The command will return an exit code of 0 if the project is available, and an exit code of -1 (255) if the project is not available.

### Restarting a Scan

There are two options to restart a scan:

1. Using the scan ID:

    ```shell
    $ kdt scan -s 5da6cafa5ab6e436faf643dc
    ```

2. Using the project and tool names:

    ```shell
    $ kdt scan -p ExampleProject -t ExampleTool
    ```

### Importing Scan Results

To import scan results as a file, use the following command:

```shell
$ kdt scan -p ExampleProject -t ExampleTool -b master
```

## Command Line Flags

KDT offers a range of useful flags to streamline the management of scans. These include both global flags, applicable to all KDT commands, and command-specific flags.

### Global Flags

The following flags can be applied across all KDT commands:

- `--host`: Defines the HTTP address of the Kondukto server, including the port.

- `--token`: Specifies the API token generated by Kondukto.

- `--config`: Points to the configuration file to use, superseding the default one (`$HOME/.kdt.yaml`).

- `--async`: Initiates an asynchronous scan that won't block the process while waiting for the scan to complete. KDT will exit gracefully once the scan has successfully started.

- `--insecure`: Bypasses the client's verification of the server's certificates and host name. Please note that this mode exposes TLS to potential man-in-the-middle attacks. It is not recommended unless you fully understand the potential risks.

- `-v` or `--verbose`: Enables verbose output, providing more detailed logs. This flag is particularly useful for debugging purposes.

### Scan Command Flags

The following flags are exclusive to the scan commands:

- `-p` or `--project`: Specifies the project name or ID.

- `-t` or `--tool`: Specifies the tool name.

- `-s` or `--scan-id`: Indicates the scan ID.

- `-b` or `--branch`: Designates the branch to be scanned.

Please note that these flags are only applicable for scan commands.

#### Additional Note
If these flags are used in conjunction with the `-v` (verbose) flag, more detailed information about the scan will be provided.

### Release Command Flags

The following flags are specific to the release commands:

- `-p` or `--project`: Specifies the project name or ID.

- `--cs`: Processes CS (Code Security) criteria status.

- `--dast`: Processes DAST (Dynamic Application Security Testing) criteria status.

- `--iac`: Processes IAC (Infrastructure as Code) criteria status.

- `--iast`: Processes IAST (Interactive Application Security Testing) criteria status.

- `--pentest`: Processes Penetration Testing criteria status.

- `--sast`: Processes SAST (Static Application Security Testing) criteria status.

- `--sca`: Processes SCA (Software Composition Analysis) criteria status.

Please note that these flags are only valid for release commands.

#### Additional Note
If these flags are used in conjunction with the `-v` (verbose) flag, more detailed information about the criteria status of the release will be provided.

### Threshold Flags

The following flags represent thresholds for the maximum number of vulnerabilities, of a specified severity, to be ignored. Should these thresholds be exceeded, KDT will terminate with a non-zero status code.

- `--threshold-crit`: Defines the threshold for critical severity vulnerabilities.

- `--threshold-high`: Sets the threshold for high severity vulnerabilities.

- `--threshold-med`: Establishes the threshold for medium severity vulnerabilities.

- `--threshold-low`: Determines the threshold for low severity vulnerabilities.

- `--threshold-risk`: Sets the risk threshold for failing tests if the scan results in a higher risk score than the previous scan's risk score. This flag is useful for maintaining a project's security level. If used with every scan in DevOps pipelines, it ensures the project's vulnerability does not increase.

Please note that the risk threshold only considers the last two scans performed with the same tool. If the project has not been scanned with the tool, KDT will fail as it cannot compare risk scores. Also, these threshold flags do not function with the `--async` flag since KDT will exit when the scan begins, and thus cannot check scan results.

#### Example Usage:

The following command scans the project "SampleProject" with the tool "SampleTool", setting thresholds for critical vulnerabilities at 3, high vulnerabilities at 10, and considering the risk:

```shell
$ kdt scan -p SampleProject -t SampleTool --threshold-crit 3 --threshold-high 10 --threshold-risk
```

## Supported scanners (tools)
KDT supports all scanners enabled in Kondukto server, to see the list simply run `kdt list scanners`.

Example Usage:

```
./kdt --config kondukto.yaml list scanners
Name       ID                          Type    Trigger     Labels
----       --                          ----    -------     ------
gosec      60eec8a83e9e5e6e2ae52d06    sast    new scan    docker,kdt
semgrep    60eec8a53e9e5e6e2ae52d05    sast    rescan      template,docker,kdt
```

### Tool list (full)
```
checkmarx
checkmarxsca
owaspzap
webinspect
netsparker
appspider
bandit
findsecbugs
dependencycheck
fortify
gosec
brakeman
securitycodescan
trivy
hclappscan
owaspzapheadless
nancy
semgrep
veracode
burpsuite
burpsuiteenterprise
```

## Advanced Usage Examples

KDT can be utilized in various ways within your pipeline. The following example demonstrates an advanced use case:

```shell
$ kdt --config kondukto-config.yml \
    --insecure \
    scan \
    --project SampleProject \
    --tool fortify \
    --file results.fpr \
    --branch develop \
    --threshold-crit 0 \
    --threshold-high 0
```

In this command:

- `--config`: Specifies the Kondukto configuration file in yaml format.

- `--insecure`: Indicates not to verify SSL certificates.

- `scan`: Initiates a scan.

- `--project`: Defines the name of the application on the Kondukto server.

- `--tool`: Specifies the AST (Application Security Testing) tool to be used, in this case, 'fortify'.

- `--file`: Determines the results filename. When this parameter is given, the scan will not be initiated, and only the results file (results.fpr) will be analyzed.

- `--branch`: Specifies the branch name.

- `--threshold-crit`: Sets the critical severity threshold value to "break the build" in the pipeline. When this parameter is provided, the entered security criteria will be overwritten.

---

```shell
$ kdt --config kondukto-config.yml \
    scan \
    --project SampleProject \
    --tool trivy \
    --image ubuntu@256:ab02134176aecfe0c0974ab4d3db43ca91eb6483a6b7fe6556b480489edd04a1 \
    --branch develop
```

In this command:

- `--config`: Specifies the Kondukto configuration file in yaml format.

- `scan`: Initiates a scan.

- `--project`: Defines the name of the application on the Kondukto server.

- `--tool`: Specifies the AST (Application Security Testing) tool to be used, in this case, 'trivy'.

- `--image`: Identifies the image to be scanned. The image name can be given with the digest or with the tag name (e.g., ubuntu:latest).

- `--branch`: Specifies the branch name.

---

The following example illustrates how to create a new project using KDT:

```shell
$ kdt --config kondukto-config.yml \
    create \
    project \ 
    --repo-id https://github.com/kondukto-io/kdt \
    --labels GDPR,Internal \
    --alm-tool github
```

In this command:

- `--config`: Specifies the Kondukto configuration file in yaml format.

- `create`: Acts as the base command for the create operation.

- `project`: Acts as a subcommand to create a new project.

- `--repo-id`: Specifies the project repository URL or ALM ID.

- `--labels`: Associates the project with a list of labels.

- `--alm-tool`: Specifies the ALM (Application Lifecycle Management) tool. This is required if more than one ALM is enabled in Kondukto.

Additional flags that can be set include:

- `--team`: Specifies a team name. By default, the team name is 'default team'.

- `--force-create`: Creates a project with a suffix `-` if there is another project with the same name.

- `--overwrite`: Overwrites the project name, eliminating the need to add a `-` suffix.

This command creates a project on Kondukto that matches the name in your ALM tool. If a project with the same name already exists, the command will print an error message and exit with a status code. You can pass the `--force-create` flag to create a project with a suffix `-`, or pass the `--overwrite` flag to overwrite the project name.

---

## Contributing to KDT

Contributions to the KDT project are highly appreciated. Whether you're reporting issues, suggesting new features, or directly helping with development, your input is valuable.

Here's how you can contribute:

- Report issues or suggest new features by creating a new issue in the repository.

- Contribute directly to the codebase by forking the repository and creating pull requests.

Before submitting your pull requests, please adhere to the following guidelines:

- Create and name your branches according to the [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/) methodology.
  - For new features: `feature/example-feature-branch`
  - For bug fixes: `bugfix/example-bugfix-branch`

- Ensure that your code is properly documented, following idiomatic [Go](https://golang.org) practices. Exported functions should always be commented.

- Write detailed PR descriptions and comments. This helps maintainers understand your changes and speeds up the review process.

Thank you for helping to improve KDT!
