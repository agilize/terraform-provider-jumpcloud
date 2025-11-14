# Complete Organization Setup Example
# This example demonstrates a full organizational setup with users, groups,
# application mappings, and device associations

# Create users
resource "jumpcloud_user" "a01_user_test" {
  username   = "a01_user_test"
  email      = "a01_user_test@agilize.com.br"
  firstname  = "a01_user_test"
  lastname   = "a01_user_test"
  password   = "SecurePassword123!" # In production, use a secure method for passwords
}

resource "jumpcloud_user" "a02_user_test" {
  username   = "a02_user_test"
  email      = "a02_user_test@agilize.com.br"
  firstname  = "a02_user_test"
  lastname   = "a02_user_test"
  password   = "SecurePassword123!" # In production, use a secure method for passwords
}

resource "jumpcloud_user" "a03_user_test" {
  username   = "a03_user_test"
  email      = "a03_user_test@agilize.com.br"
  firstname  = "a03_user_test"
  lastname   = "a03_user_test"
  password   = "SecurePassword123!" # In production, use a secure method for passwords
}
