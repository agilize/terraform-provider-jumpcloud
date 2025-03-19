# jumpcloud_policy Resource

This resource allows you to manage policies in JumpCloud. Policies are configurations that can be applied to users or systems, controlling different security aspects such as password complexity, MFA, account lockout, and system updates.

## Example Usage

```hcl
# Password complexity policy example
resource "jumpcloud_policy" "password_complexity" {
  name        = "Secure Password Policy"
  description = "Secure password complexity policy for all users"
  type        = "password_complexity"
  active      = true
  
  # Specific settings for password complexity policy
  configurations = {
    min_length             = "12"         # Minimum password length
    requires_uppercase     = "true"       # Requires uppercase letters
    requires_lowercase     = "true"       # Requires lowercase letters
    requires_number        = "true"       # Requires numbers
    requires_special_char  = "true"       # Requires special characters
    password_expires_days  = "90"         # Password expires in 90 days
    enable_password_expiry = "true"       # Enable password expiration
  }
}

# MFA (Multi-Factor Authentication) policy example
resource "jumpcloud_policy" "mfa_policy" {
  name        = "Required MFA Policy"
  description = "Policy requiring MFA for all users"
  type        = "mfa"
  active      = true
  
  configurations = {
    allow_sms_enrollment        = "true"   # Allow MFA via SMS
    allow_voice_call_enrollment = "true"   # Allow MFA via voice call
    allow_totp_enrollment       = "true"   # Allow MFA via application (TOTP)
    allow_push_notification     = "true"   # Allow MFA via push notification
    require_mfa_for_all_users   = "true"   # Require MFA for all users
  }
}

# Account lockout policy example
resource "jumpcloud_policy" "account_lockout" {
  name        = "Account Lockout Policy"
  description = "Policy for account lockout after failed login attempts"
  type        = "account_lockout_timeout"
  active      = true
  
  configurations = {
    failed_login_attempts_allowed = "5"        # Number of failed login attempts before lockout
    lockout_time_period_minutes   = "30"       # Time period for lockout in minutes
    failed_login_reset_after_mins = "15"       # Reset failed login count after minutes
  }
}

# System update policy example for Windows
resource "jumpcloud_policy" "windows_updates" {
  name        = "Windows Update Policy"
  description = "Policy managing Windows system updates"
  type        = "windows_updates"
  active      = true
  
  configurations = {
    enable_automatic_updates = "true"           # Enable automatic updates
    allow_updates_download   = "true"           # Allow updates download
    schedule_updates_day     = "Sunday"         # Schedule day for updates
    schedule_updates_time    = "01:00"          # Schedule time for updates
    notify_users_before      = "true"           # Notify users before updating
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the policy.
* `description` - (Optional) A description of the policy and its purpose.
* `type` - (Required) The type of policy. Possible values: `password_complexity`, `password_expiration`, `account_lockout_timeout`, `mfa`, `system_updates`, `windows_updates`, `mac_updates`, and others as supported by JumpCloud.
* `active` - (Optional) Whether the policy is active or not. Default is `true`.
* `configurations` - (Required) A map of configuration settings specific to the policy type. The available settings depend on the policy type.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the policy.
* `organization_id` - The organization ID the policy belongs to.
* `created` - The timestamp when the policy was created.
* `updated` - The timestamp when the policy was last updated.

## Import

Policies can be imported using their ID:

```shell
terraform import jumpcloud_policy.password_complexity 5f8b0e1b9d81b81b33c92a1c
```

## Policy Type Reference

The following policy types are supported by JumpCloud:

* `password_complexity` - Password complexity requirements
* `password_expiration` - Password expiration settings
* `account_lockout_timeout` - Account lockout after failed login attempts
* `mfa` - Multi-factor authentication settings
* `system_updates` - General system update settings
* `windows_updates` - Windows-specific update settings
* `mac_updates` - macOS-specific update settings
* `password_reused` - Password reuse restrictions

## Best Practices

1. **Start with Templates**: Use JumpCloud's policy templates as a starting point and customize as needed.
2. **Test Before Deploying**: Test policies on a small group before rolling out to the entire organization.
3. **Document**: Add clear descriptions to policies to document their purpose and scope.
4. **Regular Reviews**: Periodically review and update policies to ensure they meet current security requirements. 