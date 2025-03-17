# jumpcloud_radius_server Resource

This resource allows you to manage RADIUS servers in JumpCloud. RADIUS (Remote Authentication Dial-In User Service) is a network protocol that provides centralized authentication, authorization, and accounting for users connecting to and using network services.

## Example Usage

### Basic RADIUS Server Configuration

```hcl
resource "jumpcloud_radius_server" "corporate_vpn" {
  name          = "Corporate VPN"
  shared_secret = "s3cur3-sh4r3d-s3cr3t"
  mfa_required  = true
  
  # Source IP for communication with the RADIUS server
  network_source_ip = "10.0.1.5"
  
  # Authentication settings
  user_attribute = "username"
  user_password_expiration_action = "deny"
  user_lockout_action = "deny"
}
```

### RADIUS Server Associated with User Groups

```hcl
resource "jumpcloud_radius_server" "wifi_auth" {
  name          = "WiFi Authentication"
  shared_secret = var.radius_secret
  mfa_required  = false
  
  # Configure authentication by email instead of username
  user_attribute = "email"
  
  # Associate with specific groups
  targets = [
    jumpcloud_user_group.employees.id,
    jumpcloud_user_group.contractors.id
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the RADIUS server.
* `shared_secret` - (Required) Shared secret used for authentication between the client and RADIUS server. This value is sensitive and will not be displayed in Terraform output.
* `network_source_ip` - (Optional) Source network IP that will be used to communicate with the RADIUS server.
* `mfa_required` - (Optional) Whether multi-factor authentication is required for the RADIUS server. Default: `false`.
* `user_password_expiration_action` - (Optional) Action to take when a user's password expires. Valid values: `allow` or `deny`. Default: `allow`.
* `user_lockout_action` - (Optional) Action to take when a user is locked out. Valid values: `allow` or `deny`. Default: `deny`.
* `user_attribute` - (Optional) User attribute used for authentication. Valid values: `username` or `email`. Default: `username`.
* `targets` - (Optional) List of user group IDs associated with the RADIUS server. If not specified, the server will be available to all users.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - ID of the RADIUS server.
* `created` - Creation date of the RADIUS server.
* `updated` - Last update date of the RADIUS server.

## Import

JumpCloud RADIUS servers can be imported using the server ID:

```
terraform import jumpcloud_radius_server.example {radius_server_id}
``` 