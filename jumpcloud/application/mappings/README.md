# JumpCloud Application Mappings Package

This package provides resources for managing mappings between JumpCloud applications and users or groups.

## Resources

- `jumpcloud_application_user_mapping` - Manages mappings between JumpCloud applications and users
- `jumpcloud_application_group_mapping` - Manages mappings between JumpCloud applications and groups (user or system groups)

## Usage Examples

### Mapping a User to an Application

```hcl
resource "jumpcloud_application_user_mapping" "example" {
  application_id = jumpcloud_sso_application.example.id
  user_id        = jumpcloud_user.example.id
  
  attributes = {
    role = "admin"
    department = "IT"
  }
}
```

### Mapping a User Group to an Application

```hcl
resource "jumpcloud_application_group_mapping" "example" {
  application_id = jumpcloud_sso_application.example.id
  group_id       = jumpcloud_user_group.example.id
  type           = "user_group"
  
  attributes = {
    role = "user"
    access_level = "standard"
  }
}
```

### Mapping a System Group to an Application

```hcl
resource "jumpcloud_application_group_mapping" "systems" {
  application_id = jumpcloud_sso_application.example.id
  group_id       = jumpcloud_system_group.example.id
  type           = "system_group"
}
```

## Attribute Mapping

Both resources support application-specific attributes that can be used to customize the user or group experience in the target application. The `attributes` field is a map of key-value pairs that will be passed to the JumpCloud API when creating or updating the mapping.

## Import

Both resources can be imported using a composite ID format:

- For user mappings: `{application_id}:{user_id}`
- For group mappings: `{application_id}:{group_type}:{group_id}`

Example:

```bash
# Import a user mapping
terraform import jumpcloud_application_user_mapping.example 5f3c2b1a0e4f8d7c6b9a2d1e:5f3c2b1a0e4f8d7c6b9a2d1f

# Import a user group mapping
terraform import jumpcloud_application_group_mapping.example 5f3c2b1a0e4f8d7c6b9a2d1e:user_group:5f3c2b1a0e4f8d7c6b9a2d20

# Import a system group mapping
terraform import jumpcloud_application_group_mapping.systems 5f3c2b1a0e4f8d7c6b9a2d1e:system_group:5f3c2b1a0e4f8d7c6b9a2d21
``` 