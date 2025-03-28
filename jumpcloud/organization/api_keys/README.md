# JumpCloud API Key Management

This package contains resources and data sources for managing JumpCloud API Keys.

## Resources

### `jumpcloud_api_key`

This resource allows you to create and manage JumpCloud API keys.

#### Example Usage

```hcl
resource "jumpcloud_api_key" "example" {
  name        = "terraform-api-key"
  description = "API key for Terraform automation"
  expires     = "2023-12-31T23:59:59Z"  # Optional expiration date
}

# The sensitive API key value is available as an output
output "api_key_value" {
  value     = jumpcloud_api_key.example.key
  sensitive = true
}
```

### `jumpcloud_api_key_binding`

This resource allows you to create and manage API key bindings to control access to JumpCloud resources.

#### Example Usage

```hcl
resource "jumpcloud_api_key" "example" {
  name        = "terraform-api-key"
  description = "API key for Terraform automation"
}

# Grant read and write permissions to all users
resource "jumpcloud_api_key_binding" "users_binding" {
  api_key_id     = jumpcloud_api_key.example.id
  resource_type  = "user"
  permissions    = ["read", "write"]
}

# Grant read-only permissions to specific systems
resource "jumpcloud_api_key_binding" "systems_binding" {
  api_key_id     = jumpcloud_api_key.example.id
  resource_type  = "system"
  resource_ids   = [
    "system-id-1",
    "system-id-2"
  ]
  permissions    = ["read"]
}
```

## Resource Types

The following resource types are supported for API key bindings:

- `user` - JumpCloud users
- `system` - JumpCloud systems
- `group` - User and system groups
- `application` - JumpCloud applications
- `policy` - JumpCloud policies
- `command` - JumpCloud commands
- `organization` - JumpCloud organizations
- `radius_server` - RADIUS servers
- `directory` - JumpCloud directories
- `webhook` - JumpCloud webhooks

## Permission Types

The following permission types are supported for API key bindings:

- `read` - Read-only access
- `write` - Create and update access
- `delete` - Delete access
- `manage` - Full access (includes read, write, and delete)

## API Reference

For more information about the JumpCloud API for API keys, please refer to the official documentation:

- [API Keys API](https://docs.jumpcloud.com/api/1.0/index.html#api-keys)
- [API Key Bindings API](https://docs.jumpcloud.com/api/1.0/index.html#api-key-bindings) 