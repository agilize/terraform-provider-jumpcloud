provider "jumpcloud" {
  # API key can be provided via JUMPCLOUD_API_KEY environment variable
}

# Create a basic user group
resource "jumpcloud_user_group" "minimal" {
  name        = "minimal-test-group"
  description = "A minimal test user group"
}

# Create a basic user
resource "jumpcloud_user" "minimal" {
  username  = "minimal-test-user"
  email     = "minimal-test-user@example.com"
  firstname = "Minimal"
  lastname  = "User"
  password  = "ComplexP@ssw0rd!"
}

# Add the user to the group
resource "jumpcloud_user_group_membership" "minimal" {
  user_id  = jumpcloud_user.minimal.id
  group_id = jumpcloud_user_group.minimal.id
}

# Retrieve the user group using the data source
data "jumpcloud_user_group" "minimal" {
  group_id = jumpcloud_user_group.minimal.id
  depends_on = [jumpcloud_user_group.minimal]
}

# Output the user group details
output "user_group_id" {
  value = data.jumpcloud_user_group.minimal.id
}

output "user_group_name" {
  value = data.jumpcloud_user_group.minimal.name
}

output "user_group_description" {
  value = data.jumpcloud_user_group.minimal.description
}
