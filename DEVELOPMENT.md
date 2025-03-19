# Development Guide

This document provides guidelines and instructions for developers contributing to the Terraform JumpCloud Provider.

## Prerequisites

Before you begin development, ensure you have the following installed:

- [Go](https://golang.org/doc/install) (version 1.17 or higher)
- [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli) (version 1.0.0 or higher)
- [golangci-lint](https://golangci-lint.run/usage/install/) for code linting
- Git for version control

## Development Workflow

### Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/agilize/terraform-provider-jumpcloud.git
   cd terraform-provider-jumpcloud
   ```

2. **Initialize Go modules**:
   ```bash
   go mod tidy
   ```

3. **Install the provider locally** (for testing with Terraform):
   ```bash
   make install
   ```

### Development Process

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Follow the architecture and coding standards outlined below
   - Add appropriate documentation
   - Write tests for your changes

3. **Test your changes**:
   ```bash
   make test                # Run unit tests
   make test-integration    # Run integration tests (requires JumpCloud API credentials)
   make test-acceptance     # Run acceptance tests (requires JumpCloud API credentials)
   ```

4. **Format, lint, and vet your code**:
   ```bash
   make fmt lint vet
   ```

5. **Build the provider**:
   ```bash
   make build
   ```

6. **Commit and push your changes**:
   ```bash
   git add .
   git commit -m "Description of your changes"
   git push origin feature/your-feature-name
   ```

7. **Create a pull request** to merge your changes into the main branch

## Project Structure

```
terraform-provider-jumpcloud/
├── docs/                     # Documentation
│   ├── data-sources/         # Data source documentation
│   └── resources/            # Resource documentation
├── examples/                 # Example configurations
│   ├── data-sources/         # Data source examples
│   └── resources/            # Resource examples
├── internal/                 # Provider implementation
│   ├── client/               # API client
│   └── provider/             # Terraform provider
├── .github/                  # GitHub workflows
├── .gitignore                # Git ignore file
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
├── LICENSE                   # License file
├── Makefile                  # Build and development tasks
└── README.md                 # Project overview
```

## Architecture

The provider follows a clean architecture pattern with clear separation of concerns:

1. **Provider Layer** (`internal/provider/`)
   - Defines the Terraform provider schema
   - Implements resource and data source CRUD operations
   - Maps between Terraform and API models

2. **Client Layer** (`internal/client/`)
   - Handles API communication with JumpCloud
   - Manages authentication
   - Provides error handling
   - Maps API responses to models

## Coding Standards

### General Guidelines

- Use American English for all code, comments, and documentation
- Follow Go standards for naming and formatting
- Document all exported functions, types, and constants
- Use descriptive variable names
- Keep functions small and focused on a single responsibility
- Write tests for all new functionality

### Naming Conventions

- **Packages**: lowercase, single word (e.g., `provider`, `client`)
- **Interfaces**: PascalCase, descriptive names (e.g., `ClientInterface`)
- **Structs**: 
  - PascalCase for exported structs (e.g., `JumpCloudClient`)
  - camelCase for internal structs (e.g., `userResource`)
- **Methods/Functions**: 
  - PascalCase for exported methods (e.g., `CreateUser`)
  - camelCase for internal methods (e.g., `parseResponse`)
- **Variables**: camelCase (e.g., `userID`, `resourceData`)
- **Constants**: SNAKE_CASE uppercase (e.g., `ERROR_NOT_FOUND`)

### Error Handling

- Always check errors and provide context
- Use custom error types where appropriate
- Log relevant error details
- Never silently ignore errors

### Security Practices

- Never log sensitive information (API keys, passwords, etc.)
- Mark sensitive fields with `Sensitive: true` in schemas
- Use secure transport (HTTPS) for all API communication
- Follow the principle of least privilege

## Testing

### Unit Tests

Unit tests should focus on testing individual components in isolation:

```bash
make test
```

### Integration Tests

Integration tests verify the interaction between the provider and the JumpCloud API:

```bash
make test-integration
```

### Acceptance Tests

Acceptance tests use Terraform to create real resources in JumpCloud:

```bash
make test-acceptance
```

### Performance Tests

Performance tests validate the efficiency of the implementation:

```bash
make test-performance
```

### Security Tests

Security tests check for security vulnerabilities:

```bash
make test-security
```

## Documentation

- Document all resources and data sources in the `docs/` directory
- Provide examples in the `examples/` directory
- Use clear and concise language
- Include usage examples and descriptions of all arguments and attributes
- Reference the JumpCloud API documentation where appropriate

## Release Process

1. Update the version in `go.mod`
2. Update the CHANGELOG.md file
3. Create a new Git tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
4. Build the release artifacts:
   ```bash
   make release
   ```

## Getting Help

If you have questions or need assistance, please:

- Open an issue in the GitHub repository
- Refer to the JumpCloud API documentation
- Check the existing documentation and code comments 