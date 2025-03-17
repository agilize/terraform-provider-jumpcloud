# jumpcloud_mfa_settings Resource

Manages MFA (Multi-Factor Authentication) settings in JumpCloud. Since these settings are defined per organization, this is a singleton resource - only one instance should exist per JumpCloud organization.

## Example Usage

### Basic MFA Configuration

```hcl
resource "jumpcloud_mfa_settings" "corporate_mfa" {
  # Enable System Insights-based MFA
  system_insights_enrolled = true
  
  # Configure exclusion window days (grace period)
  exclusion_window_days = 7
  
  # Allowed MFA methods
  enabled_methods = [
    "totp",      # Time-based One-Time Password (Google Authenticator, etc.)
    "push",      # Push notifications
    "webauthn"   # FIDO2/WebAuthn (Yubikey, etc.)
  ]
}
```

### MFA Configuration in a Multi-tenant Environment

```hcl
resource "jumpcloud_mfa_settings" "child_org_mfa" {
  # Specific organization ID (for multi-tenant implementations)
  organization_id = var.child_organization_id
  
  system_insights_enrolled = true
  
  # More restrictive configuration - TOTP only
  enabled_methods = ["totp"]
  
  # No exclusion window - immediate enforcement
  exclusion_window_days = 0
}
```

## Argument Reference

The following arguments are supported:

* `system_insights_enrolled` - (Optional) Whether System Insights is enabled for MFA. Default: `false`.
* `exclusion_window_days` - (Optional) Number of days for the MFA exclusion window. This is a "grace period" during which users can access without MFA after the configuration is activated. Values from 0 to 30, where 0 means immediate enforcement. Default: `0`.
* `enabled_methods` - (Optional) List of enabled MFA methods. Valid values: `totp`, `duo`, `push`, `sms`, `email`, `webauthn`, `security_questions`.
* `organization_id` - (Optional) Organization ID for multi-tenant implementations. If not specified, the current organization ID configured in the provider will be used.

## Attribute Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - ID of the MFA settings or "current" for the current organization.
* `updated` - Date when the MFA settings were last updated.

## Import

JumpCloud MFA settings can be imported using the organization ID or "current" if there's only one:

```
terraform import jumpcloud_mfa_settings.example {organization_id}
```

or

```
terraform import jumpcloud_mfa_settings.example current
```

## Implementation Notes

This resource manages a singleton per JumpCloud organization. The deletion behavior (`terraform destroy`) resets the MFA settings to JumpCloud default values instead of deleting them completely, as MFA settings always exist for each organization. 