# jumpcloud_user Data Source

Use this data source to get information about a JumpCloud user.

## Example Usage

```hcl
# Get a user by username
data "jumpcloud_user" "by_username" {
  username = "example.user"
}

# Get a user by email
data "jumpcloud_user" "by_email" {
  email = "example.user@example.com"
}

# Get a user by ID
data "jumpcloud_user" "by_id" {
  user_id = "5f0c1b2c3d4e5f6g7h8i9j0k"
}

output "user_details" {
  value = data.jumpcloud_user.by_username.attributes
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Optional) The username of the user to retrieve. Conflicts with `user_id` and `email`.
* `email` - (Optional) The email of the user to retrieve. Conflicts with `user_id` and `username`.
* `user_id` - (Optional) The ID of the user to retrieve. Conflicts with `username` and `email`.

**Note:** Exactly one of these arguments must be provided.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the user.
* `firstname` - The first name of the user.
* `lastname` - The last name of the user.
* `middlename` - The middle name of the user.
* `displayname` - The display name of the user.
* `description` - The description of the user.
* `attributes` - A map of custom attributes for the user.
* `mfa_enabled` - Whether MFA is enabled for the user.
* `password_never_expires` - Whether the password never expires.
* `alternate_email` - An alternate email address for the user.
* `company` - The company the user belongs to.
* `cost_center` - The cost center the user is associated with.
* `department` - The department the user belongs to.
* `employee_identifier` - An identifier for the employee.
* `employee_type` - The type of employee.
* `job_title` - The job title of the user.
* `location` - The location of the user.
* `enable_managed_uid` - Whether managed UID is enabled for the user.
* `enable_user_portal_multifactor` - Whether multifactor authentication is enabled for the user portal.
* `externally_managed` - Whether the user is externally managed.
* `ldap_binding_user` - Whether the user is an LDAP binding user.
* `passwordless_sudo` - Whether passwordless sudo is enabled for the user.
* `global_passwordless_sudo` - Whether global passwordless sudo is enabled for the user.
* `public_key` - The public SSH key for the user.
* `allow_public_key` - Whether public key authentication is allowed for the user.
* `samba_service_user` - Whether the user is a Samba service user.
* `sudo` - Whether sudo access is granted to the user.
* `suspended` - Whether the user is suspended.
* `unix_uid` - The Unix UID for the user.
* `unix_guid` - The Unix GUID for the user.
* `disable_device_max_login_attempts` - Whether maximum login attempts are disabled for the user's devices.
* `password_recovery_email` - The email address used for password recovery.
* `enforce_uid_gid_consistency` - Whether UID/GID consistency is enforced.
* `delegated_authority` - The delegated authority for the user.
* `password_authority` - The password authority for the user.
* `managed_apple_id` - The managed Apple ID for the user.
* `bypass_managed_device_lockout` - Whether managed device lockout is bypassed for the user.
* `manager_id` - The ID of the user's manager in JumpCloud.
* `created` - The date the user was created.