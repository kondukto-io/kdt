<p align="center"><a href="https://kondukto.io" target="_blank" rel="noopener noreferrer"><img width="200" src="https://kondukto.io/logo.png" alt="Kondukto logo"></a></p>

# KDT
KDT is a command line client for [Kondukto](https://kondukto.io) written in [Go](https://golang.org). It interacts with Kondukto engine through public API. 

With KDT, you can list projects and their scans in **Kondukto**, and restart a scan with a specific application security tool. KDT is also easy to use in CI/CD pipelines to trigger scans and break releases if a scan fails or scan results don't met specified release criteria. 

### What is Kondukto?
[Kondukto](https://kondukto.io) is an Application Security Testing Orchestration and DevSecOps platform that helps you centralize and automate your entire AppSec related vulnerability management process. Providing an interface where security health of applications can be continuously monitored, and a command line interface where your AppSec operations can be integrated into DevOps pipelines, Kondukto lets you manage your AppSec processes automatically with ease.

## Installation
You can install the CLI with a `curl` utility script or by downloading the pre-compiled binary from the Github release page.
Once installed youl'll get the `kdt-cli` command and `kdt` alias.

Utility script with `curl`:
```
$ curl -sSL https://cli.kondukto.io | sudo sh
```

Non-root with curl:
```
$ curl -sSL https://cli.kondukto.io | sh
```

### Windows 
To install the kdt-cli on Windows go to [Releases](https://github.com/kondukto-io/kdt/releases) and download the latest kdt-cli.exe.


Or you can also simply run the following if you have an existing [Go](https://golang.org) environment:
```
go get github.com/kondukto-io/kdt
```

If you want to build it yourself, clone the source files using Github, change into the `kdt` directory and run:
```
git clone https://github.com/kondukto-io/kdt.git
cd kdt
go install
```

## Configuration
KDT needs Kondukto host and an API token for authentication. API tokens can be created under Integrations/API Tokens menu.

You can provide configuration via three different ways:

##### 1) Setting environment variables: 

*(example is for BASH shell)*
```
$ export KONDUKTO_HOST=http://localhost:8080
$ export KONDUKTO_TOKEN=WmQ2eHFDRzE3elplN0ZRbUVsRDd3VnpUSHk0TmF6Uko5OGlyQ1JvR2JOOXhoWEFtY2ZrcDJZUGtrb2tV
```
It is always better to set environment variables in shell profile files(`~/.bashrc`, `~/.zshrc`, `~/.profile` etc.)
##### 2) Providing a configuration file.

Default path for config file is `$HOME/.kdt.yaml`. If you put your config file with this name`.kdt.yaml`, the binary will load the configuration without explicitly pointing the config file. Otherwise a custom config file can be provided with `--config` command line flag.
```
kdt --config=config.yaml list projects
```

A config file example.
```
// $HOME/.kdt.yaml 
host: http://localhost:8088
token: WmQ2eHFDRzE3elplN0ZRbUVsRDd3VnpUSHk0TmF6Uko5OGlyQ1JvR2JOOXhoWEFtY2ZrcDJZUGtrb2tV
insecure: true 
```

##### 3) Using command line flags
```
kdt list projects --host http://localhost:8088 --token WmQ2eHFDRzE3elplN0ZRbUVsRDd3VnpUSHk0TmF6Uko5OGlyQ1JvR2JOOXhoWEFtY2ZrcDJZUGtrb2tV
```

## Running
KDT comes with an internal documentation. To see the documentation just type `kdt --help`. 
Most KDT commands are straightforward but for the details of a command you can always take a peak to the documentation. `kdt <command> --help` or `kdt <command> <sub-command> --help`. 

To list projects: `kdt list projects`

To list the supported scanners and scan methods: `kdt list scanners`

To list scans of a project: `kdt list scans -p ExampleProject`

To restart a scan, you can use one of the following:

- id of the scan: `kdt scan -s 5da6cafa5ab6e436faf643dc`

- project and tool names: `kdt scan -p ExampleProject -t ExampleTool`

To import scan results as a file: `kdt scan -p ExampleProject -t ExampleTool -f results.json -b master`

## Command Line Flags
KDT has several helpful flags to manage scans.

#### Global flags

Following flags are valid for all commands of KDT.

`--host`: HTTP address of Kondukto server with port

`--token`: API token generated by Kondukto

`--config`: Configuration file to use instead of default one(`$HOME/.kdt.yaml`)

`--async`: Starts an asynchronous scan that won't block process to wait for scan to finish. KDT will exit gracefully when scan gets started successfully.

`--insecure`: If provided, client skips verification of server's certificates and host name. In this mode, TLS is susceptible to man-in-the-middle attacks. Not recommended unless you really know what you are doing!

`-v` or `--verbose`: Prints more and detailed logs. Useful for debugging.

#### Scan Commands Flags
Following flags are only valid for scan commands.

`-p` or `--project` for providing project name or id

`-t` or `--tool` for providing tool name

`-s` or `--scan-id` for providing scan id

`-b` or `--branch` for providing branch to scan

##### Threshold flags

These flags represent maximum number of vulnerabilities with specified severity to ignore. If these threshold are crossed, KDT will exit with non-zero status code.


`--threshold-crit` 

`--threshold-high` 

`--threshold-med` 

`--threshold-low` 

`--threshold-risk` for failing tests if the scan causes a higher risk score than the last scan's risk score. Useful for keeping a project's security level under control. If used with every scan in DevOps pipelines, you will make sure that the project will never get more vulnerable.

*Risk threshold considers only the last two scans with the same tool. If the project does not have a scan with the tool, KDT will fail since it will not be able to compare risk scores.*

*Threshold flags don't work with `--async` flag since KDT will exit when scan gets started, and won't be able to check scan results*

Example Usage:

`kdt scan -p SampleProject -t SampleTool --threshold-crit 3 --threshold-high 10 --threshold-risk`

## Supported scanners (tools)
KDT supports all scanners enabled in Kondukto server, to see the list simply run `kdt list scanners`. Kondukto may provide slightly different triggering options for some scanners. This command also shows the `custom parameters` and alternative trigger options for each enabled scanner

Example Usage:

```
./kdt --config kondukto.yaml list scanners
Name       ID                          Type    Trigger     Labels                      Custom Parameters
----       --                          ----    -------     ------                      -----------------
gosec      60eec8a83e9e5e6e2ae52d06    sast    new scan    docker,kdt
semgrep    60eec8a53e9e5e6e2ae52d05    sast    rescan      template,docker,kdt
semgrep    5fbf6836f65176f39df675c4    sast    rescan      template,docker,kdt,import                                          
                                                                                     --params=ruleset_type: the type of rulesets, allowed values are known_ruleset: 0, ruleset_url: 1 and ruleset_path: 2
                                                                                     --params=ruleset_options.exclude: excludes the given path from the scanning scope
                                                                                     --params=ruleset_options.ruleset: a ruleset value. i.e: gosec, bandit, golang, https://semgrep.com/example ...
```

### Custom Parameters
For some tools the default behaviour of KDT is re-triggering or re-scanning of an existing scan. This means that, there should be a configured scan on Kondukto for KDT to run re-run this operation
from the CLI. However, by passing some custom arguments to the scanner, Kondukto server can create and start a scan without having a configuration. 

The scanners that supports customer parameters are shown with `--params` arguments.
A customized scan example:
```
# Run a scan using semgrep on a develop branch
# with a custom rule path
kdt scan -p SampleProject \
         -b develop -t semgrep \
	 --params=ruleset_type:2 --params=ruleset_options.ruleset:/rules/
```

## Advanced usage examples
There are multiple cases that you can use KDT in your pipeline.

```
kdt --config kondukto-config.yml \
    --insecure \
    scan \
    --project SampleProject \
    --tool fortify \
    --file results.fpr \
    --branch develop \
    --threshold-crit 0 \
    --threshold-high 0` 
```

- --config: Kondukto configuration file in yaml format
- --insecure: Don't verify SSL certificates
- scan: start scan 
- --project: Application's name in Kondukto server. 
- --tool: AST tool to be used (fortify). 
- --file: Results filename, when file parameters is given scan will not be initiated and only the results file (results.fpr) is going to be analyzed.
- --branch: Branch name.
- --threshold-crit: Threshold value to "break the build" in the pipeline. When this parameter(s) is given, entered security criteria will be overwritten.

```
kdt --config kondukto-config.yml \
    scan \
    --project SampleProject \
    --tool trivy \
    --image ubuntu@256:ab02134176aecfe0c0974ab4d3db43ca91eb6483a6b7fe6556b480489edd04a1 \
    --branch develop
```
- --config: Kondukto configuration file in yaml format
- scan: start scan 
- --project: Application's name in Kondukto server. 
- --tool: AST tool to be used (trivy). 
- --image: Image name to be scanned. Name can be given with the digest or with the tag name (ubuntu:latest).

```
kdt --config kondukto-config.yml \
    create \
    project \ 
    --repo-id https://github.com/kondukto-io/kdt \
    --labels GDPR,Internal \
    --alm-tool github
```
- --config: Kondukto configuration file in yaml format
- create: Base command for create operation.
- project: Subcommand to create project.
- --repo-id: Project repository URL or ALM ID.
- --labels: Associate project with a label list
- --alm-tool: If there is more than one ALM enabled in Kondukto you need to specify ALM tool, otherwise it is not necessary.
- --team: Specify a team name. By default team name is `default team`. 

This command will create a project on Kondukto with the same name in your ALM(Application Lifecycle Management) tool. If there is another project
with the same name, command will print an error message and exit with a status code. You can pass `--force-create` flag to overwrite this behaviour.

```
kdt --config kondukto-config.yml \
    sbom import \
    --file cyclonedx-sbom.json \
    --project SampleProject \
    --branch develop
```
- --config: Kondukto configuration file in yaml format
- sbom import: import an SBOM file  
- --file: CycloneDX SBOM output in JSON format.
- --project: Application's name in Kondukto server. 
- --branch: the branch of the application

This command will update a generated SBOM file to Kondukto. Currently only [CycloneDX](https://cyclonedx.org/) standard is supported. 

## Contributing to KDT
If you wish to get involved in KDT development, create issues for problems and missing features or fork the repository and create pull requests to help the development directly.

Before sending your PRs:
- Create and name your branches according to [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/) methodology.

    For new features: `feature/example-feature-branch`

    For bug fixes: `bugfix/example-bugfix-branch`

- Properly document your code following idiomatic [Go](https://golang.org) practices. Exported functions should always be commented.

- Write detailed PR descriptions and comments
