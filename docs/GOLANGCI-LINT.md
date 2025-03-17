# golangci-lint Configuration

This document provides guidance on using golangci-lint in the terraform-provider-jumpcloud project.

## Compatible Versions

golangci-lint requires specific compatibility with the Go version. In this project, we are using:

- **Go version:** 1.22
- **golangci-lint version:** v1.57.0

### Compatibility Table

| Go Version | Recommended minimum golangci-lint version |
|------------|-----------------------------------------|
| Go 1.22    | v1.56.0+                                |
| Go 1.21    | v1.54.0+                                |
| Go 1.20    | v1.53.0+                                |
| Go 1.19    | v1.50.0+                                |

## CI/CD Configuration

In GitHub Actions, we use golangci-lint as part of the PR verification process:

```yaml
- name: golangci-lint
  uses: golangci/golangci-lint-action@v3
  with:
    version: v1.57.0
    args: --timeout=5m --config=.golangci.yml
    skip-cache: true
```

We use a custom configuration file (`.golangci.yml`) that is configured to work with Go 1.22 by disabling problematic linters.

## Configuration File

We use a `.golangci.yml` file in the project root with the following configuration:

```yaml
# Configuration file for golangci-lint
run:
  timeout: 5m

linters:
  disable-all: true
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gosec
    - gofmt
    - goimports
    - misspell
    - unconvert

issues:
  max-issues-per-linter: 0
  max-same-issues: 0

output:
  sort-results: true
```

This configuration:
- Disables all linters by default (`disable-all: true`) and then selectively enables only the linters we want to use
- Sets a reasonable timeout
- Configures output and issue limits

By using `disable-all: true`, we prevent compatibility issues with problematic linters such as `goanalysis_metalinter` that may cause errors with Go 1.22.

## Common Errors and Solutions

### 1. "Unsupported version" Error

**Symptom:**
```
level=error msg="Running error: 1 error occurred:\n\t* can't run linter goanalysis_metalinter: buildir: failed to load package goarch: could not load export data: internal error in importing \"internal/goarch\" (unsupported version: 2); please report an issue\n\n"
```

**Cause:** Incompatibility between the Go version and the golangci-lint version, specifically with the `goanalysis_metalinter`.

**Solution:** 
1. Use a custom configuration file (`.golangci.yml`)
2. In the configuration file, use `disable-all: true` and explicitly enable only compatible linters
3. Make sure to use v1.57.0 or newer of golangci-lint

### 2. "Can't combine options --disable-all and --disable" Error

**Symptom:**
```
Error: can't combine options --disable-all and --disable
```

**Cause:** Trying to use both `disable-all: true` in the config file and `--disable` flag on the command line.

**Solution:**
Either:
- Remove the `--disable` flag from the command line (preferred approach), or
- Remove `disable-all: true` from the config file and list specific linters to disable

### 3. Timeout During Execution

**Symptom:** golangci-lint execution fails with a timeout error.

**Solution:** Increase the timeout value in the arguments:
```yaml
args: --timeout=10m
```

### 4. Cache Problems

**Symptom:** Inconsistent errors between executions.

**Solution:** Disable the cache to ensure clean execution:
```yaml
skip-cache: true
```

## Local Configuration

To run golangci-lint locally with the same configuration as CI:

```bash
# Install the correct golangci-lint version
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.0

# Run lint
golangci-lint run --timeout=5m --config=.golangci.yml
```

For more information on configuration, see the [official golangci-lint documentation](https://golangci-lint.run/usage/configuration/). 