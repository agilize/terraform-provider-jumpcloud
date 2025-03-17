# jumpcloud_application_group_mapping Resource

Manages mappings between groups and applications in JumpCloud, granting access for user groups or system groups to SSO applications.

## Example Usage

```hcl
# Mapping a user group to an application
resource "jumpcloud_application_group_mapping" "marketing_team" {
  application_id = jumpcloud_application.salesforce.id
  group_id       = jumpcloud_user_group.marketing.id
  type           = "user_group"  # Default
  
  attributes = {
    "access_level" = "standard"
    "department"   = "Marketing"
  }
}

# Mapping a system group to an application
resource "jumpcloud_application_group_mapping" "production_servers" {
  application_id = jumpcloud_application.monitoring_tool.id
  group_id       = jumpcloud_system_group.production.id
  type           = "system_group"
}

# Mapping using data sources for existing resources
resource "jumpcloud_application_group_mapping" "existing_mapping" {
  application_id = data.jumpcloud_application.existing_app.id
  group_id       = data.jumpcloud_user_group.existing_group.id
  
  attributes = {
    "role"     = "viewer"
    "region"   = "us-west"
    "enabled"  = "true"
  }
}
```

## Argument Reference

The following arguments are supported:

* `application_id` - (Required) JumpCloud application ID.
* `group_id` - (Required) JumpCloud group ID.
* `type` - (Optional) Group type: `user_group` (default) or `system_group`.
* `attributes` - (Optional) Map of custom attributes for the mapping. These attributes are specific to each application type.

## Import

JumpCloud application-group mappings can be imported using a colon-separated string in the format:

```
terraform import jumpcloud_application_group_mapping.example {application_id}:{group_type}:{group_id}
``` 