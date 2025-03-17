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
    args: --timeout=5m --disable=goanalysis_metalinter --config=.golangci.yml
    skip-cache: true
```

We explicitly disable the `goanalysis_metalinter` to prevent compatibility issues with Go 1.22 and use a custom configuration file (`.golangci.yml`).

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
- Disables all linters by default and then selectively enables the ones we want
- Sets a reasonable timeout
- Configures output and issue limits

## Common Errors and Solutions

### 1. "Unsupported version" Error

**Symptom:**
```
level=error msg="Running error: 1 error occurred:\n\t* can't run linter goanalysis_metalinter: buildir: failed to load package goarch: could not load export data: internal error in importing \"internal/goarch\" (unsupported version: 2); please report an issue\n\n"
```

**Cause:** Incompatibility between the Go version and the golangci-lint version, specifically with the `goanalysis_metalinter`.

**Solution:** 
1. Disable the problematic metalinter with `--disable=goanalysis_metalinter`
2. Use a custom configuration file to control which linters are enabled
3. Make sure to use v1.57.0 or newer of golangci-lint

### 2. Timeout During Execution

**Symptom:** golangci-lint execution fails with a timeout error.

**Solution:** Increase the timeout value in the arguments:
```yaml
args: --timeout=10m
```

### 3. Cache Problems

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
golangci-lint run --timeout=5m --disable=goanalysis_metalinter --config=.golangci.yml
```

For more information on configuration, see the [official golangci-lint documentation](https://golangci-lint.run/usage/configuration/). 