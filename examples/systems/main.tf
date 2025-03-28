provider "jumpcloud" {
  # API Key can be provided via JUMPCLOUD_API_KEY environment variable
  # api_key = "your-api-key"
  
  # Org ID can be provided via JUMPCLOUD_ORG_ID environment variable
  # org_id = "your-org-id"
}

# Create a JumpCloud system
resource "jumpcloud_system" "example" {
  display_name                       = "example-system"
  description                        = "Example system managed by Terraform"
  allow_ssh_root_login               = false
  allow_ssh_password_authentication  = true
  allow_multi_factor_authentication  = true
  tags                               = ["terraform", "example"]
  
  attributes = {
    location = "Remote"
    department = "Engineering"
  }
}

# Get system information using the data source
data "jumpcloud_system" "example" {
  system_id = jumpcloud_system.example.id
}

# Output system information
output "system_id" {
  value = data.jumpcloud_system.example.id
}

output "system_name" {
  value = data.jumpcloud_system.example.display_name
}

output "system_type" {
  value = data.jumpcloud_system.example.system_type
}

output "system_os" {
  value = data.jumpcloud_system.example.os
} 