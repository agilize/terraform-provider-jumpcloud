# jumpcloud_policy Data Source

Use this data source to get information about a specific existing JumpCloud policy. This data source allows you to retrieve details about policies such as password complexity, MFA, and other security settings.

## Example Usage

```hcl
# Get policy by name
data "jumpcloud_policy" "password_policy" {
  name = "Secure Password Policy"
}

# Get policy by ID
data "jumpcloud_policy" "mfa_policy" {
  id = "5f8b0e1b9d81b81b33c92a1c" # Example ID (replace with a real ID)
}

# Check if policy is active
output "password_policy_status" {
  value = "${data.jumpcloud_policy.password_policy.name} is ${data.jumpcloud_policy.password_policy.active ? "active" : "inactive"}"
}

# Check policy settings
output "password_policy_min_length" {
  value = lookup(data.jumpcloud_policy.password_policy.configurations, "min_length", "not specified")
}

# Display policy type and template
output "mfa_policy_details" {
  value = "Policy: ${data.jumpcloud_policy.mfa_policy.name}, Type: ${data.jumpcloud_policy.mfa_policy.type}, Created at: ${data.jumpcloud_policy.mfa_policy.created}"
}
```

## Conditional Usage

This data source is useful for conditionally creating resources based on existing policies:

```hcl
# Check if an MFA policy already exists
data "jumpcloud_policy" "existing_mfa" {
  name = "MFA Policy"
  
  # This prevents an error if the policy doesn't exist
  depends_on = [jumpcloud_policy.default_mfa]
}

# Create a new policy only if one doesn't already exist
resource "jumpcloud_policy" "default_mfa" {
  count = data.jumpcloud_policy.existing_mfa.id == "" ? 1 : 0
  
  name = "MFA Policy"
  type = "mfa"
  # other attributes...
}
```

## Groups Usage

This data source can help you understand which groups a policy is applied to:

```hcl
data "jumpcloud_policy" "windows_password" {
  name = "Windows Password Policy"
}

# List all groups the policy is applied to
output "policy_groups" {
  value = data.jumpcloud_policy.windows_password.applied_to_groups
}
```

## Argument Reference

The following arguments are supported. **Note:** Exactly one of these arguments must be specified:

* `id` - (Optional) The ID of the policy to retrieve.
* `name` - (Optional) The name of the policy to retrieve.

## Attribute Reference

In addition to all the arguments above, the following attributes are exported:

* `type` - The type of policy (e.g., "password", "mfa", "lockout").
* `name` - The name of the policy.
* `active` - Whether the policy is active or not.
* `template` - The template ID used for the policy.
* `description` - Description of the policy's purpose.
* `configurations` - Map of configuration settings specific to the policy type.
* `created` - Timestamp when the policy was created.
* `updated` - Timestamp when the policy was last updated.
* `applied_to_groups` - List of group IDs this policy is applied to.
* `applied_to_systems` - List of system IDs this policy is applied to.

## Import

Policies found using this data source can be used to import resources:

```shell
terraform import jumpcloud_policy.imported ${data.jumpcloud_policy.existing.id}
``` 