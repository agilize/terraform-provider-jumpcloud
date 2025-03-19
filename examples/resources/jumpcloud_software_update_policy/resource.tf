resource "jumpcloud_software_update_policy" "macos_policy" {
  name        = "macOS Updates Policy"
  description = "Regular security updates for macOS systems"
  os_family   = "macos"
  enabled     = true
  
  # Schedule updates to run every Sunday at 2:00 AM
  schedule = jsonencode({
    type     = "weekly"
    dayOfWeek = "sunday"
    hour     = 2
    minute   = 0
  })
  
  # Apply to all macOS packages
  all_packages = true
  
  # Auto-approve updates
  auto_approve = true
  
  # Target specific systems
  system_targets = [
    "5f43a12e71f9a42f55656cc0",  # Replace with actual system ID
    "5f43a1a271f9a42f55656cc1"   # Replace with actual system ID
  ]
  
  # Target system groups
  system_group_targets = [
    "5f43a22171f9a42f55656cc2"  # Replace with actual system group ID
  ]
}

# Windows updates policy with specific package IDs
resource "jumpcloud_software_update_policy" "windows_policy" {
  name        = "Windows Critical Updates"
  description = "Only apply critical Windows security updates"
  os_family   = "windows"
  enabled     = true
  
  # Schedule updates for the last Friday of each month at 10:00 PM
  schedule = jsonencode({
    type         = "monthly"
    dayOfMonth   = "last-friday"
    hour         = 22
    minute       = 0
    timeZone     = "America/New_York"
  })
  
  # Specify particular package IDs to update
  package_ids = [
    "5f43a30571f9a42f55656cc3",  # Replace with actual package ID
    "5f43a33d71f9a42f55656cc4"   # Replace with actual package ID
  ]
  
  # Require manual approval
  auto_approve = false
  
  # Target a system group
  system_group_targets = [
    "5f43a3a671f9a42f55656cc5"  # Replace with actual system group ID
  ]
}

# Linux server updates policy
resource "jumpcloud_software_update_policy" "linux_policy" {
  name        = "Linux Server Updates"
  description = "Security updates for Linux servers"
  os_family   = "linux"
  enabled     = true
  
  # Schedule updates to run daily at 3:00 AM
  schedule = jsonencode({
    type   = "daily"
    hour   = 3
    minute = 0
  })
  
  # Apply to all Linux packages
  all_packages = true
  
  # Require manual approval
  auto_approve = false
  
  # Target Linux servers system group
  system_group_targets = [
    "5f43a41b71f9a42f55656cc6"  # Replace with actual system group ID
  ]
} 