# Basic Provider Configuration Example

# Configure the JumpCloud Provider
terraform {
  required_providers {
    jumpcloud = {
      source  = "registry.terraform.io/agilize/jumpcloud"
      version = "~> 1.0"
    }
  }
}

# Option 1: Configure with environment variables
# Set JUMPCLOUD_API_KEY and JUMPCLOUD_ORG_ID before running terraform
provider "jumpcloud" {}

# Option 2: Configure with static credentials (not recommended for production)
# provider "jumpcloud" {
#   api_key = "your-api-key-here"
#   org_id  = "your-org-id-here"  # Optional
# }

# Option 3: Configure with variables (recommended)
# provider "jumpcloud" {
#   api_key = var.jumpcloud_api_key
#   org_id  = var.jumpcloud_org_id
# }

# Example resource (uncomment to test)
# resource "jumpcloud_user_group" "example" {
#   name        = "Example Group"
#   description = "This is an example group"
# }

# Output the provider version
output "provider_version" {
  value = "JumpCloud Provider ${jumpcloud.version}"
} 