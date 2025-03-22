package iplist

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestResourceListAssignmentSchema testa o schema do recurso de atribuição de lista de IPs
func TestResourceListAssignmentSchema(t *testing.T) {
	s := ResourceListAssignment()
	if s.Schema["ip_list_id"] == nil {
		t.Error("Expected ip_list_id in schema, but it does not exist")
	}
	if s.Schema["ip_list_id"].Type != schema.TypeString {
		t.Error("Expected ip_list_id to be of type string")
	}
	if !s.Schema["ip_list_id"].Required {
		t.Error("Expected ip_list_id to be required")
	}
	if !s.Schema["ip_list_id"].ForceNew {
		t.Error("Expected ip_list_id to be ForceNew")
	}

	if s.Schema["resource_id"] == nil {
		t.Error("Expected resource_id in schema, but it does not exist")
	}
	if s.Schema["resource_id"].Type != schema.TypeString {
		t.Error("Expected resource_id to be of type string")
	}
	if !s.Schema["resource_id"].Required {
		t.Error("Expected resource_id to be required")
	}
	if !s.Schema["resource_id"].ForceNew {
		t.Error("Expected resource_id to be ForceNew")
	}

	if s.Schema["resource_type"] == nil {
		t.Error("Expected resource_type in schema, but it does not exist")
	}
	if s.Schema["resource_type"].Type != schema.TypeString {
		t.Error("Expected resource_type to be of type string")
	}
	if !s.Schema["resource_type"].Required {
		t.Error("Expected resource_type to be required")
	}
	if !s.Schema["resource_type"].ForceNew {
		t.Error("Expected resource_type to be ForceNew")
	}

	if s.Schema["org_id"] == nil {
		t.Error("Expected org_id in schema, but it does not exist")
	}
	if s.Schema["org_id"].Type != schema.TypeString {
		t.Error("Expected org_id to be of type string")
	}
	if s.Schema["org_id"].Required {
		t.Error("Expected org_id to be optional")
	}
	if !s.Schema["org_id"].ForceNew {
		t.Error("Expected org_id to be ForceNew")
	}

	if s.Schema["created"] == nil {
		t.Error("Expected created in schema, but it does not exist")
	}
	if s.Schema["created"].Type != schema.TypeString {
		t.Error("Expected created to be of type string")
	}
	if !s.Schema["created"].Computed {
		t.Error("Expected created to be computed")
	}

	if s.Schema["updated"] == nil {
		t.Error("Expected updated in schema, but it does not exist")
	}
	if s.Schema["updated"].Type != schema.TypeString {
		t.Error("Expected updated to be of type string")
	}
	if !s.Schema["updated"].Computed {
		t.Error("Expected updated to be computed")
	}
}

// TestResourceTypeValidation testa a validação do campo resource_type
func TestResourceTypeValidation(t *testing.T) {
	s := ResourceListAssignment()
	validateFunc := s.Schema["resource_type"].ValidateFunc

	for _, validType := range []string{"radius_server", "ldap_server", "system", "system_group", "organization", "application", "directory"} {
		_, errs := validateFunc(validType, "resource_type")
		if len(errs) > 0 {
			t.Errorf("Expected %s to be a valid resource_type, but got errors: %v", validType, errs)
		}
	}

	_, errs := validateFunc("invalid_type", "resource_type")
	if len(errs) == 0 {
		t.Error("Expected 'invalid_type' to be an invalid resource_type, but no errors were returned")
	}
}

// TestAccJumpCloudIPListAssignment_basic tests creating and deleting an IP list assignment
func TestAccJumpCloudIPListAssignment_basic(t *testing.T) {
	t.Skip("Skipping acceptance test until CI environment is set up")

	t.Run("basic", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: `
resource "jumpcloud_ip_list" "test" {
  name        = "test_ip_list"
  description = "Test IP list"
  type        = "allow"
  ips = [
    {
      address     = "192.168.1.1"
      description = "Test IP"
    },
    {
      address     = "10.0.0.0/24"
      description = "Test CIDR"
    }
  ]
}

resource "jumpcloud_ip_list_assignment" "test" {
  ip_list_id    = jumpcloud_ip_list.test.id
  resource_id   = "5f1b7e6e1e6e1e6e1e6e1e6e" # This would be a real resource ID in actual tests
  resource_type = "radius_server"
}
`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"jumpcloud_ip_list_assignment.test", "ip_list_id",
							"jumpcloud_ip_list.test", "id",
						),
						resource.TestCheckResourceAttr("jumpcloud_ip_list_assignment.test", "resource_type", "radius_server"),
					),
				},
			},
		})
	})
}
