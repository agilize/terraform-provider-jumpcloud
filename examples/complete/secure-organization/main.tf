# Secure Organization Setup Example
# This example demonstrates a comprehensive JumpCloud setup for a secure organization
# with proper user management, system groups, and authentication policies.

terraform {
  required_providers {
    jumpcloud = {
      source  = "registry.terraform.io/agilize/jumpcloud"
      version = "~> 1.0"
    }
  }
}

provider "jumpcloud" {}

# ---------------------------------------------------
# Organization Structure - User Groups
# ---------------------------------------------------

# Department Groups
resource "jumpcloud_user_group" "it_admin" {
  name        = "IT Administrators"
  description = "IT administrators with elevated privileges"
  
  attributes = {
    department = "IT"
    privilege  = "admin"
  }
}

resource "jumpcloud_user_group" "developers" {
  name        = "Developers"
  description = "Software development team"
  
  attributes = {
    department = "Engineering"
    role       = "development"
  }
}

resource "jumpcloud_user_group" "finance" {
  name        = "Finance"
  description = "Finance department"
  
  attributes = {
    department = "Finance"
    role       = "staff"
  }
}

resource "jumpcloud_user_group" "contractors" {
  name        = "External Contractors"
  description = "External contractors with limited access"
  
  attributes = {
    type      = "external"
    privilege = "limited"
  }
}

# ---------------------------------------------------
# System Groups
# ---------------------------------------------------

resource "jumpcloud_system_group" "production_servers" {
  name        = "Production Servers"
  description = "Production environment servers"
}

resource "jumpcloud_system_group" "development_servers" {
  name        = "Development Servers"
  description = "Development environment servers"
}

resource "jumpcloud_system_group" "finance_systems" {
  name        = "Finance Systems"
  description = "Systems containing financial data"
}

resource "jumpcloud_system_group" "workstations" {
  name        = "Employee Workstations"
  description = "Employee laptops and desktops"
}

# ---------------------------------------------------
# Users
# ---------------------------------------------------

resource "jumpcloud_user" "admin_user" {
  username   = "admin.user"
  email      = "admin@example.com"
  firstname  = "Admin"
  lastname   = "User"
  password   = "SecurePassword123!" # In production, use a secure method
}

resource "jumpcloud_user" "dev_user" {
  username   = "dev.user"
  email      = "dev@example.com"
  firstname  = "Developer"
  lastname   = "User"
  password   = "SecurePassword123!" # In production, use a secure method
}

resource "jumpcloud_user" "finance_user" {
  username   = "finance.user"
  email      = "finance@example.com"
  firstname  = "Finance"
  lastname   = "User"
  password   = "SecurePassword123!" # In production, use a secure method
}

resource "jumpcloud_user" "contractor" {
  username   = "contractor"
  email      = "contractor@external.com"
  firstname  = "External"
  lastname   = "Contractor"
  password   = "SecurePassword123!" # In production, use a secure method
}

# ---------------------------------------------------
# User Group Memberships
# ---------------------------------------------------

resource "jumpcloud_user_group_membership" "admin_membership" {
  user_group_id = jumpcloud_user_group.it_admin.id
  user_id       = jumpcloud_user.admin_user.id
}

resource "jumpcloud_user_group_membership" "dev_membership" {
  user_group_id = jumpcloud_user_group.developers.id
  user_id       = jumpcloud_user.dev_user.id
}

resource "jumpcloud_user_group_membership" "finance_membership" {
  user_group_id = jumpcloud_user_group.finance.id
  user_id       = jumpcloud_user.finance_user.id
}

resource "jumpcloud_user_group_membership" "contractor_membership" {
  user_group_id = jumpcloud_user_group.contractors.id
  user_id       = jumpcloud_user.contractor.id
}

# ---------------------------------------------------
# Authentication Policies
# ---------------------------------------------------

# Admin Policy - Strict security for IT admins
resource "jumpcloud_policies_policy" "admin_policy" {
  name        = "Admin Security Policy"
  description = "Strict security policy for administrators"
  
  rule {
    type = "AUTHENTICATION"
    
    conditions {
      resource {
        type = "USER_GROUP"
        id   = jumpcloud_user_group.it_admin.id
      }
    }
    
    effects {
      allow_ssh_password_authentication    = false
      allow_multi_factor_authentication    = true
      force_multi_factor_authentication    = true
      require_password_reset               = false
      allow_password_management_self_serve = true
    }
  }
}

# Developer Policy - Standard security
resource "jumpcloud_policies_policy" "developer_policy" {
  name        = "Developer Security Policy"
  description = "Security policy for development team"
  
  rule {
    type = "AUTHENTICATION"
    
    conditions {
      resource {
        type = "USER_GROUP"
        id   = jumpcloud_user_group.developers.id
      }
    }
    
    effects {
      allow_ssh_password_authentication    = false
      allow_multi_factor_authentication    = true
      force_multi_factor_authentication    = true
      require_password_reset               = false
      allow_password_management_self_serve = true
    }
  }
}

# Contractor Policy - Restricted access
resource "jumpcloud_policies_policy" "contractor_policy" {
  name        = "Contractor Security Policy"
  description = "Restricted policy for external contractors"
  
  rule {
    type = "AUTHENTICATION"
    
    conditions {
      resource {
        type = "USER_GROUP"
        id   = jumpcloud_user_group.contractors.id
      }
    }
    
    effects {
      allow_ssh_password_authentication    = false
      allow_multi_factor_authentication    = true
      force_multi_factor_authentication    = true
      require_password_reset               = true
      allow_password_management_self_serve = false
      max_password_age_days                = 30
    }
  }
}

# ---------------------------------------------------
# IP Security
# ---------------------------------------------------

# Define allowed office IP ranges
resource "jumpcloud_ip_list" "office_ips" {
  name        = "Office IP Ranges"
  description = "Approved office IP addresses"
  
  addresses = [
    "192.168.1.0/24",  # Example office network
    "10.0.0.0/16"      # Example VPN network
  ]
}

# Apply IP restrictions to contractor access
resource "jumpcloud_policies_conditional_access_rule" "contractor_ip_rule" {
  name        = "Contractor IP Restriction"
  description = "Restrict contractor access to office IPs only"
  
  user_groups = [jumpcloud_user_group.contractors.id]
  
  ip_addresses {
    include_lists = [jumpcloud_ip_list.office_ips.id]
  }
  
  effective_start_time = "00:00"
  effective_end_time   = "23:59"
  
  action = "ALLOW" # Only allow access from these IPs
}

# ---------------------------------------------------
# Password Policies
# ---------------------------------------------------

resource "jumpcloud_password_policy" "organization_policy" {
  name                    = "Organization Password Policy"
  description             = "Standard password policy for the organization"
  minimum_length          = 12
  maximum_length          = 64
  require_uppercase       = true
  require_lowercase       = true
  require_number          = true
  require_special         = true
  password_history        = 5
  password_expiry_days    = 90
  lockout_attempts        = 5
  lockout_time_minutes    = 30
  user_lockout_action     = "DISABLE"
  minimum_password_age    = 1
}

# ---------------------------------------------------
# Outputs
# ---------------------------------------------------

output "admin_group_id" {
  value = jumpcloud_user_group.it_admin.id
}

output "contractor_policy_id" {
  value = jumpcloud_policies_policy.contractor_policy.id
}

output "password_policy_id" {
  value = jumpcloud_password_policy.organization_policy.id
} 