package webhooks

import (
	"testing"
)

func TestDataSourceWebhookSchema(t *testing.T) {
	s := DataSourceWebhook().Schema

	// ID or name is required for lookup
	if !s["id"].Optional {
		t.Errorf("Expected 'id' to be optional")
	}
	if !s["name"].Optional {
		t.Errorf("Expected 'name' to be optional")
	}

	// Computed fields
	for _, computed := range []string{"url", "enabled", "event_types", "description", "created", "updated"} {
		if !s[computed].Computed {
			t.Errorf("Expected %s to be computed", computed)
		}
	}
}

// Following are the acceptance tests that would require actual API calls
// These would be uncommented and implemented when setting up acceptance testing infrastructure

/*
func TestAccDataSourceWebhook_basic(t *testing.T) {
	resourceName := "jumpcloud_webhook.test"
	dataSourceName := "data.jumpcloud_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWebhookConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "url", resourceName, "url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "event_types.#", resourceName, "event_types.#"),
				),
			},
		},
	})
}

func TestAccDataSourceWebhook_byName(t *testing.T) {
	resourceName := "jumpcloud_webhook.test"
	dataSourceName := "data.jumpcloud_webhook.by_name"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWebhookByNameConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "url", resourceName, "url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "event_types.#", resourceName, "event_types.#"),
				),
			},
		},
	})
}

func testAccDataSourceWebhookConfig() string {
	return `
resource "jumpcloud_webhook" "test" {
  name        = "test-webhook"
  url         = "https://example.com/webhook"
  enabled     = true
  event_types = ["user.created", "system.connected"]
}

data "jumpcloud_webhook" "test" {
  id = jumpcloud_webhook.test.id
}
`
}

func testAccDataSourceWebhookByNameConfig() string {
	return `
resource "jumpcloud_webhook" "test" {
  name        = "test-webhook"
  url         = "https://example.com/webhook"
  enabled     = true
  event_types = ["user.created", "system.connected"]
}

data "jumpcloud_webhook" "by_name" {
  name = jumpcloud_webhook.test.name
}
`
}
*/
