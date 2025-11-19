---
page_title: "Provider: JumpCloud"
description: |-
  The JumpCloud provider is used to interact with resources supported by JumpCloud's Directory-as-a-Service platform.
---

# JumpCloud Provider

The JumpCloud provider is used to manage resources in [JumpCloud's Directory-as-a-Service platform](https://jumpcloud.com). The provider needs to be configured with the proper credentials before it can be used.

## Example Usage

```terraform
terraform {
  required_providers {
    jumpcloud = {
      source  = "agilize/jumpcloud"
      version = "~> 0.1.14"
    }
  }
}

# Configure the JumpCloud Provider
provider "jumpcloud" {
  api_key = "your_api_key" # or use JUMPCLOUD_API_KEY environment variable
}

# Create a JumpCloud user
resource "jumpcloud_user" "example" {
  username  = "john.doe"
  email     = "john.doe@example.com"
  firstname = "John"
  lastname  = "Doe"
  
  password = "securePassword123!"
  
  tags = ["dev", "engineering"]
}
```

## Authentication

The JumpCloud provider requires an API key to communicate with the JumpCloud API. You can retrieve your API key from the JumpCloud Admin Console under Settings > API Settings.

You can provide the API key in the following ways:

* Set the `api_key` parameter in the provider configuration
* Set the `JUMPCLOUD_API_KEY` environment variable

## Provider Arguments

The provider supports the following arguments:

* `api_key` - (Required) JumpCloud API key. This can also be specified with the `JUMPCLOUD_API_KEY` environment variable.
* `api_url` - (Optional) Custom JumpCloud API URL. Default is the standard JumpCloud API URL.
* `organization_id` - (Optional) JumpCloud Organization ID for multi-tenant operations.

## Resources and Data Sources

The JumpCloud provider offers resources and data sources organized by functional area:

### User Management

**Resources:**
* `jumpcloud_user` - Manage users
* `jumpcloud_user_group` - Manage user groups

**Data Sources:**
* `jumpcloud_user` - Get information about users
* `jumpcloud_user_group` - Get information about user groups

### Device Management

**Resources:**
* `jumpcloud_system` - Manage devices (systems)
* `jumpcloud_policy` - Manage device policies
* `jumpcloud_policy_association` - Associate policies with device groups
* `jumpcloud_user_system_association` - Associate users with devices

**Data Sources:**
* `jumpcloud_system` - Get information about devices
* `jumpcloud_policy` - Get information about policies
* `jumpcloud_user_system_association` - Check user-device associations

### Software Management

**Resources:**
* `jumpcloud_software_update_policy` - Manage software update policies

**Data Sources:**
* `jumpcloud_software_update_policies` - List software update policies

### User Authentication

**Resources:**
* `jumpcloud_radius_server` - Manage RADIUS servers
* `jumpcloud_scim_server` - Manage SCIM servers

**Data Sources:**
* `jumpcloud_radius_server` - Get information about RADIUS servers
* `jumpcloud_scim_servers` - Get information about SCIM servers

### Security Management

**Resources:**
* `jumpcloud_mfa_settings` - Manage MFA settings

**Data Sources:**
* `jumpcloud_mfa_settings` - Get MFA settings

### Organization Settings

**Resources:**
* `jumpcloud_organization` - Manage organizations
* `jumpcloud_organization_settings` - Manage organization settings
* `jumpcloud_api_key` - Manage API keys
* `jumpcloud_api_key_binding` - Manage API key permissions
* `jumpcloud_webhook` - Manage webhooks
* `jumpcloud_webhook_subscription` - Manage webhook subscriptions

**Data Sources:**
* `jumpcloud_webhook` - Get information about webhooks

### Application Management

**Resources:**
* `jumpcloud_application` - Manage applications
* `jumpcloud_application_user_mapping` - Manage application access for users
* `jumpcloud_application_group_mapping` - Manage application access for groups

**Data Sources:**
* `jumpcloud_application` - Get information about applications