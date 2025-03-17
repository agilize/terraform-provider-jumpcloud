# Retrieve all software update policies
data "jumpcloud_software_update_policies" "all" {
}

# Get macOS-specific update policies
data "jumpcloud_software_update_policies" "macos" {
  os_family = "macos"
  enabled   = true
}

# Get Windows update policies that auto-approve updates
data "jumpcloud_software_update_policies" "windows_auto" {
  os_family    = "windows"
  auto_approve = true
  limit        = 10
  sort         = "name"
  sort_dir     = "asc"
}

# Search for policies by name
data "jumpcloud_software_update_policies" "security" {
  search = "security"
  limit  = 20
}

# Output the total number of policies
output "total_policies" {
  value = data.jumpcloud_software_update_policies.all.total
}

# Output all macOS policy names
output "macos_policy_names" {
  value = [for policy in data.jumpcloud_software_update_policies.macos.policies : policy.name]
}

# Reference a specific policy by its index
output "first_auto_approve_windows_policy" {
  value = length(data.jumpcloud_software_update_policies.windows_auto.policies) > 0 ? data.jumpcloud_software_update_policies.windows_auto.policies[0].name : "No matching policy found"
}

# Use policy data to create references to target systems
data "jumpcloud_software_update_policies" "daily" {
  search = "daily"
}

# Output policy details for use in other resources
output "daily_policy_details" {
  value = length(data.jumpcloud_software_update_policies.daily.policies) > 0 ? {
    id           = data.jumpcloud_software_update_policies.daily.policies[0].id
    name         = data.jumpcloud_software_update_policies.daily.policies[0].name
    os_family    = data.jumpcloud_software_update_policies.daily.policies[0].os_family
    target_count = data.jumpcloud_software_update_policies.daily.policies[0].target_count
  } : null
} 