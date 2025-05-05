provider "jumpcloud" {
  # API key can be provided via JUMPCLOUD_API_KEY environment variable
}

# Create a user group with custom attributes
resource "jumpcloud_user_group" "complete" {
  name        = "complete-test-group"
  description = "A complete test user group with all features"
  
  # Custom attributes with special characters in names
  attributes = {
    "department"      = "Engineering"
    "location"        = "Remote"
    "cost-center"     = "CC-123456"  # Contains hyphen
    "project.name"    = "Terraform"  # Contains period
    "security_level"  = "High"
  }
}

# Create multiple users with different attributes
resource "jumpcloud_user" "user1" {
  username  = "complete-user1"
  email     = "complete-user1@example.com"
  firstname = "Complete"
  lastname  = "User One"
  password  = "ComplexP@ssw0rd!"
  
  # Standard fields
  company    = "Example Corp"
  department = "Engineering"
  costcenter = "CC-123456"
  
  # Phone numbers with formatting
  phone_numbers {
    type   = "work"
    number = "555-123-4567"  # Contains hyphens
  }
  
  phone_numbers {
    type   = "mobile"
    number = "(555) 987-6543"  # Contains parentheses and spaces
  }
  
  # Custom attributes with special characters
  attributes = {
    "employee.id"     = "EMP-12345"
    "access-level"    = "Admin"
    "start_date"      = "2023-01-15"
  }
}

resource "jumpcloud_user" "user2" {
  username  = "complete-user2"
  email     = "complete-user2@example.com"
  firstname = "Complete"
  lastname  = "User Two"
  password  = "ComplexP@ssw0rd!"
  
  company    = "Example Corp"
  department = "Product"
  
  # Boolean fields
  mfa_enabled                  = true
  enable_managed_uid           = true
  bypass_managed_device_lockout = true
  
  # Special string fields
  password_recovery_email      = "recovery@example.com"
  delegated_authority          = "DELEGATED"
  password_authority           = "MANAGED"
}

# Add users to the group
resource "jumpcloud_user_group_membership" "user1" {
  user_id  = jumpcloud_user.user1.id
  group_id = jumpcloud_user_group.complete.id
}

resource "jumpcloud_user_group_membership" "user2" {
  user_id  = jumpcloud_user.user2.id
  group_id = jumpcloud_user_group.complete.id
}

# Retrieve the user group using the data source by name
data "jumpcloud_user_group" "by_name" {
  name = jumpcloud_user_group.complete.name
  depends_on = [jumpcloud_user_group.complete]
}

# Retrieve the user group using the data source by ID
data "jumpcloud_user_group" "by_id" {
  group_id = jumpcloud_user_group.complete.id
  depends_on = [jumpcloud_user_group.complete]
}

# Retrieve users using the data source
data "jumpcloud_user" "user1" {
  username = jumpcloud_user.user1.username
  depends_on = [jumpcloud_user.user1]
}

data "jumpcloud_user" "user2" {
  username = jumpcloud_user.user2.username
  depends_on = [jumpcloud_user.user2]
}

# Outputs
output "user_group_id" {
  value = jumpcloud_user_group.complete.id
}

output "user_group_name" {
  value = jumpcloud_user_group.complete.name
}

output "user_group_attributes" {
  value = jumpcloud_user_group.complete.attributes
}

output "data_source_group_by_name" {
  value = data.jumpcloud_user_group.by_name.id
}

output "data_source_group_by_id" {
  value = data.jumpcloud_user_group.by_id.attributes
}

output "user1_id" {
  value = jumpcloud_user.user1.id
}

output "user1_attributes" {
  value = jumpcloud_user.user1.attributes
}

output "user2_id" {
  value = jumpcloud_user.user2.id
}

output "user2_mfa_enabled" {
  value = jumpcloud_user.user2.mfa_enabled
}
