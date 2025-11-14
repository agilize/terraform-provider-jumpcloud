# Complete Organization Setup Example
# This example demonstrates a full organizational setup with users, groups,
# application mappings, and device associations

terraform {
  required_providers {
    jumpcloud = {
      source  = "agilize/jumpcloud"
    }
  }
}

provider "jumpcloud" {
}