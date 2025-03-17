# jumpcloud_application_user_mapping Resource

Manages mappings between users and applications in JumpCloud, granting access for individual users to SSO applications.

## Example Usage

```hcl
# Basic user to application mapping
resource "jumpcloud_application_user_mapping" "admin_salesforce" {
  application_id = jumpcloud_application.salesforce.id
  user_id        = jumpcloud_user.admin.id
}

# Mapping with custom attributes
resource "jumpcloud_application_user_mapping" "dev_jira" {
  application_id = jumpcloud_application.jira.id
  user_id        = jumpcloud_user.developer.id
  
  attributes = {
    "role"     = "developer"
    "projects" = "alpha,beta,gamma"
    "admin"    = "false"
  }
}

# Mapping using data sources for existing resources
resource "jumpcloud_application_user_mapping" "existing_mapping" {
  application_id = data.jumpcloud_application.existing_app.id
  user_id        = data.jumpcloud_user.existing_user.id
  
  attributes = {
    "access_level" = "standard"
    "department"   = "marketing"
  }
}
```

## Argument Reference

The following arguments are supported:

* `application_id` - (Required) JumpCloud application ID.
* `user_id` - (Required) JumpCloud user ID.
* `attributes` - (Optional) Map of custom attributes for the mapping. These attributes are specific to each application type and can be used to define roles, permissions, or other application-specific settings for the user.

## Import

JumpCloud application-user mappings can be imported using a colon-separated string in the format:

```
terraform import jumpcloud_application_user_mapping.example {application_id}:{user_id}
``` 