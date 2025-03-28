# JumpCloud User Groups Module

This module provides resources for managing JumpCloud user groups and memberships.

## Resources

### jumpcloud_user_group

The `jumpcloud_user_group` resource allows you to create and manage JumpCloud user groups.

#### Example Usage

```hcl
resource "jumpcloud_user_group" "engineering" {
  name        = "Engineering Team"
  description = "Group for all engineering staff"
  
  attributes = {
    department = "Engineering"
    location   = "Remote"
  }
}
```

#### Argument Reference

* `name` - (Required) The name of the user group.
* `description` - (Optional) A description for the user group.
* `type` - (Optional) The type of the user group.
* `attributes` - (Optional) A map of attributes to assign to the user group.

#### Attribute Reference

* `id` - The ID of the user group.
* `member_count` - The number of users in the group.
* `created` - The creation timestamp of the user group.

### jumpcloud_user_group_membership

The `jumpcloud_user_group_membership` resource allows you to manage user memberships in JumpCloud user groups.

#### Example Usage

```hcl
resource "jumpcloud_user" "example" {
  username  = "example"
  email     = "example@example.com"
  firstname = "Example"
  lastname  = "User"
}

resource "jumpcloud_user_group" "engineering" {
  name        = "Engineering Team"
  description = "Group for all engineering staff"
}

resource "jumpcloud_user_group_membership" "example_membership" {
  user_group_id = jumpcloud_user_group.engineering.id
  user_id       = jumpcloud_user.example.id
}
```

#### Argument Reference

* `user_group_id` - (Required) The ID of the user group.
* `user_id` - (Required) The ID of the user to add to the group.

#### Attribute Reference

* `id` - The ID of the membership (format: `group_id:user_id`).

## Relationship with Other Resources

User groups can be associated with:
- Users (via `jumpcloud_user_group_membership`)
- Systems (via system group associations)
- Applications (via application assignments)
- Policies (via policy bindings) 