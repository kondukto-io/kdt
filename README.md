# KDT
KDT is a command line client for [Kondukto](https://kondukto.io) written in [Go](https://golang.org). It interacts with Kondukto engine through public API. 

With KDT, you can list projects and their scans in Kondukto, and restart a scan with a specific application security tool. KDT is also easy to use in CI/CD processes to trigger scans and breaking releases if a scan fails or scan results doesn't met specified release criteria. 

## Configuration
KDT needs Kondukto host and an API token for authentication. You can provide these with 3 methods. API tokens can be created under Integrations/API Tokens menu.

1. Setting environment variables: (example is for BASH shell)
```
$ export KONDUKTO_HOST=http://localhost:8080
$ export KONDUKTO_TOKEN=WmQ2eHFDRzE3elplN0ZRbUVsRDd3VnpUSHk0TmF6Uko5OGlyQ1JvR2JOOXhoWEFtY2ZrcDJZUGtrb2tV
```

2. Providing a configuration file.


Default path for config file is `$HOME/.kdt.yaml`. Another file can be provided with `--config` command line flag.
```
// $HOME.kdt.yaml 
```
