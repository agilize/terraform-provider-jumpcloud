# JumpCloud Password Manager Package

This package provides resources and data sources for managing password safes and entries in JumpCloud.

## Resources

### `jumpcloud_password_safe`
Manages password safes in JumpCloud, which are secure containers for storing password entries.

### `jumpcloud_password_entry`
Manages password entries within a password safe, including credentials, notes, and metadata.

## Data Sources

### `jumpcloud_password_safes`
Provides a list of password safes in JumpCloud with filtering and pagination support.

## Usage Examples

### Creating a Password Safe

```hcl
resource "jumpcloud_password_safe" "team_safe" {
  name        = "Development Team Safe"
  description = "Shared credentials for the development team"
  type        = "team"
  status      = "active"
  owner_id    = "5f0e9e9f9d8f7b0123456789"
  
  member_ids = [
    "5f0e9e9f9d8f7b0123456789",
    "5f0e9e9f9d8f7b0123456790"
  ]
  
  group_ids = [
    "5f0e9e9f9d8f7b0123456791"  # Development group
  ]
}
```

### Creating a Password Entry

```hcl
resource "jumpcloud_password_entry" "database_credentials" {
  safe_id     = jumpcloud_password_safe.team_safe.id
  name        = "Production Database"
  description = "Credentials for the production PostgreSQL database"
  type        = "database"
  
  username    = "admin"
  password    = "supersecurepassword"  # Consider using variables or secrets management
  url         = "postgres://db.example.com:5432"
  notes       = "Primary database for the application. Contact DevOps for access."
  
  tags = [
    "production",
    "database",
    "postgresql"
  ]
  
  metadata = {
    "database_type" = "postgresql",
    "version"       = "13.4",
    "host"          = "db.example.com",
    "port"          = "5432"
  }
  
  folder    = "Production/Databases"
  favorite  = true
}
```

### Fetching Password Safes

```hcl
data "jumpcloud_password_safes" "team_safes" {
  type     = "team"
  status   = "active"
  owner_id = "5f0e9e9f9d8f7b0123456789"
  limit    = 50
  sort     = "name"
  sort_dir = "asc"
}

output "team_safes" {
  value = data.jumpcloud_password_safes.team_safes.safes
}
```

## Safe Types

JumpCloud supports three types of password safes:

- **personal**: Personal safes are owned by a single user and cannot have members or groups.
- **team**: Team safes are designed for team collaboration and can have multiple members and groups.
- **shared**: Shared safes are for broader organizational use and can have multiple members and groups.

## Entry Types

Password entries can be of various types, including:

- `site`: Website credentials
- `application`: Application credentials
- `database`: Database connection details
- `ssh`: SSH keys and server access information
- `server`: Server access credentials
- `email`: Email account credentials
- `note`: Secure notes
- `creditcard`: Credit card information
- `identity`: Personal identity information
- `file`: File attachments (encrypted)
- `wifi`: WiFi network credentials
- `custom`: Custom type for other credentials 