<p align="center"><a href="https://kondukto.io" target="_blank" rel="noopener noreferrer"><img width="200" src="https://kondukto.io/logo.png" alt="Kondukto logo"></a></p>

# KDT

KDT is an open-source command line interface for [Kondukto](https://kondukto.io), an Application Security Posture Management (ASPM) platform. Written in [Go](https://golang.org), KDT interacts with the Kondukto engine through its public API.

With KDT, you can list projects and their scans in **Kondukto**, trigger scans with specific application security tools, import scan results, manage SBOM files, and break releases if scan results don't meet specified release criteria. KDT is designed to seamlessly integrate with CI/CD pipelines for automated DevSecOps workflows.

## What is Kondukto?

[Kondukto](https://kondukto.io) is an Application Security Posture Management (ASPM) platform that helps you centralize and automate your entire AppSec vulnerability management process. It provides:
- Centralized security health monitoring for applications
- DevSecOps pipeline integration
- Automated AppSec workflow orchestration
- Release criteria enforcement
- SBOM management

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Global Flags](#global-flags)
- [Commands](#commands)
  - [Health Checks](#health-checks)
  - [Scan Command](#scan-command)
  - [Release Command](#release-command)
  - [List Commands](#list-commands)
  - [Create Commands](#create-commands)
  - [SBOM Commands](#sbom-commands)
  - [Endpoint Commands](#endpoint-commands)
  - [Status Command](#status-command)
  - [Project Commands](#project-commands)
- [Advanced Usage Examples](#advanced-usage-examples)
- [Contributing](#contributing)

## Installation

You can install KDT using several methods:

### Using curl (Linux/macOS)

**With sudo (installs system-wide):**
```shell
curl -sSL https://cli.kondukto.io | sudo sh
```

**Without sudo (user installation):**
```shell
curl -sSL https://cli.kondukto.io | sh
```

### Windows

Download the latest `kdt-cli.exe` from [Releases](https://github.com/kondukto-io/kdt/releases).

### Using Go

If you have a Go environment:
```shell
go get github.com/kondukto-io/kdt
```

### Building from Source

```shell
git clone https://github.com/kondukto-io/kdt.git
cd kdt
go build . -o kdt
```

Or simply:
```shell
make all
```

## Configuration

KDT requires a Kondukto host URL and an API token for authentication. API tokens can be created under **Integrations > API Tokens** in the Kondukto UI.

### Configuration Methods

#### 1. Environment Variables

```shell
export KONDUKTO_HOST=https://your-kondukto-instance.com
export KONDUKTO_TOKEN=your_api_token_here
```

For persistence, add these to your shell profile (`~/.bashrc`, `~/.zshrc`, `~/.profile`).

#### 2. Configuration File

Default location: `$HOME/.kdt.yaml`

```yaml
host: https://your-kondukto-instance.com
token: your_api_token_here
insecure: false
verbose: false
```

You can specify a custom config file:
```shell
kdt --config=/path/to/config.yaml list projects
```

#### 3. Command Line Flags

```shell
kdt --host https://your-kondukto-instance.com --token your_api_token list projects
```

**Configuration Priority:** Command line flags > Environment variables > Configuration file

## Global Flags

These flags can be used with any KDT command:

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to configuration file | `$HOME/.kdt.yaml` |
| `--host` | Kondukto server host URL | - |
| `--token` | Kondukto API token | - |
| `--insecure` | Skip TLS certificate verification (not recommended for production) | `false` |
| `-v, --verbose` | Enable verbose logging for debugging | `false` |
| `--exit-code` | Override the exit code | `0` |

**Example:**
```shell
kdt --config=prod-config.yaml --verbose scan -p MyProject -t semgrep -b main
```

## Commands

### Health Checks

#### Verify Connection
Test connectivity to Kondukto service:
```shell
kdt ping
```

#### Validate API Token
Verify that your API token is valid:
```shell
kdt ping -a
```

### Scan Command

The `scan` command is the primary command for triggering security scans and importing scan results.

#### Scan Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--async` | - | Run scan asynchronously (non-blocking) | `false` |
| `--project` | `-p` | Project name or ID | - |
| `--tool` | `-t` | Scanner tool name | - |
| `--scan-id` | `-s` | Scan ID to restart | - |
| `--branch` | `-b` | Branch name | - |
| `--file` | `-f` | Scan result file to import | - |
| `--image` | `-I` | Container image to scan | - |
| `--agent` | `-a` | Agent name for agent-based scanners | - |
| `--meta` | `-m` | Metadata | - |
| `--scan-tag` | - | Tag for the scan | - |
| `--env` | - | Environment: `production`, `staging`, `develop`, `feature` | - |
| `--timeout` | - | Minutes to wait for scan completion (0 = no timeout) | `0` |
| `--release-timeout` | - | Minutes to wait for release criteria check | `5` |

#### Pull Request Scanning Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--merge-target` | `-M` | Target branch for PR scans |
| `--pr-number` | - | PR number for decoration |
| `--pr-decoration-scanner-types` | - | Scanner types for PR decoration (e.g., `all`, `sast`, `dast`, `sca`) |
| `--override` | - | Override old analysis results for PR scans |
| `--no-decoration` | - | Disable PR decoration (deprecated) |

> **Note:** For pull request scans, the target branch (specified with `--merge-target`) must be scanned at least once before triggering PR scans. This baseline scan is required for comparison.

#### Fork Scanning Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--fork-scan` | `-B` | Enable fork scan based on default branch |
| `--fork-source` | - | Source branch for fork scans |
| `--override-fork-source` | - | Override project's fork source branch |

> **Note:** For fork scans, the source branch (specified with `--fork-source` or the project's default branch) must be scanned at least once before triggering fork scans. This establishes the baseline for comparison.

#### Threshold Flags

Break the build if vulnerabilities exceed thresholds:

| Flag | Description |
|------|-------------|
| `--threshold-crit` | Maximum critical vulnerabilities |
| `--threshold-high` | Maximum high vulnerabilities |
| `--threshold-med` | Maximum medium vulnerabilities |
| `--threshold-low` | Maximum low vulnerabilities |
| `--threshold-risk` | Fail if risk score increases |
| `--break-by-scanner-type` | Only break for specific scanner type |

#### Project Creation Flags

Automatically create projects during scan:

| Flag | Short | Description |
|------|-------|-------------|
| `--create-project` | - | Create project if not found |
| `--project-name` | - | Name for new project |
| `--repo-id` | `-r` | Repository URL or ID |
| `--alm-tool` | - | ALM tool name (e.g., `github`, `gitlab`) |
| `--team` | `-T` | Team name |
| `--labels` | `-l` | Comma-separated labels |
| `--product-name` | `-P` | Product name |
| `--default-branch` | - | Default branch | `main` |
| `--disable-clone` | - | Disable repository cloning |
| `--criticality-level` | - | Business criticality: 4=Major, 3=High, 2=Medium, 1=Low, 0=None, -1=Auto |
| `--feature-branch-retention` | - | Days to retain feature branches |
| `--feature-branch-infinite-retention` | - | Never delete feature branches |
| `--scope-include-empty` | - | Include vulnerabilities with no path |
| `--scope-included-paths` | - | Comma-separated paths for mono-repo scoping |
| `--scope-included-files` | - | Comma-separated file names for scoping |

#### Custom Parameters

| Flag | Description |
|------|-------------|
| `--params` | Custom scanner parameters (format: `key:value`) |
| `--incremental-scan` | `-i` | Enable incremental scanning (Semgrep only) |

#### Scan Examples

**1. Restart an existing scan by scan ID:**
```shell
kdt scan -s 5da6cafa5ab6e436faf643dc
```

**2. Trigger scan with project and tool:**
```shell
kdt scan -p MyProject -t semgrep -b main
```

**3. Import scan results from file:**
```shell
kdt scan -p MyProject -t checkmarx -b develop -f results.xml
```

**4. Scan with thresholds (break build):**
```shell
kdt scan -p MyProject -t trivy -b main \
  --threshold-crit 0 \
  --threshold-high 5 \
  --threshold-med 10
```

**5. Async scan (non-blocking):**
```shell
kdt scan -p MyProject -t gosec -b main --async
```

**6. Container image scan:**
```shell
kdt scan -p MyProject -t trivy \
  --image myapp:latest \
  -b main
```

**7. Pull request scan:**
```shell
kdt scan -p MyProject -t semgrep \
  -b feature/new-feature \
  -M main \
  --pr-number 123
```

**8. Fork scan (feature branch vs default):**
```shell
kdt scan -p MyProject -t semgrep \
  -b feature/test \
  --fork-scan \
  --env feature
```

**9. Create project and scan:**
```shell
kdt scan -p NewProject -t semgrep -b main \
  --create-project \
  --repo-id https://github.com/org/repo \
  --alm-tool github \
  --team security
```

**10. Custom parameters:**
```shell
kdt scan -p MyProject -t semgrep -b develop \
  --params=ruleset_type:2 \
  --params=ruleset_options.ruleset:/custom/rules/
```

**11. Risk threshold (prevent regression):**
```shell
kdt scan -p MyProject -t sonarqube -b main --threshold-risk
```

**12. Incremental scan (Semgrep):**
```shell
kdt scan -p MyProject -t semgrep -b main \
  -f semgrep-results.json \
  --incremental-scan
```

### Release Command

Check if a project passes release criteria.

#### Release Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--project` | `-p` | Project name or ID (required) |
| `--branch` | `-b` | Branch name (default: project's default branch) |
| `--timeout` | - | Minutes to wait for criteria check | `5` |
| `--sast` | - | Check SAST criteria |
| `--dast` | - | Check DAST criteria |
| `--sca` | - | Check SCA criteria |
| `--iac` | - | Check IaC criteria |
| `--cs` | - | Check Code Security criteria |
| `--iast` | - | Check IAST criteria |
| `--pentest` | - | Check Penetration Testing criteria |
| `--mast` | - | Check MAST criteria |
| `--sbom` | - | Check SBOM criteria |

#### Release Examples

**1. Check all release criteria:**
```shell
kdt release -p MyProject -b main
```

**2. Check specific criteria only:**
```shell
kdt release -p MyProject -b main --sast --sca
```

**3. With verbose output:**
```shell
kdt -v release -p MyProject -b main --sast --dast
```

**4. With custom timeout:**
```shell
kdt release -p MyProject -b main --timeout 10
```

### List Commands

#### List Projects

```shell
kdt list projects
```

#### List Scans

**Flags:**
- `-p, --project`: Project name or ID (required)

**Example:**
```shell
kdt list scans -p MyProject
```

#### List Scanners

View all available scanners:
```shell
kdt list scanners
```

**Example output:**
```
Name       ID                          Type    Trigger     Labels
----       --                          ----    -------     ------
gosec      60eec8a83e9e5e6e2ae52d06    sast    new scan    docker,kdt
semgrep    60eec8a53e9e5e6e2ae52d05    sast    rescan      template,docker,kdt
trivy      60eec8a73e9e5e6e2ae52d07    sca     new scan    docker,kdt,container
```

#### List Agents

```shell
kdt list agents
```

#### List Products

```shell
kdt list products
```

### Create Commands

#### Create Project

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--repo-id` | `-r` | Repository URL or ID (required) |
| `--project-name` | - | Project name |
| `--alm-tool` | `-a` | ALM tool name |
| `--team` | `-t` | Team name |
| `--labels` | `-l` | Comma-separated labels |
| `--product-name` | `-P` | Product name |
| `--force-create` | - | Create with suffix if name exists |
| `--overwrite` | `-w` | Overwrite existing project |
| `--default-branch` | - | Default branch | `main` |
| `--disable-clone` | - | Disable repository cloning |
| `--fork-source` | - | Source branch for feature branches |
| `--criticality-level` | - | Business criticality (0-4, -1=Auto) |
| `--feature-branch-retention` | - | Days to retain feature branches |
| `--feature-branch-infinite-retention` | - | Never delete feature branches |
| `--scope-include-empty` | - | Include vulnerabilities with no path |
| `--scope-included-paths` | - | Paths for mono-repo scoping |
| `--scope-included-files` | - | File names for scoping |

**Examples:**

**1. Create project from repository:**
```shell
kdt create project \
  --repo-id https://github.com/kondukto-io/kdt \
  --alm-tool github \
  --labels GDPR,Internal \
  --team security
```

**2. Create with custom name:**
```shell
kdt create project \
  --repo-id https://gitlab.com/org/app \
  --project-name MyCustomName \
  --alm-tool gitlab \
  --default-branch develop
```

**3. Create with product:**
```shell
kdt create project \
  --repo-id https://github.com/org/repo \
  --alm-tool github \
  --product-name "Mobile_Apps" \
  --criticality-level 4
```

**4. Mono-repo with scoping:**
```shell
kdt create project \
  --repo-id https://github.com/org/monorepo \
  --project-name backend-api \
  --alm-tool github \
  --scope-included-paths "services/api,shared/common" \
  --scope-included-files "package.json,go.mod"
```

#### Create Team

**Flags:**
- `-n, --name`: Team name (required)
- `-r, --responsible`: Responsible user name

**Example:**
```shell
kdt create team --name "security-team" --responsible "john.doe"
```

#### Create Label

**Flags:**
- `-n, --name`: Label name (required)
- `-c, --color`: Label color in hex format (default: `000000`)

**Examples:**
```shell
# Create label with default color
kdt create label --name "GDPR"

# Create label with custom color
kdt create label --name "Critical" --color "FF0000"
```

#### Create Product

**Flags:**
- `-n, --name`: Product name (required)
- `-p, --projects`: Comma-separated project names or IDs

**Examples:**
```shell
# Create empty product
kdt create product --name "mobile-apps"

# Create product with projects
kdt create product --name "web-services" --projects "api-service,web-app,auth-service"
```

### SBOM Commands

#### Import SBOM

Import Software Bill of Materials (CycloneDX format).

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | SBOM file path (JSON format, required) |
| `--project` | `-p` | Project name or ID |
| `--repo-id` | `-r` | Repository URL or ID |
| `--branch` | `-b` | Branch name |
| `--sbom-type` | `-s` | Type: `source_dir`, `image`, `application`, `os`, `container` |
| `--allow-empty` | `-a` | Allow empty components |

**Examples:**

**1. Import SBOM for project:**
```shell
kdt sbom import \
  -f cyclonedx-sbom.json \
  -p MyProject \
  -b main
```

**2. Import with specific type:**
```shell
kdt sbom import \
  -f sbom.json \
  -p MyProject \
  -b main \
  --sbom-type image
```

**3. Import using repository ID:**
```shell
kdt sbom import \
  -f sbom.json \
  --repo-id https://github.com/org/repo \
  -b main
```

### Endpoint Commands

#### Import Endpoint

Import API endpoint definitions (Swagger/OpenAPI).

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Endpoint file path (required) |
| `--project` | `-p` | Project name or ID (required) |

**Example:**
```shell
kdt endpoint import -f swagger.json -p MyProject
```

### Status Command

Query project status and vulnerability counts.

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--project` | `-p` | Project name or ID |
| `--branch` | `-b` | Branch name |
| `--event` | `-e` | Event ID |
| `--threshold-crit` | - | Critical threshold |
| `--threshold-high` | - | High threshold |
| `--threshold-med` | - | Medium threshold |
| `--threshold-low` | - | Low threshold |
| `--threshold-risk` | - | Risk threshold |

**Examples:**

**1. Get project status:**
```shell
kdt status -p MyProject -b main
```

**2. Check status with thresholds:**
```shell
kdt status -p MyProject -b main \
  --threshold-crit 0 \
  --threshold-high 5
```

**3. Query by event ID:**
```shell
kdt status -e 5da6cafa5ab6e436faf643dc
```

### Project Commands

#### Check Project Availability

Check if a project exists in ALM.

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--alm-tool` | `-a` | ALM tool name |
| `--repo-id` | `-r` | Repository URL or ID |

**Example:**
```shell
kdt project available \
  --alm-tool github \
  --repo-id https://github.com/kondukto-io/kdt
```

Returns exit code 0 if available, -1 (255) if not.

## Advanced Usage Examples

### CI/CD Pipeline Integration

#### GitHub Actions

```yaml
- name: Run Security Scan
  run: |
    kdt scan \
      -p ${{ github.event.repository.name }} \
      -t semgrep \
      -b ${{ github.ref_name }} \
      --threshold-crit 0 \
      --threshold-high 10
```

#### GitLab CI

```yaml
security_scan:
  script:
    - kdt scan -p ${CI_PROJECT_NAME} -t trivy -b ${CI_COMMIT_BRANCH} --threshold-crit 0
```

#### Jenkins

```groovy
stage('Security Scan') {
  steps {
    sh '''
      kdt scan \
        -p ${JOB_NAME} \
        -t checkmarx \
        -b ${GIT_BRANCH} \
        --threshold-crit 0 \
        --threshold-high 5
    '''
  }
}
```

### Complex Workflow Examples

#### 1. Complete DevSecOps Pipeline

```shell
# Import scan results from local tool
kdt scan -p MyProject -t fortify -f results.fpr -b develop \
  --threshold-crit 0 --threshold-high 0

# Check release criteria
kdt release -p MyProject -b develop --sast --sca --dast
```

#### 2. Container Security Workflow

```shell
# Scan container image
kdt scan -p MyProject -t trivy \
  --image myapp:${VERSION} \
  -b main \
  --threshold-crit 0

# Import SBOM
kdt sbom import -f sbom.json -p MyProject -b main --sbom-type container
```

#### 3. Pull Request Workflow

```shell
# Trigger PR scan
kdt scan -p MyProject -t semgrep \
  -b feature/new-feature \
  -M main \
  --pr-number ${PR_NUMBER} \
  --pr-decoration-scanner-types all
```

#### 4. Multi-Environment Setup

```shell
# Development
kdt scan -p MyProject -t semgrep -b develop --env develop

# Staging
kdt scan -p MyProject -t sonarqube -b staging --env staging \
  --threshold-high 10

# Production
kdt scan -p MyProject -t checkmarx -b main --env production \
  --threshold-crit 0 --threshold-high 0 \
  --release-timeout 10
```

#### 5. Custom Parameters for Advanced Configuration

```shell
# Semgrep with custom rules
kdt scan -p MyProject -t semgrep -b main \
  --params=ruleset_type:2 \
  --params=ruleset_options.ruleset:/custom/rules/ \
  --params=ruleset_options.config:auto

# Container scan with custom registry
kdt scan -p MyProject -t trivy \
  --image registry.example.com/myapp:latest \
  --params=registry.username:user \
  --params=registry.password:pass
```

## Exit Codes

KDT uses the following exit codes:

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Warning |
| `100` | Not authorized |
| `-1` (255) | Negative response (e.g., project not available) |

## Supported Scanners

KDT supports all scanners enabled in your Kondukto instance. To view available scanners:

```shell
kdt list scanners
```

## Troubleshooting

### Enable Verbose Logging

```shell
kdt -v scan -p MyProject -t semgrep -b main
```

### Test Connection

```shell
kdt ping
kdt ping -a  # With authentication
```

### Verify Configuration

```shell
# Check if host and token are set
echo $KONDUKTO_HOST
echo $KONDUKTO_TOKEN

# Or use a test command
kdt list projects
```

## Contributing

Contributions to KDT are welcome! Here's how you can contribute:

### Reporting Issues

Create an issue in the [GitHub repository](https://github.com/kondukto-io/kdt/issues) with:
- Clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- KDT version (`kdt version`)

### Pull Requests

1. Fork the repository
2. Create a feature/bugfix branch following [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/):
   - Features: `feature/example-feature`
   - Bugfixes: `bugfix/example-bugfix`
3. Write idiomatic Go code
4. Document exported functions
5. Write detailed PR description
6. Ensure tests pass

### Development Setup

```shell
git clone https://github.com/kondukto-io/kdt.git
cd kdt
go mod download
go build -o kdt
./kdt --help
```

## License

See the [LICENSE](LICENSE) file for details.

## Support

- Documentation: [https://docs.kondukto.io](https://docs.kondukto.io)
- Issues: [GitHub Issues](https://github.com/kondukto-io/kdt/issues)
- Website: [https://kondukto.io](https://kondukto.io)
