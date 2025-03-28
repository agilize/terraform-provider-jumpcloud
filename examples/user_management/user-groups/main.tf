# User and Group Management Example

terraform {
  required_providers {
    jumpcloud = {
      source  = "registry.terraform.io/agilize/jumpcloud"
      version = "~> 1.0"
    }
  }
}

provider "jumpcloud" {}

# Create departments as user groups
resource "jumpcloud_user_group" "engineering" {
  name        = "Engineering Team"
  description = "All members of the engineering team"
  
  attributes = {
    department = "Engineering"
    location   = "Global"
  }
}

resource "jumpcloud_user_group" "marketing" {
  name        = "Marketing Team"
  description = "All members of the marketing team"
  
  attributes = {
    department = "Marketing"
    location   = "Global"
  }
}

resource "jumpcloud_user_group" "finance" {
  name        = "Finance Team"
  description = "All members of the finance team"
  
  attributes = {
    department = "Finance"
    location   = "Global"
  }
}

# Create project-based user groups
resource "jumpcloud_user_group" "project_alpha" {
  name        = "Project Alpha"
  description = "Members working on Project Alpha"
  
  attributes = {
    project_type = "Development"
    priority     = "High"
  }
}

# Create users
resource "jumpcloud_user" "john_doe" {
  username   = "john.doe"
  email      = "john.doe@example.com"
  firstname  = "John"
  lastname   = "Doe"
  password   = "SecurePassword123!" # In production, use a secure method for passwords
}

resource "jumpcloud_user" "jane_smith" {
  username   = "jane.smith"
  email      = "jane.smith@example.com"
  firstname  = "Jane"
  lastname   = "Smith"
  password   = "SecurePassword123!" # In production, use a secure method for passwords
}

resource "jumpcloud_user" "sam_johnson" {
  username   = "sam.johnson"
  email      = "sam.johnson@example.com"
  firstname  = "Sam"
  lastname   = "Johnson"
  password   = "SecurePassword123!" # In production, use a secure method for passwords
}

# Assign users to department groups
resource "jumpcloud_user_group_membership" "john_engineering" {
  user_group_id = jumpcloud_user_group.engineering.id
  user_id       = jumpcloud_user.john_doe.id
}

resource "jumpcloud_user_group_membership" "jane_marketing" {
  user_group_id = jumpcloud_user_group.marketing.id
  user_id       = jumpcloud_user.jane_smith.id
}

resource "jumpcloud_user_group_membership" "sam_finance" {
  user_group_id = jumpcloud_user_group.finance.id
  user_id       = jumpcloud_user.sam_johnson.id
}

# Assign users to project groups
resource "jumpcloud_user_group_membership" "john_project_alpha" {
  user_group_id = jumpcloud_user_group.project_alpha.id
  user_id       = jumpcloud_user.john_doe.id
}

resource "jumpcloud_user_group_membership" "jane_project_alpha" {
  user_group_id = jumpcloud_user_group.project_alpha.id
  user_id       = jumpcloud_user.jane_smith.id
}

# Outputs
output "engineering_group_id" {
  value = jumpcloud_user_group.engineering.id
}

output "project_alpha_members" {
  value = [
    jumpcloud_user.john_doe.username,
    jumpcloud_user.jane_smith.username
  ]
} 