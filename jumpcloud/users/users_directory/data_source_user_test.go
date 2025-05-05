package users_directory

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud"
)

// testAccPreCheck validates the necessary test API keys exist
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("JUMPCLOUD_API_KEY"); v == "" {
		t.Skip("JUMPCLOUD_API_KEY must be set for acceptance tests")
	}
}

// testAccProviders is a map of providers used for testing
var testAccProviders map[string]*schema.Provider

func init() {
	testAccProvider := jumpcloud.Provider()
	testAccProviders = map[string]*schema.Provider{
		"jumpcloud": testAccProvider,
	}
}

func TestAccDataSourceUser_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1 is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user.test", "id"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test", "username", "testuser"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test", "email", "testuser@example.com"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test", "firstname", "Test"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test", "lastname", "User"),
				),
			},
		},
	})
}

func testAccDataSourceUserConfig_basic() string {
	return fmt.Sprintf(`
resource "jumpcloud_user" "test" {
  username  = "testuser"
  email     = "testuser@example.com"
  firstname = "Test"
  lastname  = "User"
  password  = "TestPassword123!"
}

data "jumpcloud_user" "test" {
  username = jumpcloud_user.test.username
  depends_on = [jumpcloud_user.test]
}
`)
}

func TestAccDataSourceUser_byEmail(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1 is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserConfig_byEmail(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user.test_email", "id"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test_email", "email", "emailuser@example.com"),
				),
			},
		},
	})
}

func testAccDataSourceUserConfig_byEmail() string {
	return fmt.Sprintf(`
resource "jumpcloud_user" "test_email" {
  username  = "emailuser"
  email     = "emailuser@example.com"
  firstname = "Email"
  lastname  = "User"
  password  = "TestPassword123!"
}

data "jumpcloud_user" "test_email" {
  email = jumpcloud_user.test_email.email
  depends_on = [jumpcloud_user.test_email]
}
`)
}

func TestAccDataSourceUser_byID(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1 is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserConfig_byID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user.test_id", "id"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test_id", "username", "iduser"),
				),
			},
		},
	})
}

func testAccDataSourceUserConfig_byID() string {
	return fmt.Sprintf(`
resource "jumpcloud_user" "test_id" {
  username  = "iduser"
  email     = "iduser@example.com"
  firstname = "ID"
  lastname  = "User"
  password  = "TestPassword123!"
}

data "jumpcloud_user" "test_id" {
  user_id = jumpcloud_user.test_id.id
  depends_on = [jumpcloud_user.test_id]
}
`)
}

func TestAccDataSourceUser_complete(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless TF_ACC=1 is set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserConfig_complete(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_user.test_complete", "id"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test_complete", "username", "completeuser"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test_complete", "email", "completeuser@example.com"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test_complete", "firstname", "Complete"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test_complete", "lastname", "User"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test_complete", "company", "Test Company"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test_complete", "department", "IT"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test_complete", "attributes.department", "IT"),
					resource.TestCheckResourceAttr("data.jumpcloud_user.test_complete", "attributes.location", "Remote"),
				),
			},
		},
	})
}

func testAccDataSourceUserConfig_complete() string {
	return fmt.Sprintf(`
resource "jumpcloud_user" "test_complete" {
  username  = "completeuser"
  email     = "completeuser@example.com"
  firstname = "Complete"
  lastname  = "User"
  password  = "TestPassword123!"
  company   = "Test Company"
  department = "IT"

  attributes = {
    department = "IT"
    location   = "Remote"
  }
}

data "jumpcloud_user" "test_complete" {
  username = jumpcloud_user.test_complete.username
  depends_on = [jumpcloud_user.test_complete]
}
`)
}
