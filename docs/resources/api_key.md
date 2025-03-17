# jumpcloud_api_key

Manages API keys in JumpCloud. This resource allows you to create and manage API keys that can be used to authenticate requests to JumpCloud APIs.

## Example Usage

### API Key for Automation
```hcl
# Create an API key for automation
resource "jumpcloud_api_key" "automation" {
  name        = "Automation API Key"
  description = "API key for process automation"
  expires     = timeadd(timestamp(), "8760h") # Expires in 1 year
}

# Configure permissions for the key
resource "jumpcloud_api_key_binding" "automation_user_management" {
  api_key_id    = jumpcloud_api_key.automation.id
  resource_type = "user"
  permissions   = ["read", "list", "create", "update"]
}

# Export the key securely
output "automation_api_key" {
  value     = jumpcloud_api_key.automation.key
  sensitive = true
}
```

### Temporary API Key
```hcl
# Create a temporary API key for a project
resource "jumpcloud_api_key" "temp_project" {
  name        = "Temporary Project Key"
  description = "Temporary key for migration project"
  expires     = timeadd(timestamp(), "168h") # Expires in 1 week
}

# Configure minimal required permissions
resource "jumpcloud_api_key_binding" "temp_project_access" {
  api_key_id    = jumpcloud_api_key.temp_project.id
  resource_type = "system"
  permissions   = ["read", "list"]
}
```

### Read-Only API Key
```hcl
# Create a read-only API key for reporting
resource "jumpcloud_api_key" "reporting" {
  name        = "Reporting API Key"
  description = "Read-only key for reporting systems"
  expires     = timeadd(timestamp(), "4320h") # Expires in 6 months
}

# Set read-only permissions for users
resource "jumpcloud_api_key_binding" "reporting_users" {
  api_key_id    = jumpcloud_api_key.reporting.id
  resource_type = "user"
  permissions   = ["read", "list"]
}

# Set read-only permissions for systems
resource "jumpcloud_api_key_binding" "reporting_systems" {
  api_key_id    = jumpcloud_api_key.reporting.id
  resource_type = "system"
  permissions   = ["read", "list"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the API key. Must be unique within the organization.
* `description` - (Optional) Description of the API key's purpose.
* `expires` - (Optional) The timestamp when the API key will expire. If not specified, the key will not expire.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the API key.
* `key` - The API key value. This is only returned once after creation and cannot be retrieved again.
* `created` - The timestamp when the API key was created.
* `organization_id` - The organization ID the API key belongs to.

## Import

API keys can be imported using their ID:

```shell
terraform import jumpcloud_api_key.automation 5f1b881dc9e9a9b7e8d6c5a4
```

## Security Considerations

1. **Expiration**: Always set an expiration date for API keys. Long-lived keys pose a security risk.
2. **Least Privilege**: Only grant the minimum permissions required for the key's purpose.
3. **Secure Storage**: Store API key values securely using Terraform's sensitive output handling or a secrets manager.
4. **Key Rotation**: Implement a regular rotation schedule for API keys.
5. **Monitoring**: Set up monitoring to detect unusual activities with API keys.

## Permission Management

API key permissions are managed through the `jumpcloud_api_key_binding` resource. Each binding defines the permissions for a specific resource type.

```hcl
resource "jumpcloud_api_key_binding" "example" {
  api_key_id    = jumpcloud_api_key.example.id
  resource_type = "user"
  permissions   = ["read", "list", "create", "update", "delete"]
}
```

### Available Resource Types

* `user` - User management operations
* `system` - System management operations
* `user_group` - User group operations
* `system_group` - System group operations 
* `application` - Application management
* `policy` - Policy management
* `command` - Command operations
* `directory` - Directory operations
* `organization` - Organization settings

### Available Permissions

* `read` - Ability to read individual resources
* `list` - Ability to list resources
* `create` - Ability to create new resources
* `update` - Ability to update existing resources
* `delete` - Ability to delete resources

## Best Practices

1. **Naming Convention**: Use descriptive names that indicate the key's purpose and owner.
2. **Documentation**: Use the description field to document who created the key, its purpose, and when it should be rotated.
3. **Separation of Duties**: Create different keys for different functionalities or systems.
4. **Emergency Process**: Have a process for emergency revocation if a key is compromised. 