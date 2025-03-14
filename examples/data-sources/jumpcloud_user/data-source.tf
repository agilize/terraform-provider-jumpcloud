provider "jumpcloud" {
  api_key = var.jumpcloud_api_key
  org_id  = var.jumpcloud_org_id
}

data "jumpcloud_user" "by_username" {
  username = "existing.user"
}

data "jumpcloud_user" "by_email" {
  email = "existing.user@example.com"
}

output "user_details" {
  value = data.jumpcloud_user.by_username
  sensitive = true
} 