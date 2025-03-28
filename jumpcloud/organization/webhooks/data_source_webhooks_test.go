package webhooks_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	commonTesting "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
)

func TestAccJumpCloudDataSourceWebhooks_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceWebhooksConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_webhooks.all", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_webhooks.all", "webhooks.#"),
				),
			},
		},
	})
}

func TestAccJumpCloudDataSourceWebhooks_filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceWebhooksConfig_filtered(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_webhooks.filtered", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_webhooks.filtered", "webhooks.#"),
				),
			},
		},
	})
}

func TestAccJumpCloudDataSourceWebhooks_withWebhook(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { commonTesting.AccPreCheck(t) },
		ProviderFactories: commonTesting.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudDataSourceWebhooksConfig_withWebhook(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.jumpcloud_webhooks.with_webhook", "id"),
					resource.TestCheckResourceAttrSet("data.jumpcloud_webhooks.with_webhook", "webhooks.#"),
				),
			},
		},
	})
}

func testAccJumpCloudDataSourceWebhooksConfig_basic() string {
	return `
data "jumpcloud_webhooks" "all" {}
`
}

func testAccJumpCloudDataSourceWebhooksConfig_filtered() string {
	return `
data "jumpcloud_webhooks" "filtered" {
  filter {
    field = "status"
    operator = "eq"
    value = "active"
  }
  sort {
    field = "name"
    direction = "asc"
  }
}

output "webhook_count" {
  value = data.jumpcloud_webhooks.filtered.webhooks.#
}
`
}

func testAccJumpCloudDataSourceWebhooksConfig_withWebhook() string {
	return `
resource "jumpcloud_webhook" "test" {
  name = "test-webhook"
  url = "https://example.com/webhook"
  events = ["user.created", "user.updated"]
}

data "jumpcloud_webhooks" "with_webhook" {
  filter {
    field = "name"
    operator = "eq"
    value = jumpcloud_webhook.test.name
  }
}

output "webhook_count" {
  value = data.jumpcloud_webhooks.with_webhook.webhooks.#
}
`
}
