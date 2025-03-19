# jumpcloud_policy_association Resource

This resource allows you to associate JumpCloud policies with user or system groups, applying the security and compliance configurations defined in the policies to the group members.

## Example Usage

```hcl
# Create a password complexity policy
resource "jumpcloud_policy" "password_complexity" {
  name        = "Secure Password Policy"
  description = "Password complexity policy for the finance department"
  type        = "password_complexity"
  active      = true
  
  configurations = {
    min_length             = "12"
    requires_uppercase     = "true"
    requires_lowercase     = "true"
    requires_number        = "true"
    requires_special_char  = "true"
  }
}

# Create a user group
resource "jumpcloud_user_group" "finance" {
  name        = "Finance Department"
  description = "Finance department user group"
}

# Associate the policy with the user group
resource "jumpcloud_policy_association" "finance_password_policy" {
  policy_id = jumpcloud_policy.password_complexity.id
  group_id  = jumpcloud_user_group.finance.id
  type      = "user_group"
}

# Associate policy with a system group
resource "jumpcloud_system_group" "servers" {
  name        = "Production Servers"
  description = "Production servers group"
}

resource "jumpcloud_policy" "system_updates" {
  name        = "System Updates Policy"
  description = "Policy for system update control"
  type        = "system_updates"
  active      = true
  
  configurations = {
    auto_update_enabled = "true"
    auto_update_time    = "02:00"
  }
}

resource "jumpcloud_policy_association" "servers_update_policy" {
  policy_id = jumpcloud_policy.system_updates.id
  group_id  = jumpcloud_system_group.servers.id
  type      = "system_group"
}
```

## Argument Reference

The following arguments are supported:

* `policy_id` - (Required) The ID of the policy to be associated.
* `group_id` - (Required) The ID of the group (user or system) to which the policy will be associated.
* `type` - (Required) The type of group. Valid values: `user_group` or `system_group`.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

* `id` - The ID of the policy to group association, in the format `policy_id:group_id:type`.

## Import

Policy associations can be imported using the composite ID in the format `policy_id:group_id:type`:

```
terraform import jumpcloud_policy_association.example 5f0c1b2c3d4e5f6g7h8i9j0k:6a7b8c9d0e1f2g3h4i5j6k7l:user_group
```

## Common Use Cases

### Applying Multiple Policies

```hcl
# MFA policy for all employees
resource "jumpcloud_policy" "mfa_policy" {
  name        = "Required MFA Policy"
  description = "Global MFA policy for all users"
  type        = "mfa"
  active      = true
  
  configurations = {
    allow_totp_enrollment      = "true"
    require_mfa_for_all_users  = "true"
  }
}

# Associate MFA policy with multiple groups
resource "jumpcloud_user_group" "it" {
  name = "IT Department"
}

resource "jumpcloud_user_group" "executives" {
  name = "Executive Team"
}

resource "jumpcloud_policy_association" "it_mfa" {
  policy_id = jumpcloud_policy.mfa_policy.id
  group_id  = jumpcloud_user_group.it.id
  type      = "user_group"
}

resource "jumpcloud_policy_association" "executives_mfa" {
  policy_id = jumpcloud_policy.mfa_policy.id
  group_id  = jumpcloud_user_group.executives.id
  type      = "user_group"
}
```

### Conditional Management

```hcl
# Check if the policy is already associated with the group
data "jumpcloud_user_group" "existing_group" {
  name = "Finance Department"
}

data "jumpcloud_policy" "existing_policy" {
  name = "Secure Password Policy"
}

# Fetch all associated policies (fictional example)
data "jumpcloud_policy_associations" "existing_associations" {
  group_id = data.jumpcloud_user_group.existing_group.id
  type     = "user_group"
}

locals {
  # Check if the policy is already associated
  policy_already_associated = contains(data.jumpcloud_policy_associations.existing_associations.policy_ids, data.jumpcloud_policy.existing_policy.id)
}

# Create association only if it doesn't exist
resource "jumpcloud_policy_association" "conditional_association" {
  count = local.policy_already_associated ? 0 : 1
  
  policy_id = data.jumpcloud_policy.existing_policy.id
  group_id  = data.jumpcloud_user_group.existing_group.id
  type      = "user_group"
}
```

## Security Considerations

* Associate critical security policies with specific groups to ensure that only appropriate users are affected.
* When associating MFA or complex password policies, consider the impact on user experience and prepare appropriate communications.
* For policies that affect systems, verify compatibility before applying them in production environments.
* Consider implementing policies in phases, starting with smaller groups before expanding to the entire organization.
* Implement a regular review process for policy associations to ensure they remain appropriate for security needs. 