# JumpCloud Provider Integration Tests

This directory contains integration tests for the JumpCloud Terraform Provider. 
Unlike unit tests and resource-specific acceptance tests, these tests verify the interaction between different resources and data sources across modules.

## Structure

- `scenarios/`: Contains test scenarios organized by use case
  - `user_management/`: Tests for user, group and access management workflows
  - `application_access/`: Tests for application deployment and access control
  - `system_management/`: Tests for system provisioning and configuration
  - `authentication/`: Tests for authentication policies and access rules

## Running Tests

Integration tests require a JumpCloud account with valid credentials. Set the following environment variables:

```bash
export JUMPCLOUD_API_KEY="your-api-key"
export JUMPCLOUD_ORG_ID="your-org-id"  # Optional, for multi-tenant environments
export TF_ACC=1  # Enables acceptance testing
```

To run all integration tests:

```bash
go test -v ./integration_tests/...
```

To run a specific scenario:

```bash
go test -v ./integration_tests/scenarios/user_management
```

## Creating New Tests

When creating new integration tests:

1. Place them in an appropriate scenario directory
2. Ensure they're properly isolated to avoid conflicts with other tests
3. Use unique name prefixes for all created resources
4. Add appropriate cleanup logic to handle test failures

Each test should verify a specific workflow that includes multiple resource types working together. 