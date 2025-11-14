# Complete Organization Setup Example
# This example demonstrates a full organizational setup with users, groups,
# application mappings, and device associations


# ============================================================================
# APPLICATION MAPPINGS - GROUP LEVEL
# ============================================================================

# Data source to find application by ID
# Try using one of the application IDs from the list above
# For example, using Slack:
data "jumpcloud_application_sso_application" "a01_app" {
  id = "6807961f7442689dcfe38c93"  # Slack application
}

resource "jumpcloud_user_group_application_association" "a01_usergroup_apps_mappings" {
  user_group_id  = jumpcloud_user_group.a01_user_group_test.id
  application_id = data.jumpcloud_application_sso_application.a01_app.id
}

resource "jumpcloud_user_group_application_association" "a02_usergroup_apps_mappings" {
  user_group_id  = jumpcloud_user_group.a02_user_group_test.id
  application_id = data.jumpcloud_application_sso_application.a01_app.id
}

resource "jumpcloud_user_group_application_association" "a03_usergroup_apps_mappings" {
  user_group_id  = jumpcloud_user_group.a03_user_group_test.id
  application_id = data.jumpcloud_application_sso_application.a01_app.id
}

resource "jumpcloud_user_group_application_association" "a04_usergroup_apps_mappings" {
  user_group_id  = jumpcloud_user_group.a04_user_group_test.id
  application_id = data.jumpcloud_application_sso_application.a01_app.id
}
