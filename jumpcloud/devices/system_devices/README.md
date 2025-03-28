# JumpCloud Devices Module

This module provides resources for managing JumpCloud devices (devices).

## Resources

### jumpcloud_device

The `jumpcloud_device` resource allows you to manage a JumpCloud device.

#### Example Usage

```hcl
resource "jumpcloud_device" "example" {
  display_name                       = "example-device"
  description                        = "Example device managed by Terraform"
  allow_ssh_root_login               = false
  allow_ssh_password_authentication  = true
  allow_multi_factor_authentication  = true
  tags                               = ["terraform", "example"]
  
  attributes = {
    location = "Remote"
    department = "Engineering"
  }
}
```

#### Argument Reference

The following arguments are supported:

* `display_name` - (Required) The name to display for the device.
* `description` - (Optional) A description of the device.
* `allow_ssh_root_login` - (Optional) Whether to allow SSH root login. Defaults to `false`.
* `allow_ssh_password_authentication` - (Optional) Whether to allow SSH password authentication. Defaults to `true`.
* `allow_multi_factor_authentication` - (Optional) Whether to allow multi-factor authentication. Defaults to `false`.
* `tags` - (Optional) A list of tags to apply to the device.
* `attributes` - (Optional) A map of custom attributes for the device.

#### Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the device.
* `system_type` - The type of device.
* `os` - The operating device of the device.
* `version` - The version of the operating device.
* `agent_version` - The version of the JumpCloud agent.
* `created` - The timestamp when the device was created.
* `updated` - The timestamp when the device was last updated.
* `last_contact` - The timestamp when the device last contacted JumpCloud.
* `remote_ip` - The remote IP of the device.
* `active` - Whether the device is active.
* `has_active_agent` - Whether the device has an active agent.
* `mdm_managed` - Whether the device is managed by MDM.
* `enrollment_status` - The enrollment status of the device.
* `hostname` - The hostname of the device.
* `serial_number` - The serial number of the device.

## Data Sources

### jumpcloud_device

The `jumpcloud_device` data source allows you to retrieve information about a JumpCloud device.

#### Example Usage

```hcl
# Get device by ID
data "jumpcloud_device" "example" {
  system_id = "5f8d3f5c9d5abe5214e0812a"
}

# Get device by display name
data "jumpcloud_device" "by_name" {
  display_name = "example-device"
}
```

#### Argument Reference

One of the following arguments must be provided:

* `system_id` - (Optional) The ID of the device to retrieve.
* `display_name` - (Optional) The display name of the device to retrieve.

#### Attribute Reference

The same attributes are available as for the `jumpcloud_device` resource. 