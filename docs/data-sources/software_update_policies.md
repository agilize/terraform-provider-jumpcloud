---
page_title: "JumpCloud: jumpcloud_software_update_policies"
subcategory: "Software Management"
description: |-
  Lists software update policies in JumpCloud
---

# jumpcloud_software_update_policies

This data source provides a list of software update policies in JumpCloud. Use this data source to query and filter existing software update policies.

## Example Usage

```terraform
# Retrieve all software update policies
data "jumpcloud_software_update_policies" "all" {
}

# Get macOS-specific update policies
data "jumpcloud_software_update_policies" "macos" {
  os_family = "macos"
  enabled   = true
}

# Get Windows update policies that auto-approve updates
data "jumpcloud_software_update_policies" "windows_auto" {
  os_family    = "windows"
  auto_approve = true
  limit        = 10
  sort         = "name"
  sort_dir     = "asc"
}

# Search for policies by name or description
data "jumpcloud_software_update_policies" "security" {
  search = "security"
  limit  = 20
}

# Output the total number of policies
output "total_policies" {
  value = data.jumpcloud_software_update_policies.all.total
}

# Output all macOS policy names
output "macos_policy_names" {
  value = [for policy in data.jumpcloud_software_update_policies.macos.policies : policy.name]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Filter policies by name (partial match).
* `os_family` - (Optional) Filter policies by operating system family. Valid values are `windows`, `macos`, or `linux`.
* `enabled` - (Optional) Filter policies by their enabled status.
* `auto_approve` - (Optional) Filter policies by their auto-approve setting.
* `search` - (Optional) Search term to match against policy names and descriptions.
* `sort` - (Optional) Field to sort results by. Valid values: `name`, `osFamily`, `enabled`, `created`, `updated`. Default: `name`.
* `sort_dir` - (Optional) Sort direction. Valid values: `asc`, `desc`. Default: `asc`.
* `limit` - (Optional) Maximum number of results to return. Default: `50`. Maximum: `1000`.
* `skip` - (Optional) Number of results to skip for pagination. Default: `0`.
* `org_id` - (Optional) Organization ID for multi-tenant environments.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `policies` - List of software update policies matching the specified criteria. Each policy contains:
  * `id` - The ID of the policy.
  * `name` - Name of the policy.
  * `description` - Description of the policy.
  * `os_family` - Operating system family (windows, macos, linux).
  * `enabled` - Whether the policy is enabled.
  * `all_packages` - Whether the policy applies to all packages.
  * `auto_approve` - Whether updates are auto-approved.
  * `status` - Current status of the policy.
  * `schedule` - Schedule configuration in JSON format.
  * `target_count` - Number of targets for this policy.
  * `org_id` - Organization ID if specified.
  * `created` - Timestamp when the policy was created.
  * `updated` - Timestamp when the policy was last updated.

* `total` - Total number of policies matching the filter criteria. 