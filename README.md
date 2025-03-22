# JumpCloud Terraform Provider

This Terraform provider allows you to manage JumpCloud resources through Terraform. It provides a convenient way to create, read, update, and delete JumpCloud resources such as users, systems, user groups, and more.

## Folder Structure

The provider follows a clean architecture approach with the following folder structure:

```
terraform-provider-jumpcloud/
├── cmd/                 # Main CLI entry point 
├── pkg/                 # Shared packages
│   ├── apiclient/       # JumpCloud API client
│   ├── errors/          # Standardized error handling
│   └── utils/           # Shared utilities
├── internal/            # Implementation details
│   ├── provider/        # Provider implementation
│   │   ├── resources/   # Resource implementations (domain-specific)
│   │   │   ├── user/    # User resource domain
│   │   │   ├── system/  # System resource domain
│   │   │   └── ...
│   │   └── provider.go  # Provider definition
└── test/                # Test suite
    ├── acceptance/      # Acceptance tests
    ├── integration/     # Integration tests
    └── unit/            # Unit tests
```

## Getting Started

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.14.x
- [Go](https://golang.org/doc/install) >= 1.18 (for building the provider)

### Building the Provider

Clone the repository:

```shell
git clone https://github.com/agilize/terraform-provider-jumpcloud.git
```

Build the provider:

```shell
cd terraform-provider-jumpcloud
go build -o terraform-provider-jumpcloud
```

### Using the Provider

To use the provider, configure it with your JumpCloud API key:

```hcl
provider "jumpcloud" {
  api_key = "your_api_key"      # Or use JUMPCLOUD_API_KEY environment variable
  org_id  = "your_org_id"       # For multi-tenant environments (optional)
  api_url = "custom_api_url"    # Override the default API URL (optional)
}

# Create a JumpCloud user
resource "jumpcloud_user" "example" {
  username             = "example.user"
  email                = "example.user@example.com"
  first_name           = "Example"
  last_name            = "User"
  password             = "SecurePassw0rd!"
  mfa_enabled          = true
  password_never_expires = false
}
```

## Developing the Provider

### Testing

This provider uses a comprehensive testing approach:

- **Unit Tests**: Tests individual components in isolation
- **Integration Tests**: Tests API client against the real JumpCloud API
- **Acceptance Tests**: Full end-to-end tests of resources using real infrastructure

To run tests:

```shell
# Run unit tests
go test ./... -v

# Run acceptance tests (creates real resources)
TF_ACC=1 JUMPCLOUD_API_KEY=your_api_key go test ./internal/provider -v
```

### Adding a New Resource

To add a new resource, follow these steps:

1. Create a new directory under `internal/provider/resources/` for the resource domain
2. Implement the resource in a `resource.go` file in that directory
3. Use the `errors` package for standardized error handling
4. Add the resource to the provider in `internal/provider/provider.go`
5. Add tests in the appropriate test directories

Example:

```go
package newresource

import (
    "context"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
    "registry.terraform.io/agilize/jumpcloud/pkg/errors"
)

// Resource structure and functions go here
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the [Mozilla Public License v2.0](LICENSE).

## Support

For support, please contact the project maintainers or open an issue on GitHub. 