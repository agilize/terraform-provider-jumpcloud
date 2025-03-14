# jumpcloud_system Data Source

Use this data source to get information about a JumpCloud system (device).

## Example Usage

```hcl
# Get a system by ID
data "jumpcloud_system" "by_id" {
  system_id = "5f0c1b2c3d4e5f6g7h8i9j0k"
}

# Get a system by display name
data "jumpcloud_system" "by_name" {
  display_name = "web-server-01"
}

output "system_os" {
  value = data.jumpcloud_system.by_name.os
}

output "system_version" {
  value = data.jumpcloud_system.by_name.version
}
```

## Argument Reference

The following arguments are supported:

* `system_id` - (Optional) The ID of the system to retrieve.
* `display_name` - (Optional) The display name of the system to retrieve.

**Note:** Exactly one of these arguments must be provided.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the system.
* `system_type` - The type of the system.
* `os` - The operating system of the system.
* `version` - The OS version of the system.
* `agent_version` - The version of the JumpCloud agent installed on the system.
* `allow_ssh_root_login` - Whether SSH root login is allowed.
* `allow_ssh_password_authentication` - Whether SSH password authentication is allowed.
* `allow_multi_factor_authentication` - Whether multi-factor authentication is allowed.
* `tags` - A list of tags associated with the system.
* `description` - The description of the system.
* `attributes` - A map of attributes associated with the system.
* `agent_bound` - Whether the system is bound to a JumpCloud agent.
* `ssh_root_enabled` - Whether SSH root login is enabled for this system.
* `organization_id` - The organization ID the system belongs to.
* `created` - The date the system was created.
* `updated` - The date the system was last updated. 