# Terraform Provider JumpCloud

[![Build Status](https://github.com/ferreirafa/terraform-provider-jumpcloud/workflows/Unified%20Build%20and%20Release%20Pipeline/badge.svg)](https://github.com/ferreirafa/terraform-provider-jumpcloud/actions)
[![Latest Release](https://img.shields.io/github/v/release/ferreirafa/terraform-provider-jumpcloud?include_prereleases&sort=semver)](https://github.com/ferreirafa/terraform-provider-jumpcloud/releases)
[![GitHub Packages](https://img.shields.io/badge/GitHub%20Packages-Provider-blue)](https://github.com/ferreirafa/terraform-provider-jumpcloud/pkgs/container/terraform-provider-jumpcloud)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The JumpCloud Terraform Provider allows you to manage resources on the [JumpCloud](https://jumpcloud.com) platform through Terraform.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.20 (for development)
- [JumpCloud API Key](https://jumpcloud.com/support/api-key)

## Usage

### Provider Installation

#### Option 1: Using GitHub Container Registry (Recommended)

Configure the GitHub Container Registry as a source for the provider. Add the following to your `~/.terraformrc` file:

```hcl
provider_installation {
  network_mirror {
    url = "https://ghcr.io/ferreirafav/terraform-provider-jumpcloud"
    include = ["ghcr.io/ferreirafav/jumpcloud"]
  }
  direct {
    exclude = ["ghcr.io/ferreirafav/jumpcloud"]
  }
}
```

Then, in your Terraform configuration:

```hcl
terraform {
  required_providers {
    jumpcloud = {
      source  = "ghcr.io/ferreirafav/jumpcloud"
      version = "~> 0.1.0"
    }
  }
}

provider "jumpcloud" {
  api_key = "your-api-key"  # Can also be set via JUMPCLOUD_API_KEY environment variable
}
```

For detailed instructions, see [Using the Provider via GHCR](docs/ghcr-usage.md).

#### Option 2: Using GitHub Releases

To use the provider from GitHub Releases, add the following block to your Terraform configuration file:

```hcl
terraform {
  required_providers {
    jumpcloud = {
      source  = "ferreirafa/jumpcloud"
      version = "~> 0.1.0"
    }
  }
}

provider "jumpcloud" {
  api_key = "your-api-key"  # Can also be set via JUMPCLOUD_API_KEY environment variable
}
```

### Resource Management

#### Users

```hcl
resource "jumpcloud_user" "example" {
  username = "john.doe"
  email    = "john.doe@example.com"
  firstname = "John"
  lastname  = "Doe"
  
  password = "securePassword123!"
  
  tags = ["dev", "engineering"]
}
```

#### User Groups

```hcl
resource "jumpcloud_user_group" "engineering" {
  name = "Engineering Team"
  description = "Group for engineering team members"
}

resource "jumpcloud_user_group_membership" "john_engineering" {
  user_id       = jumpcloud_user.example.id
  user_group_id = jumpcloud_user_group.engineering.id
}
```

#### Systems

```hcl
data "jumpcloud_system" "laptop" {
  display_name = "MacBook-Pro"
}

resource "jumpcloud_system_group" "dev_laptops" {
  name = "Development Laptops"
  description = "Group for development team laptops"
}

resource "jumpcloud_system_group_membership" "laptop_dev" {
  system_id       = data.jumpcloud_system.laptop.id
  system_group_id = jumpcloud_system_group.dev_laptops.id
}
```

#### Commands

```hcl
resource "jumpcloud_command" "update_packages" {
  name = "Update Packages"
  command = "apt-get update && apt-get upgrade -y"
  user = "root"
  
  system_ids = [
    data.jumpcloud_system.laptop.id
  ]
  
  trigger = "manual"
}
```

## Available Resources

| Resource Type | Description |
|---------------|-------------|
| `jumpcloud_user` | Manages JumpCloud users |
| `jumpcloud_user_group` | Manages JumpCloud user groups |
| `jumpcloud_user_group_membership` | Manages user membership in groups |
| `jumpcloud_system_group` | Manages system groups |
| `jumpcloud_system_group_membership` | Manages system membership in groups |
| `jumpcloud_command` | Manages commands to be executed on systems |
| `jumpcloud_policy` | Manages policies |
| `jumpcloud_policy_association` | Manages policy associations |
| `jumpcloud_user_system_association` | Manages associations between users and systems |
| `jumpcloud_command_association` | Manages associations between commands and targets |
| `jumpcloud_webhook` | Manages webhooks |
| `jumpcloud_api_key` | Manages API keys |
| `jumpcloud_mdm_policy` | Manages mobile device management policies |
| `jumpcloud_mdm_configuration` | Manages MDM configurations |
| `jumpcloud_authentication_policy` | Manages authentication policies |
| `jumpcloud_password_policy` | Manages password policies |
| `jumpcloud_organization_settings` | Manages organization settings |
| `jumpcloud_notification_channel` | Manages notification channels |
| `jumpcloud_app_catalog_application` | Manages application catalog applications |

## Available Data Sources

| Data Source | Description |
|-------------|-------------|
| `jumpcloud_user` | Retrieves information about a JumpCloud user |
| `jumpcloud_user_group` | Retrieves information about a user group |
| `jumpcloud_system` | Retrieves information about a system |
| `jumpcloud_system_group` | Retrieves information about a system group |
| `jumpcloud_command` | Retrieves information about a command |
| `jumpcloud_policy` | Retrieves information about a policy |
| `jumpcloud_user_system_association` | Verifies if a user is associated with a system |
| `jumpcloud_organization` | Retrieves information about the JumpCloud organization |
| `jumpcloud_ip_list` | Retrieves information about an IP list |
| `jumpcloud_platform_administrator` | Retrieves information about a platform administrator |
| `jumpcloud_mdm_devices` | Retrieves information about MDM devices |

## Authentication

The provider supports the following authentication methods:

1. Static credentials in the provider configuration:
   ```hcl
   provider "jumpcloud" {
     api_key = "your-api-key"
   }
   ```

2. Environment variables:
   ```bash
   export JUMPCLOUD_API_KEY="your-api-key"
   export JUMPCLOUD_ORG_ID="your-org-id"  # Optional
   ```

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for information on developing the provider.

## Testing

Testing documentation is available in the [tests README.md](tests/README.md).

## Contributing

Contributions are welcome! Please read the contribution guidelines before submitting a pull request.

## License

This provider is distributed under the MIT License. See [LICENSE](LICENSE) for more information. 