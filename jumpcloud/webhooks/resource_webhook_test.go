package webhooks

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	jctest "registry.terraform.io/agilize/jumpcloud/jumpcloud/testing"
)

// TestResourceWebhookSchema tests the schema structure of the webhook resource
func TestResourceWebhookSchema(t *testing.T) {
	s := ResourceWebhook().Schema

	// Required fields
	for _, required := range []string{"name", "url"} {
		if !s[required].Required {
			t.Errorf("Expected %s to be required", required)
		}
	}

	// Computed fields
	for _, computed := range []string{"id", "created", "updated"} {
		if !s[computed].Computed {
			t.Errorf("Expected %s to be computed", computed)
		}
	}

	// Optional fields
	for _, optional := range []string{"secret", "enabled", "event_types", "description"} {
		if !s[optional].Optional {
			t.Errorf("Expected %s to be optional", optional)
		}
	}
}

// Test helper functions
func TestValidateEventTypes(t *testing.T) {
	cases := []struct {
		Value    interface{}
		Expected bool
	}{
		{[]interface{}{"user.created"}, true},
		{[]interface{}{"system.connected"}, true},
		{[]interface{}{"invalid_event"}, false},
		{[]interface{}{"user.created", "system.connected"}, true},
		{[]interface{}{"user.created", "invalid_event"}, false},
	}

	// Define validation function since it's not exported from the resource file
	validateEventTypes := func() schema.SchemaValidateFunc {
		return func(v interface{}, k string) (ws []string, errors []error) {
			eventTypes := []string{}
			for _, item := range v.([]interface{}) {
				eventTypes = append(eventTypes, item.(string))
			}

			err := ValidateEventTypes(eventTypes)
			if err != nil {
				errors = append(errors, fmt.Errorf("invalid event type: %s", err))
			}
			return
		}
	}

	for _, tc := range cases {
		_, errors := validateEventTypes()(tc.Value, "event_types")
		if tc.Expected && len(errors) > 0 {
			t.Errorf("Expected %v to be valid", tc.Value)
		}
		if !tc.Expected && len(errors) == 0 {
			t.Errorf("Expected %v to be invalid", tc.Value)
		}
	}
}

// Acceptance testing
func TestAccResourceWebhook_basic(t *testing.T) {
	resourceName := "jumpcloud_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudWebhookConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudWebhookExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-webhook"),
					resource.TestCheckResourceAttr(resourceName, "url", "https://example.com/webhook"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceWebhook_update(t *testing.T) {
	resourceName := "jumpcloud_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudWebhookConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudWebhookExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-webhook"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccJumpCloudWebhookConfig_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudWebhookExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "updated-test-webhook"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated test webhook"),
				),
			},
		},
	})
}

func TestAccResourceWebhook_eventTypes(t *testing.T) {
	resourceName := "jumpcloud_webhook.test_event_types"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { jctest.TestAccPreCheck(t) },
		Providers:    jctest.TestAccProviders,
		CheckDestroy: testAccCheckJumpCloudWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudWebhookConfig_eventTypes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudWebhookExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "test-webhook-events"),
					resource.TestCheckResourceAttr(resourceName, "event_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "event_types.0", "user.created"),
					resource.TestCheckResourceAttr(resourceName, "event_types.1", "user.updated"),
				),
			},
		},
	})
}

func testAccCheckJumpCloudWebhookDestroy(s *terraform.State) error {
	// Implementation would verify that the resource is deleted on the API side
	// This is a placeholder for the actual implementation
	return nil
}

func testAccCheckJumpCloudWebhookExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Webhook ID is set")
		}

		return nil
	}
}

func testAccJumpCloudWebhookConfig_basic() string {
	return `
resource "jumpcloud_webhook" "test" {
  name        = "test-webhook"
  url         = "https://example.com/webhook"
  enabled     = true
  description = "Test webhook"
}
`
}

func testAccJumpCloudWebhookConfig_update() string {
	return `
resource "jumpcloud_webhook" "test" {
  name        = "updated-test-webhook"
  url         = "https://example.com/webhook"
  enabled     = false
  description = "Updated test webhook"
}
`
}

func testAccJumpCloudWebhookConfig_eventTypes() string {
	return `
resource "jumpcloud_webhook" "test_event_types" {
  name        = "test-webhook-events"
  url         = "https://example.com/webhook-events"
  enabled     = true
  description = "Test webhook with event types"
  
  event_types = [
    "user.created",
    "user.updated"
  ]
}
`
}
