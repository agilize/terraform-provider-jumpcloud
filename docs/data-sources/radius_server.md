---
page_title: "JumpCloud: jumpcloud_radius_server"
subcategory: "User Authentication"
description: |-
  Get information about a RADIUS server in JumpCloud
---

# jumpcloud_radius_server (Data Source)

This data source allows you to get information about a specific RADIUS server configured in JumpCloud. It can be useful for referencing existing RADIUS servers without needing to recreate them in your Terraform code.

## Example Usage

### Get by ID

```hcl
data "jumpcloud_radius_server" "vpn_server" {
  id = "5f8d3e9c9d1d8b6a8c4c2f7g"
}

output "vpn_server_mfa_required" {
  value = data.jumpcloud_radius_server.vpn_server.mfa_required
}
```

### Get by Name

```hcl
data "jumpcloud_radius_server" "wifi_auth" {
  name = "WiFi Authentication"
}

# Use the ID in another resource or association
resource "jumpcloud_user_group" "wifi_users" {
  name        = "WiFi Users"
  description = "Users with access to authenticated WiFi network"
}

resource "jumpcloud_radius_server" "new_wifi_radius" {
  name          = "New WiFi Authentication"
  shared_secret = var.radius_secret

  # Associate with the same group as the existing server
  targets = [
    jumpcloud_user_group.wifi_users.id
  ]
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) The ID of the RADIUS server in JumpCloud. Conflicts with `name`.
* `name` - (Optional) The name of the RADIUS server in JumpCloud. Conflicts with `id`.

**Note**: Exactly one of `id` or `name` must be specified.

## Attribute Reference

The following attributes are exported:

* `network_source_ip` - Network source IP used for communication with the RADIUS server.
* `mfa_required` - Whether multi-factor authentication is required for the RADIUS server.
* `user_password_expiration_action` - Action to take when a user's password expires (`allow` or `deny`).
* `user_lockout_action` - Action to take when a user is locked out (`allow` or `deny`).
* `user_attribute` - User attribute used for authentication (`username` or `email`).
* `targets` - List of group IDs associated with the RADIUS server.
* `created` - Creation date of the RADIUS server.
* `updated` - Last update date of the RADIUS server.

**Note**: The `shared_secret` attribute is not exported for security reasons.