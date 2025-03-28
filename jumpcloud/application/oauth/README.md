# JumpCloud OAuth Package

This package provides resources and data sources for managing OAuth components in JumpCloud.

## Resources

### `jumpcloud_oauth_authorization`
Manages OAuth authorizations for JumpCloud applications, defining client details and allowed scopes.

### `jumpcloud_oauth_user`
Manages OAuth user assignments, controlling which users have OAuth access to specific applications and what scopes they have.

## Data Sources

### `jumpcloud_oauth_users`
Provides a list of OAuth users associated with a JumpCloud application, supporting filtering and pagination.

## Usage Examples

### Creating an OAuth Authorization

```hcl
resource "jumpcloud_oauth_authorization" "example" {
  application_id      = "5f0e9e9f9d8f7b0123456789"
  expires_at          = "2024-12-31T23:59:59Z"
  client_name         = "Example Client"
  client_description  = "OAuth client for example application"
  client_contact_email = "admin@example.com"
  client_redirect_uris = [
    "https://example.com/oauth/callback",
    "https://app.example.com/auth"
  ]
  scopes = [
    "read:users",
    "write:users",
    "read:groups"
  ]
}
```

### Assigning OAuth Access to a User

```hcl
resource "jumpcloud_oauth_user" "example_user" {
  application_id = jumpcloud_oauth_authorization.example.application_id
  user_id = jumpcloud_user.john.id
  scopes = [
    "read:users",
    "read:groups"
  ]
}
```

### Fetching OAuth Users for an Application

```hcl
data "jumpcloud_oauth_users" "app_users" {
  application_id = jumpcloud_oauth_authorization.example.application_id
  limit = 100
  sort = "username"
  sort_dir = "asc"
  filter = "username:contains:admin"
}

output "admin_users" {
  value = data.jumpcloud_oauth_users.app_users.users
}
```

## Key Concepts

### OAuth Authorizations
OAuth authorizations define the client details and scopes for a JumpCloud application. They have an expiration date and can be configured with redirect URIs for OAuth flows.

### OAuth Users
OAuth users are JumpCloud users that have been granted OAuth access to specific applications. Each user assignment defines the scopes the user is granted.

### Scopes
Scopes define the permissions granted to OAuth clients and users. Common scopes include:
- `read:users` - Read user information
- `write:users` - Create and update users
- `read:groups` - Read group information
- `write:groups` - Create and update groups 