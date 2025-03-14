provider "jumpcloud" {
  api_key = var.jumpcloud_api_key
  org_id  = var.jumpcloud_org_id
}

resource "jumpcloud_user" "example" {
  username    = "example.user"
  email       = "example.user@example.com"
  firstname   = "Example"
  lastname    = "User"
  password    = "securePassword123!"
  description = "Created by Terraform"
  
  attributes = {
    department = "IT"
    location   = "Remote"
  }
  
  mfa_enabled          = true
  password_never_expires = false
} 