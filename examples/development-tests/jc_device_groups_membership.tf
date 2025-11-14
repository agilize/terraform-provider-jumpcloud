# Device Group Membership Example
# This example demonstrates assigning devices to device groups in JumpCloud
# Device group memberships allow you to organize devices into logical groups

# ============================================================================
# DEVICE GROUP MEMBERSHIPS
# ============================================================================

# NOTE: These resources are commented out because they depend on device resources,
# which cannot be created via the API. Devices must be registered by installing
# the JumpCloud agent. Once you have imported devices into Terraform, you can
# uncomment these resources to manage device group memberships.

# # Assign devices to production servers group
# resource "jumpcloud_devices_group_membership" "dgm01_prod_server_1" {
#   device_group_id = jumpcloud_devices_group.dg01_production_servers.id
#   device_id       = jumpcloud_devices.d01_device_test.id
# }
#
# resource "jumpcloud_devices_group_membership" "dgm02_prod_server_2" {
#   device_group_id = jumpcloud_devices_group.dg01_production_servers.id
#   device_id       = jumpcloud_devices.d02_device_test.id
# }
#
# # Assign devices to development workstations group
# resource "jumpcloud_devices_group_membership" "dgm03_dev_workstation_1" {
#   device_group_id = jumpcloud_devices_group.dg02_development_workstations.id
#   device_id       = jumpcloud_devices.d03_device_test.id
# }
#
# # Assign devices to QA devices group
# resource "jumpcloud_devices_group_membership" "dgm04_qa_device_1" {
#   device_group_id = jumpcloud_devices_group.dg03_qa_devices.id
#   device_id       = jumpcloud_devices.d01_device_test.id
# }
#
# # Assign multiple devices to the same group (Linux servers)
# resource "jumpcloud_devices_group_membership" "dgm05_linux_server_1" {
#   device_group_id = jumpcloud_devices_group.dg04_all_linux_servers.id
#   device_id       = jumpcloud_devices.d01_device_test.id
# }
#
# resource "jumpcloud_devices_group_membership" "dgm06_linux_server_2" {
#   device_group_id = jumpcloud_devices_group.dg04_all_linux_servers.id
#   device_id       = jumpcloud_devices.d02_device_test.id
# }
#
# resource "jumpcloud_devices_group_membership" "dgm07_linux_server_3" {
#   device_group_id = jumpcloud_devices_group.dg04_all_linux_servers.id
#   device_id       = jumpcloud_devices.d03_device_test.id
# }

# ============================================================================
# NOTES
# ============================================================================
# - A device can be a member of multiple device groups
# - Device group memberships are managed independently from the devices themselves
# - Removing a membership does not delete the device or the group
# - Use device groups to apply policies, commands, or access controls to multiple devices

