# Terraform Provider for JumpCloud

[![Go Report Card](https://goreportcard.com/badge/github.com/jumpcloud/terraform-provider-jumpcloud)](https://goreportcard.com/report/github.com/jumpcloud/terraform-provider-jumpcloud)
[![GoDoc](https://godoc.org/github.com/jumpcloud/terraform-provider-jumpcloud?status.svg)](https://godoc.org/github.com/jumpcloud/terraform-provider-jumpcloud)
[![Release](https://img.shields.io/github/release/jumpcloud/terraform-provider-jumpcloud.svg)](https://github.com/jumpcloud/terraform-provider-jumpcloud/releases)
[![License](https://img.shields.io/github/license/jumpcloud/terraform-provider-jumpcloud.svg)](https://github.com/jumpcloud/terraform-provider-jumpcloud/blob/master/LICENSE)

This provider enables Terraform to manage JumpCloud resources.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 0.12.x
- [Go](https://golang.org/doc/install) >= 1.16 (to build the provider plugin)

## Using the Provider

To use the provider, add the following terraform block to your configuration to specify the required provider:

```hcl
terraform {
  required_providers {
    jumpcloud = {
      source  = "registry.terraform.io/agilize/jumpcloud"
      version = "~> 1.0"
    }
  }
}

provider "jumpcloud" {
  api_key = var.jumpcloud_api_key # Or use JUMPCLOUD_API_KEY env var
  org_id  = var.jumpcloud_org_id  # Optional: Or use JUMPCLOUD_ORG_ID env var
}
```

### Authentication

The provider supports the following authentication methods:

1. Static credentials: Set the `api_key` (required) and `org_id` (optional) values in the provider block.
2. Environment variables:
   - `JUMPCLOUD_API_KEY`: API key for JumpCloud operations.
   - `JUMPCLOUD_ORG_ID`: Organization ID for multi-tenant environments.

## Example: Managing Users and Groups

```hcl
# Create a user
resource "jumpcloud_user" "example" {
  username  = "johndoe"
  email     = "john.doe@example.com"
  firstname = "John"
  lastname  = "Doe"
  password  = "securePassword123!"
}

# Create a user group
resource "jumpcloud_user_group" "engineering" {
  name        = "Engineering Team"
  description = "Group for engineering staff"
}

# Add the user to the group
resource "jumpcloud_user_group_membership" "example_membership" {
  user_group_id = jumpcloud_user_group.engineering.id
  user_id       = jumpcloud_user.example.id
}
```

## Example: Authentication Policies

```hcl
# Create an authentication policy
resource "jumpcloud_auth_policy" "secure_policy" {
  name        = "Secure Access Policy"
  description = "Requires MFA for all users"
  
  rule {
    type = "AUTHENTICATION"
    
    conditions {
      resource {
        type = "USER_GROUP"
        id   = jumpcloud_user_group.engineering.id
      }
    }
    
    effects {
      allow_ssh_password_authentication    = false
      allow_multi_factor_authentication    = true
      force_multi_factor_authentication    = true
      require_password_reset               = false
      allow_password_management_self_serve = true
    }
  }
}
```

## Documentation

Comprehensive documentation for each module is available in their respective directories:

- [Authentication](jumpcloud/authentication/README.md)
- [App Catalog](jumpcloud/app_catalog/README.md)
- [Admin](jumpcloud/admin/README.md)
- [IP List](jumpcloud/iplist/README.md)
- [Password Policies](jumpcloud/password_policies/README.md)
- [RADIUS](jumpcloud/radius/README.md)
- [SCIM](jumpcloud/scim/README.md)
- [System Groups](jumpcloud/system_groups/README.md)
- [User Associations](jumpcloud/user_associations/README.md)
- [User Groups](jumpcloud/user_groups/README.md)

## Development

### Building the Provider

Clone the repository:

```bash
git clone https://github.com/jumpcloud/terraform-provider-jumpcloud.git
```

Build the provider:

```bash
cd terraform-provider-jumpcloud
go build
```

### Testing

To run the tests, you will need:

- A JumpCloud API key
- Go installed on your machine

Set the environment variable:

```bash
export JUMPCLOUD_API_KEY="your-api-key"
export JUMPCLOUD_ORG_ID="your-org-id"  # Optional
export TF_ACC=1  # For acceptance tests
```

Run the tests:

```bash
go test ./...
```

For acceptance tests:

```bash
go test ./... -v -run=TestAcc
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This provider is distributed under the [Apache License, Version 2.0](LICENSE). 