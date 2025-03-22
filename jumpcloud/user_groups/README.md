# User Groups

This directory contains resources related to JumpCloud user groups.

## Resources

- `jumpcloud_user_group` - Manages user groups in JumpCloud
- `jumpcloud_user_group_membership` - Manages the association of users to user groups in JumpCloud

## Usage Examples

### User Group

```terraform
resource "jumpcloud_user_group" "example" {
  name        = "Development Team"
  description = "Group for development team members"
}
```

### User Group Membership

```terraform
resource "jumpcloud_user_group_membership" "example" {
  user_group_id = jumpcloud_user_group.example.id
  user_id       = jumpcloud_user.example.id
}
```

## Relationship with Other Resources

User groups can be associated with:
- Users (via `jumpcloud_user_group_membership`)
- Systems (via system group associations)
- Applications (via application assignments)
- Policies (via policy bindings) 