# jumpcloud_user Data Source

Use this data source to get information about a JumpCloud user.

## Example Usage

```hcl
# Get a user by username
data "jumpcloud_user" "by_username" {
  username = "example.user"
}

# Get a user by email
data "jumpcloud_user" "by_email" {
  email = "example.user@example.com"
}

# Get a user by ID
data "jumpcloud_user" "by_id" {
  user_id = "5f0c1b2c3d4e5f6g7h8i9j0k"
}

output "user_details" {
  value = data.jumpcloud_user.by_username.attributes
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Optional) The username of the user to retrieve.
* `email` - (Optional) The email of the user to retrieve.
* `user_id` - (Optional) The ID of the user to retrieve.

**Note:** Exactly one of these arguments must be provided.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the user.
* `firstname` - The first name of the user.
* `lastname` - The last name of the user.
* `description` - The description of the user.
* `attributes` - A map of attributes for the user.
* `mfa_enabled` - Whether MFA is enabled for the user.
* `password_never_expires` - Whether the password never expires.
* `created` - The date the user was created. 