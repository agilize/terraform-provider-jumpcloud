# JumpCloud MFA Domain

This directory contains resources and data sources for managing Multi-Factor Authentication (MFA) in JumpCloud.

## Resources

- `jumpcloud_mfa_configuration` - Manages MFA configuration settings
- `jumpcloud_mfa_settings` - Manages MFA global settings

## Data Sources

- `jumpcloud_mfa_settings` - Retrieves MFA settings information
- `jumpcloud_mfa_stats` - Provides statistics about MFA usage across the organization

## Example Usage

### MFA Settings

```hcl
resource "jumpcloud_mfa_settings" "example" {
  system_insights_enrolled = true
  exclusion_window_days    = 7
  enabled_methods          = ["totp", "push", "sms"]
}
```

### MFA Configuration

```hcl
resource "jumpcloud_mfa_configuration" "example" {
  enabled            = true
  exclusive_enabled  = false
  system_mfa_required = true
  user_portal_mfa    = true
  admin_console_mfa  = true
  totp_enabled       = true
  push_enabled       = true
  default_mfa_type   = "totp"
}
```

### MFA Settings Data Source

```hcl
data "jumpcloud_mfa_settings" "current" {
}

output "current_mfa_methods" {
  value = data.jumpcloud_mfa_settings.current.enabled_methods
}
```

### MFA Stats Data Source

```hcl
data "jumpcloud_mfa_stats" "current" {
  # Optional: Specify a date range for analysis
  # start_date = "2023-01-01T00:00:00Z"
  # end_date   = "2023-12-31T23:59:59Z"
}

output "mfa_enrollment_rate" {
  value = data.jumpcloud_mfa_stats.current.mfa_enrollment_rate
}

output "popular_mfa_methods" {
  value = data.jumpcloud_mfa_stats.current.method_stats
}
``` 