# User-Device Association Example
# This example demonstrates associating users with devices in JumpCloud
# User-device associations grant users access to specific devices

# ============================================================================
# USER-DEVICE ASSOCIATIONS
# ============================================================================

# NOTE: These resources are commented out because they depend on device resources,
# which cannot be created via the API. Devices must be registered by installing
# the JumpCloud agent. Once you have imported devices into Terraform, you can
# uncomment these resources to manage user-device associations.

# # Associate users with production servers
# resource "jumpcloud_user_device_association" "uda01_user1_to_prod_server1" {
#   user_id   = jumpcloud_user.a01_user_test.id
#   device_id = jumpcloud_devices.d01_device_test.id
# }
#
# resource "jumpcloud_user_device_association" "uda02_user1_to_prod_server2" {
#   user_id   = jumpcloud_user.a01_user_test.id
#   device_id = jumpcloud_devices.d02_device_test.id
# }
#
# # Associate multiple users with the same device
# resource "jumpcloud_user_device_association" "uda03_user2_to_prod_server1" {
#   user_id   = jumpcloud_user.a02_user_test.id
#   device_id = jumpcloud_devices.d01_device_test.id
# }
#
# resource "jumpcloud_user_device_association" "uda04_user3_to_prod_server1" {
#   user_id   = jumpcloud_user.a03_user_test.id
#   device_id = jumpcloud_devices.d01_device_test.id
# }
#
# # Associate user with development workstation
# resource "jumpcloud_user_device_association" "uda05_user2_to_dev_workstation" {
#   user_id   = jumpcloud_user.a02_user_test.id
#   device_id = jumpcloud_devices.d03_device_test.id
# }
#
# resource "jumpcloud_user_device_association" "uda06_user3_to_dev_workstation" {
#   user_id   = jumpcloud_user.a03_user_test.id
#   device_id = jumpcloud_devices.d03_device_test.id
# }

# ============================================================================
# DATA SOURCES - Query existing user-device associations
# ============================================================================

# Note: Data sources require existing associations in your JumpCloud account
# Uncomment the examples below after creating the associations above

# # Query a specific user-device association
# data "jumpcloud_user_device_association" "example_association" {
#   user_id   = jumpcloud_user.a01_user_test.id
#   device_id = jumpcloud_devices.d01_device_test.id
# }

# # Output association information
# output "user_device_association_info" {
#   value = {
#     id        = data.jumpcloud_user_device_association.example_association.id
#     user_id   = data.jumpcloud_user_device_association.example_association.user_id
#     device_id = data.jumpcloud_user_device_association.example_association.device_id
#   }
# }

# ============================================================================
# NOTES
# ============================================================================
# - User-device associations grant users access to specific devices
# - A user can be associated with multiple devices
# - A device can have multiple users associated with it
# - Removing an association revokes the user's access to the device
# - Associations are independent of user groups and device groups
# - Use associations to control which users can access which devices

