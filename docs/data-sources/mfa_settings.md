---
page_title: "JumpCloud: jumpcloud_mfa_settings"
subcategory: "Security Management"
description: |-
  Get information about MFA settings in JumpCloud
---

# jumpcloud_mfa_settings (Data Source)

This data source allows you to get information about MFA (Multi-Factor Authentication) settings for a JumpCloud organization. It can be useful for evaluating existing MFA policies before making changes or for monitoring settings across multiple organizations.

## Example Usage

### Get settings for the current organization

```hcl
data "jumpcloud_mfa_settings" "current" {}

output "mfa_methods_enabled" {
  value = data.jumpcloud_mfa_settings.current.enabled_methods
}

output "system_insights_status" {
  value = data.jumpcloud_mfa_settings.current.system_insights_enrolled ? "Enabled" : "Disabled"
}
```

### Get settings for a specific organization (multi-tenant)

```hcl
data "jumpcloud_mfa_settings" "child_org" {
  organization_id = var.child_organization_id
}

# Validate settings for compliance
output "mfa_exclusion_window_compliant" {
  value = data.jumpcloud_mfa_settings.child_org.exclusion_window_days <= 7
  description = "Compliant if exclusion window is less than or equal to 7 days"
}
```

## Argument Reference

The following arguments are supported:

* `organization_id` - (Optional) Organization ID to get MFA settings for. If not specified, the current organization configured in the provider will be used.

## Attribute Reference

The following attributes are exported:

* `id` - ID of the MFA settings or "current" for the current organization.
* `system_insights_enrolled` - Whether System Insights is enabled for MFA.
* `exclusion_window_days` - Number of exclusion window days for MFA (grace period).
* `enabled_methods` - List of enabled MFA methods, such as `totp`, `duo`, `push`, `sms`, `email`, `webauthn`, and `security_questions`.
* `updated` - Last update date of the MFA settings.