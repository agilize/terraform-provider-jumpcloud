---
page_title: "JumpCloud: jumpcloud_software_update_policy"
subcategory: "Software Management"
description: |-
  Manages a software update policy in JumpCloud
---

# jumpcloud_software_update_policy

This resource allows you to create, update, and delete software update policies in JumpCloud. Software update policies define how updates are delivered to systems, what packages are included, and when updates are applied.

## Example Usage

```terraform
# macOS update policy example
resource "jumpcloud_software_update_policy" "macos_policy" {
  name        = "macOS Updates Policy"
  description = "Regular security updates for macOS systems"
  os_family   = "macos"
  enabled     = true
  
  # Schedule updates to run every Sunday at 2:00 AM
  schedule = jsonencode({
    type     = "weekly"
    dayOfWeek = "sunday"
    hour     = 2
    minute   = 0
  })
  
  # Apply to all macOS packages
  all_packages = true
  
  # Auto-approve updates
  auto_approve = true
  
  # Target system group
  system_group_targets = [
    "5f43a22171f9a42f55656cc2"
  ]
}

# Windows updates policy with specific package IDs
resource "jumpcloud_software_update_policy" "windows_policy" {
  name        = "Windows Critical Updates"
  description = "Only apply critical Windows security updates"
  os_family   = "windows"
  enabled     = true
  
  # Schedule updates for the last Friday of each month
  schedule = jsonencode({
    type       = "monthly"
    dayOfMonth = "last-friday"
    hour       = 22
    minute     = 0
    timeZone   = "America/New_York"
  })
  
  # Specify particular package IDs to update
  package_ids = [
    "5f43a30571f9a42f55656cc3",
    "5f43a33d71f9a42f55656cc4"
  ]
  
  # Require manual approval
  auto_approve = false
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the software update policy.
* `os_family` - (Required, Forces new resource) Operating system family this policy applies to. Valid values are `windows`, `macos`, or `linux`.
* `schedule` - (Required) JSON-encoded schedule configuration. Structure varies depending on schedule type.
* `description` - (Optional) Description of the update policy.
* `enabled` - (Optional) Indicates if the policy is active. Default is `true`.
* `package_ids` - (Optional) List of specific package IDs to update. Conflicts with `all_packages`.
* `all_packages` - (Optional) When set to `true`, all compatible packages will be updated. Conflicts with `package_ids`. Default is `false`.
* `auto_approve` - (Optional) When set to `true`, updates are applied automatically without manual approval. Default is `false`.
* `system_targets` - (Optional) List of system IDs to target with this policy.
* `system_group_targets` - (Optional) List of system group IDs to target with this policy.
* `org_id` - (Optional, Forces new resource) Organization ID for multi-tenant environments.

### Schedule Configuration

The `schedule` attribute is a JSON-encoded object that defines when updates are applied. The following schedule types are supported:

* **Daily Schedule**:
  ```json
  {
    "type": "daily",
    "hour": 3,
    "minute": 0
  }
  ```

* **Weekly Schedule**:
  ```json
  {
    "type": "weekly",
    "dayOfWeek": "sunday",
    "hour": 2,
    "minute": 0
  }
  ```

* **Monthly Schedule**:
  ```json
  {
    "type": "monthly",
    "dayOfMonth": 15,
    "hour": 1,
    "minute": 30
  }
  ```

  The `dayOfMonth` can also be specified as `"first-monday"`, `"last-friday"`, etc.

* **All schedules can optionally include a `timeZone` attribute**:
  ```json
  {
    "timeZone": "America/New_York"
  }
  ```

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the software update policy.
* `created` - Timestamp when the policy was created.
* `updated` - Timestamp when the policy was last updated.

## Import

Software update policies can be imported using the policy ID, e.g.,

```
$ terraform import jumpcloud_software_update_policy.example 5f43a41b71f9a42f55656cc6
```

For multi-tenant environments, specify the organization ID:

```
$ terraform import jumpcloud_software_update_policy.example 5f43a41b71f9a42f55656cc6,org_id=5f43a52971f9a42f55656cc7
``` 