# KDT
KDT is a command line client for [Kondukto](https://kondukto.io) written in [Go](https://golang.org). It interacts with Kondukto engine through public API. 

With KDT, you can list projects and their scans in **Kondukto**, and restart a scan with a specific application security tool. KDT is also easy to use in CI/CD pipelines to trigger scans and break releases if a scan fails or scan results don't met specified release criteria. 

## Installation
If you just wish to download and run a portable binary, you can get one of pre-compiled binaries for your system from Github releases page.

You can also simply run the following if you have an existing [Go](https://golang.org) environment:
```
go install github.com/kondukto-io/kdt
```

If you want to build it yourself, clone the source files using Github, change into the `kdt` directory and run:
```
git clone https://github.com/kondukto-io/kdt.git
cd kdt
go install
```

## Configuration
KDT needs Kondukto host and an API token for authentication. API tokens can be created under Integrations/API Tokens menu.

You can provide configuration by:

##### 1) Setting environment variables: 

*(example is for BASH shell)*
```
$ export KONDUKTO_HOST=http://localhost:8080
$ export KONDUKTO_TOKEN=WmQ2eHFDRzE3elplN0ZRbUVsRDd3VnpUSHk0TmF6Uko5OGlyQ1JvR2JOOXhoWEFtY2ZrcDJZUGtrb2tV
```
It is always better to set environment variables in shell profile files(`~/.bashrc`, `~/.zshrc`, `~/profile` etc.)
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

## Running
Most KDT commands are straightforward.

To list projects: `kdt list project`

To list scans of a project: `kdt list scans -p ExampleProject`

To restart a scan, you can use of the following:

- id of the scan: `kdt scan -s 5da6cafa5ab6e436faf643dc`

- project and tool names: `kdt scan -p ExampleProject -t ExampleTool`

## Command Line Flags
KDT has several helpful flags to manage scans.

#### Global flags

Following flags are valid for all commands of KDT.

`--async`: Starts an asynchronous scan that won't block process to wait for scan to finish. KDT will exit gracefully when scan gets started successfully.

`--insecure`: If provided, client skips verification of server's certificates and host name. In this mode, TLS is susceptible to man-in-the-middle attacks. Not recommended unless you really know what you are doing!

`-v` or `--verbose`: Prints more and detailed logs. Useful for debugging.

#### Scan Commands Flags
Following flags are only valid for scan commands.

`-p` or `--project` for providing project name or id

`-t` or `--tool` for providing tool name

`-s` or `--scan-id` for providing scan id


## Contributing to KDT
If you wish to get involved in KDT development, create issues for problems and missing features or fork the repository and create pull requests to help the development directly.

Before sending your PRs:
- Create and name your branches according to [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/) methodology.

    For new features: `feature/example-feature-branch`

    For bug fixes: `bugfix/example-bugfix-branch`

- Properly document your code following idiomatic [Go](https://golang.org) practices. Exported functions should always be commented.

- Write detailed PR descriptions and comments
