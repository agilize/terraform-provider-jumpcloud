provider "jumpcloud" {
  # API key can be provided via JUMPCLOUD_API_KEY environment variable
}

# Basic Linux administrators group with sudo access
resource "jumpcloud_user_group" "linux_admins" {
  name        = "linux-administrators"
  description = "Group for Linux administrators with sudo access"
  
  attributes = {
    # Sudo settings as a nested object
    sudo = {
      enabled         = true
      withoutPassword = false
    }
    # Enable Samba authentication
    sambaEnabled = true
    # Create Linux group with posixGroups as an array
    posixGroups = [
      {
        name = "admins"
      }
    ]
  }
}

# Linux developers group with limited sudo access
resource "jumpcloud_user_group" "linux_devs" {
  name        = "linux-developers"
  description = "Group for Linux developers with limited sudo access"
  
  attributes = {
    # Sudo settings with passwordless sudo enabled
    sudo = {
      enabled         = true
      withoutPassword = true
    }
    # Create Linux group
    posixGroups = [
      {
        name = "developers"
      }
    ]
  }
}

# Create users
resource "jumpcloud_user" "admin_user" {
  username  = "admin.user"
  email     = "admin.user@example.com"
  firstname = "Admin"
  lastname  = "User"
  password  = "SecurePassword123!"
}

resource "jumpcloud_user" "dev_user" {
  username  = "dev.user"
  email     = "dev.user@example.com"
  firstname = "Dev"
  lastname  = "User"
  password  = "SecurePassword123!"
}

# Add users to groups
resource "jumpcloud_user_group_membership" "admin_membership" {
  user_id       = jumpcloud_user.admin_user.id
  user_group_id = jumpcloud_user_group.linux_admins.id
}

resource "jumpcloud_user_group_membership" "dev_membership" {
  user_id       = jumpcloud_user.dev_user.id
  user_group_id = jumpcloud_user_group.linux_devs.id
}

# Output group IDs for reference
output "linux_admins_group_id" {
  description = "ID of the Linux administrators group"
  value       = jumpcloud_user_group.linux_admins.id
}

output "linux_devs_group_id" {
  description = "ID of the Linux developers group"
  value       = jumpcloud_user_group.linux_devs.id
}
