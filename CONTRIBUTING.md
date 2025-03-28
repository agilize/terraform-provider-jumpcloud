# Contributing to the JumpCloud Terraform Provider

Thank you for your interest in contributing to the JumpCloud Terraform Provider! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting Started

### Fork and Clone the Repository

1. Fork the repository on GitHub.
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/terraform-provider-jumpcloud.git
   cd terraform-provider-jumpcloud
   ```
3. Add the original repository as an upstream remote:
   ```bash
   git remote add upstream https://github.com/jumpcloud/terraform-provider-jumpcloud.git
   ```

### Development Environment Setup

1. Install Go (version 1.16 or later).
2. Install Terraform (version 0.12 or later).
3. Install development dependencies:
   ```bash
   go mod download
   ```

## Development Workflow

### Creating a New Feature

1. Create a new branch for your feature:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. Make your changes following our coding standards and organization:
   - Put domain-specific resources in their own packages under `jumpcloud/`.
   - Create appropriate test files for each resource.
   - Add documentation in README.md files.

3. Run tests to ensure your changes work correctly:
   ```bash
   go test ./...
   ```

4. For acceptance tests (requires a JumpCloud account):
   ```bash
   export JUMPCLOUD_API_KEY="your-api-key"
   export JUMPCLOUD_ORG_ID="your-org-id"  # Optional
   export TF_ACC=1
   go test ./... -v -run=TestAcc
   ```

5. Commit your changes with a clear, descriptive commit message:
   ```bash
   git commit -m "feature: Add support for X resource"
   ```

### Pull Request Process

1. Push your branch to your fork:
   ```bash
   git push origin feature/my-new-feature
   ```

2. Open a pull request against the main repository's main branch.

3. In your pull request description, include:
   - A clear description of the changes
   - Any relevant documentation updates
   - Test cases you've added
   - Any issues that are fixed by this PR

4. Your pull request will be reviewed by maintainers who may request changes.

5. Once approved, your PR will be merged by a maintainer.

## Coding Standards

### Go Code

- Follow the standard Go formatting guidelines (use `go fmt`).
- Use meaningful variable and function names.
- Add comments for exported functions and complex logic.
- Use proper error handling.
- Create appropriate tests for all code.

### Resource Implementation

- Each resource should be in its domain-appropriate directory.
- Resources should include:
  - CRUD functions
  - Schema definition
  - Proper error handling
  - Documentation comments

### Testing

- Write unit tests for all functions where possible.
- Create acceptance tests for resources that interact with the JumpCloud API.
- Test both success and error cases.

## Documentation

- Update the README.md in the appropriate directory when adding or modifying resources.
- Include examples of how to use the new resources.
- Document all resource arguments and attributes.

## Adding a New Resource

When adding a new resource, follow this process:

1. Determine the appropriate domain package for the resource.
2. Create the resource file in that package.
3. Implement the CRUD operations.
4. Create test files for unit and acceptance tests.
5. Update the provider.go file to register the new resource.
6. Add documentation in the appropriate README.md file.
7. Create examples in the examples directory.

## Releasing

Provider releases are managed by the maintainers. To suggest a release:

1. Update CHANGELOG.md with your changes.
2. Ensure all tests pass.
3. Open an issue suggesting a new release with a version number following semantic versioning.

## Getting Help

If you have questions or need help, you can:

- Open an issue in the GitHub repository
- Contact the maintainers via [appropriate contact method]

Thank you for contributing to the JumpCloud Terraform Provider! 