# Complete Organization Setup Example
# This example demonstrates a full organizational setup with users, groups,
# application mappings, and device associations

# Create departments as user groups
resource "jumpcloud_user_group" "a01_user_group_test" {
  name        = "a01_user_group_test"
  description = "a01_user_group_test"

  attributes = {
    department = "a01_user_group_test"
    location   = "a01_user_group_test"
  }
}

resource "jumpcloud_user_group" "a02_user_group_test" {
  name        = "a02_user_group_test"
  description = "a02_user_group_test"

  attributes = {
    department = "a02_user_group_test"
    location   = "a02_user_group_test"
  }
}

resource "jumpcloud_user_group" "a03_user_group_test" {
  name        = "a03_user_group_test"
  description = "a03_user_group_test"

  attributes = {
    department = "a03_user_group_test"
    location   = "a03_user_group_test"
  }
}

resource "jumpcloud_user_group" "a04_user_group_test" {
  name        = "a04_user_group_test"
  description = "a04_user_group_test"

  attributes = {
    department = "a04_user_group_test"
    location   = "a04_user_group_test"
  }
}