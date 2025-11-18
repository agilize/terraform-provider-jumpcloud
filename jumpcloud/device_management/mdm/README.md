# JumpCloud MDM Domain

This directory contains resources and data sources for managing Mobile Device Management (MDM) in JumpCloud.

## Resources

- `jumpcloud_mdm_configuration` - Manages the global MDM configuration settings
- `jumpcloud_mdm_enrollment_profile` - Manages MDM enrollment profiles
- `jumpcloud_mdm_policy` - Manages MDM policies
- `jumpcloud_mdm_profile` - Manages MDM profiles for device configuration

## Data Sources

- `jumpcloud_mdm_devices` - Retrieves information about MDM-managed devices
- `jumpcloud_mdm_policies` - Retrieves information about MDM policies
- `jumpcloud_mdm_stats` - Provides statistics about MDM usage across the organization

## Example Usage

### MDM Configuration

```hcl
resource "jumpcloud_mdm_configuration" "example" {
  enabled                      = true
  apple_mdm_enabled            = true
  auto_enrollment_enabled      = true
  device_user_authentication   = true
  default_device_management_id = jumpcloud_user_group.mdm_managed.id
}
```

### MDM Enrollment Profile

```hcl
resource "jumpcloud_mdm_enrollment_profile" "example" {
  name                = "Corporate iOS Devices"
  description         = "Profile for corporate-owned iOS devices"
  platform            = "ios"
  enrollment_method   = "user_initiated"
  group_id            = jumpcloud_user_group.mobile_users.id
  allow_byod          = false
  require_passcode    = true
  user_authentication = true
}
```

### MDM Policy

```hcl
resource "jumpcloud_mdm_policy" "example" {
  name        = "Corporate Mobile Security Policy"
  description = "Security settings for corporate mobile devices"
  platform    = "ios"
  
  settings = jsonencode({
    passcode_required : true,
    passcode_min_length : 8,
    passcode_max_age_days : 90,
    encryption_required : true,
    allow_app_store : false,
    allow_cloud_backup : false,
    allow_camera : true,
    allow_screen_capture : false
  })
  
  scope_type = "group"
  scope_ids  = [jumpcloud_user_group.executives.id]
}
```

### MDM Devices Data Source

```hcl
data "jumpcloud_mdm_devices" "corporate" {
  filter {
    field    = "ownership"
    operator = "eq"
    value    = "corporate"
  }
}

output "corporate_device_count" {
  value = length(data.jumpcloud_mdm_devices.corporate.devices)
}
```

### MDM Stats Data Source

```hcl
data "jumpcloud_mdm_stats" "current" {
}

output "enrolled_devices_count" {
  value = data.jumpcloud_mdm_stats.current.total_devices
}

output "compliance_rate" {
  value = data.jumpcloud_mdm_stats.current.compliance_percentage
}
``` 