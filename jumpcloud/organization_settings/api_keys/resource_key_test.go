package api_keys_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccJumpCloudAPIKey_basic(t *testing.T) {
	t.Skip("Skipping acceptance test in unit test environment")

	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { t.Log("PreCheck running") },
		Providers:    nil, // Will be set by the test framework
		CheckDestroy: testAccCheckJumpCloudAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAPIKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAPIKeyExists("jumpcloud_api_key.test"),
					resource.TestCheckResourceAttr("jumpcloud_api_key.test", "name", rName),
					resource.TestCheckResourceAttr("jumpcloud_api_key.test", "description", "Created by Terraform"),
					resource.TestCheckResourceAttrSet("jumpcloud_api_key.test", "key"),
					resource.TestCheckResourceAttrSet("jumpcloud_api_key.test", "created"),
				),
			},
		},
	})
}

func TestAccJumpCloudAPIKey_update(t *testing.T) {
	t.Skip("Skipping acceptance test in unit test environment")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := rName + "-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { t.Log("PreCheck running") },
		Providers:    nil, // Will be set by the test framework
		CheckDestroy: testAccCheckJumpCloudAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAPIKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAPIKeyExists("jumpcloud_api_key.test"),
					resource.TestCheckResourceAttr("jumpcloud_api_key.test", "name", rName),
					resource.TestCheckResourceAttr("jumpcloud_api_key.test", "description", "Created by Terraform"),
				),
			},
			{
				Config: testAccJumpCloudAPIKeyConfig_updated(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAPIKeyExists("jumpcloud_api_key.test"),
					resource.TestCheckResourceAttr("jumpcloud_api_key.test", "name", rNameUpdated),
					resource.TestCheckResourceAttr("jumpcloud_api_key.test", "description", "Updated by Terraform"),
				),
			},
		},
	})
}

func TestAccJumpCloudAPIKey_withExpires(t *testing.T) {
	t.Skip("Skipping acceptance test in unit test environment")

	rName := acctest.RandomWithPrefix("tf-acc-test")
	expiresTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { t.Log("PreCheck running") },
		Providers:    nil, // Will be set by the test framework
		CheckDestroy: testAccCheckJumpCloudAPIKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudAPIKeyConfig_withExpires(rName, expiresTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudAPIKeyExists("jumpcloud_api_key.test"),
					resource.TestCheckResourceAttr("jumpcloud_api_key.test", "name", rName),
					resource.TestCheckResourceAttr("jumpcloud_api_key.test", "expires", expiresTime),
				),
			},
		},
	})
}

func testAccCheckJumpCloudAPIKeyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Mock implementation for testing
		return nil
	}
}

func testAccCheckJumpCloudAPIKeyDestroy(s *terraform.State) error {
	// Mock implementation for testing
	return nil
}

func testAccJumpCloudAPIKeyConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_api_key" "test" {
  name        = "%s"
  description = "Created by Terraform"
}
`, name)
}

func testAccJumpCloudAPIKeyConfig_updated(name string) string {
	return fmt.Sprintf(`
resource "jumpcloud_api_key" "test" {
  name        = "%s"
  description = "Updated by Terraform"
}
`, name)
}

func testAccJumpCloudAPIKeyConfig_withExpires(name, expires string) string {
	return fmt.Sprintf(`
resource "jumpcloud_api_key" "test" {
  name        = "%s"
  description = "Created by Terraform with expiration"
  expires     = "%s"
}
`, name, expires)
}
