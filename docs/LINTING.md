# Linting in JumpCloud Provider

This document describes the linting errors identified in the project and the plan to gradually fix them.

## Overview

We use the `tfproviderlint` tool to check code compliance with best practices for Terraform providers. Several issues have been identified and will be fixed in phases.

We also use `golangci-lint` for general Go linting and static analysis. This project uses Go 1.20.

## Fixed Errors

- ✅ **R014**: Parameters of type `interface{}` must be named `meta`. This error has been fixed in all files.

## Pending Errors

The following linting errors will be fixed in future phases:

### Errors in Acceptance Tests

- **AT001**: Missing CheckDestroy - acceptance tests should include a destruction check to ensure resources are properly cleaned up.
- **AT005**: Acceptance test function names should begin with `TestAcc`.
- **AT012**: File contains multiple acceptance test name prefixes, which can cause confusion.

### Errors in Resources

- **R001**: The key argument for `ResourceData.Set()` must be a string literal, not a variable.
- **R017**: Schema attributes should be stable across Terraform runs to avoid state issues.
- **R019**: `d.HasChanges()` has many arguments, consider using `d.HasChangesExcept()`.

### Validation Errors

- **V013**: Custom SchemaValidateFunc should be replaced with `validation.StringInSlice()` or `validation.StringNotInSlice()`.

## Correction Plan

To facilitate the correction process, we will follow this order:

1. **Phase 1**: Fix R001 (ResourceData.Set with string literal)
2. **Phase 2**: Fix R019 (HasChanges → HasChangesExcept)
3. **Phase 3**: Fix V013 (SchemaValidateFunc → validation.StringInSlice)
4. **Phase 4**: Fix R017 (Schema attributes should be stable)
5. **Phase 5**: Fix AT* (acceptance test issues)

## CI/CD Configuration

To prevent linting errors from blocking development while they are gradually fixed, we have implemented the following solutions:

### GitHub Actions

The pull request verification workflow (`.github/workflows/pr-check.yml`) has been configured to:

1. Run tfproviderlint verification only for critical errors and R014 (already fixed)
2. Temporarily ignore errors that will be fixed in phases
3. List pending errors for reference

As each correction phase is completed, the workflow will be updated to enable verification of the corresponding rules.

### golangci-lint

In addition to tfproviderlint, we also use golangci-lint for more comprehensive static code analysis. For information on configuration and troubleshooting with golangci-lint, see the [GOLANGCI-LINT.md](./GOLANGCI-LINT.md) document.

We've configured golangci-lint with a suitable configuration for Go 1.20, enabling the most useful linters.

### Local Scripts

We provide several scripts in the `scripts/linting/` directory to help with local verification:

- `scripts/linting/check_required_lint.sh`: Checks only critical tfproviderlint errors, ignoring those to be addressed in phases.
- `scripts/linting/run_linter.sh`: Provides options for checking specific tfproviderlint errors and information on how to run the linter.
- `scripts/linting/run_golangci_lint.sh`: Runs golangci-lint with the same configuration as CI to ensure consistent results.

## How to Contribute to Corrections

If you want to contribute to fixing linting errors, follow these steps:

1. Choose a phase to work on based on the correction plan.
2. Run the specific lint for the rule you're fixing:
   ```
   $HOME/go/bin/tfproviderlint -AT=false -R=false -S=false -V=false -<RULE>=true ./...
   ```
   Or use our helper script:
   ```
   ./scripts/linting/run_linter.sh
   ```
3. For general Go linting, use:
   ```
   ./scripts/linting/run_golangci_lint.sh
   ```
4. Make the necessary corrections to the indicated files.
5. Run tests to ensure your changes haven't caused regressions.
6. Submit a PR with a clear description of the corrections made.

## Linting Error Details

### R001: ResourceData.Set() with string literal

```go
// Incorrect
key := "attribute_name"
d.Set(key, value)

// Correct
d.Set("attribute_name", value)
```

### R019: HasChanges → HasChangesExcept

```go
// Incorrect
if d.HasChanges("attr1", "attr2", "attr3", "attr4", "attr5") {
    // ...
}

// Correct
if d.HasChangesExcept("attr6", "attr7") {
    // ...
}
```

### V013: SchemaValidateFunc → validation.StringInSlice

```go
// Incorrect
ValidateFunc: func(v interface{}, k string) (warns []string, errs []error) {
    value := v.(string)
    validValues := []string{"one", "two", "three"}
    valid := false
    for _, val := range validValues {
        if value == val {
            valid = true
            break
        }
    }
    if !valid {
        errs = append(errs, fmt.Errorf("%s must be one of %v, got: %s", k, validValues, value))
    }
    return
},

// Correct
ValidateFunc: validation.StringInSlice([]string{"one", "two", "three"}, false),
```