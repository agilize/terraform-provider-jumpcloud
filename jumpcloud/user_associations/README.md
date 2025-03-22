# User Associations

This directory contains resources related to JumpCloud user associations with other entities.

## Resources

- `jumpcloud_user_system_association` - Manages the association between users and systems in JumpCloud

## Usage Examples

### User System Association

```terraform
resource "jumpcloud_user_system_association" "example" {
  user_id   = jumpcloud_user.example.id
  system_id = jumpcloud_system.example.id
}
```

## Relationship with Other Resources

User associations connect users with:
- Systems (via `jumpcloud_user_system_association`)
- Groups (via `jumpcloud_user_group_membership`)
- Applications (via application mappings)
- Policies (via policy bindings)

These association resources enable you to define the relationships between users and other resources within JumpCloud. 