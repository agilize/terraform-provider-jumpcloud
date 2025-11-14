# Complete Organization Setup Example
# This example demonstrates a full organizational setup with users, groups,
# application mappings, and device associations


# Assign users to department groups
resource "jumpcloud_user_group_membership" "a01_user_group_test_membership" {
  user_group_id = jumpcloud_user_group.a01_user_group_test.id
  user_id       = jumpcloud_user.a01_user_test.id
}

resource "jumpcloud_user_group_membership" "a02_user_group_test_membership" {
  user_group_id = jumpcloud_user_group.a02_user_group_test.id
  user_id       = jumpcloud_user.a02_user_test.id
}

resource "jumpcloud_user_group_membership" "a03_user_group_test_membership" {
  user_group_id = jumpcloud_user_group.a03_user_group_test.id
  user_id       = jumpcloud_user.a03_user_test.id
}

resource "jumpcloud_user_group_membership" "a04_user_group_test_membership" {
  user_group_id = jumpcloud_user_group.a04_user_group_test.id
  user_id       = jumpcloud_user.a01_user_test.id
}

resource "jumpcloud_user_group_membership" "a05_user_group_test_membership" {
  user_group_id = jumpcloud_user_group.a04_user_group_test.id
  user_id       = jumpcloud_user.a02_user_test.id
}

resource "jumpcloud_user_group_membership" "a06_user_group_test_membership" {
  user_group_id = jumpcloud_user_group.a04_user_group_test.id
  user_id       = jumpcloud_user.a03_user_test.id
}
