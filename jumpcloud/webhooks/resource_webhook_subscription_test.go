package webhooks

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceWebhookSubscriptionSchema(t *testing.T) {
	s := ResourceWebhookSubscription().Schema

	// Required fields
	for _, required := range []string{"webhook_id", "event_type"} {
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
	for _, optional := range []string{"description"} {
		if !s[optional].Optional {
			t.Errorf("Expected %s to be optional", optional)
		}
	}
}

func TestValidateWebhookSubscriptionEventTypes(t *testing.T) {
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

	// Define validation function for webhook subscription event types
	validateWebhookSubscriptionEventTypes := func() schema.SchemaValidateFunc {
		return func(v interface{}, k string) (ws []string, errors []error) {
			eventTypes := []string{}
			for _, item := range v.([]interface{}) {
				eventTypes = append(eventTypes, item.(string))
			}

			// Use the same validation function as regular webhooks
			err := ValidateEventTypes(eventTypes)
			if err != nil {
				errors = append(errors, fmt.Errorf("invalid event type: %s", err))
			}
			return
		}
	}

	for _, tc := range cases {
		_, errors := validateWebhookSubscriptionEventTypes()(tc.Value, "event_types")
		if tc.Expected && len(errors) > 0 {
			t.Errorf("Expected %v to be valid", tc.Value)
		}
		if !tc.Expected && len(errors) == 0 {
			t.Errorf("Expected %v to be invalid", tc.Value)
		}
	}
}

// Following are the acceptance tests that would require actual API calls
// These would be uncommented and implemented when setting up acceptance testing infrastructure

/*
func TestAccResourceWebhookSubscription_basic(t *testing.T) {
	var webhook WebhookSubscription
	resourceName := "jumpcloud_webhook_subscription.test"
	webhookResourceName := "jumpcloud_webhook.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWebhookSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWebhookSubscriptionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookSubscriptionExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "name", "test-webhook-subscription"),
					resource.TestCheckResourceAttrPair(resourceName, "webhook_id", webhookResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_types.0", "user.created"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceWebhookSubscription_update(t *testing.T) {
	var webhook WebhookSubscription
	resourceName := "jumpcloud_webhook_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWebhookSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWebhookSubscriptionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookSubscriptionExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "name", "test-webhook-subscription"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccResourceWebhookSubscriptionConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookSubscriptionExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "name", "test-webhook-subscription-updated"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccResourceWebhookSubscription_eventTypes(t *testing.T) {
	var webhook WebhookSubscription
	resourceName := "jumpcloud_webhook_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckWebhookSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWebhookSubscriptionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookSubscriptionExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "event_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_types.0", "user.created"),
				),
			},
			{
				Config: testAccResourceWebhookSubscriptionConfigMultipleEventTypes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebhookSubscriptionExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "event_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "event_types.0", "user.created"),
					resource.TestCheckResourceAttr(resourceName, "event_types.1", "system.connected"),
				),
			},
		},
	})
}

func testAccCheckWebhookSubscriptionExists(n string, webhook *WebhookSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		client := testAccProvider.Meta().(*apiclient.Client)
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/webhooks/subscriptions/%s", client.URL, rs.Primary.ID), nil)
		if err != nil {
			return err
		}

		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			return fmt.Errorf("Status code %d was not expected", res.StatusCode)
		}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(body, webhook)
		return err
	}
}

func testAccCheckWebhookSubscriptionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*apiclient.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "jumpcloud_webhook_subscription" {
			continue
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/webhooks/subscriptions/%s", client.URL, rs.Primary.ID), nil)
		if err != nil {
			return err
		}

		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != 404 {
			return fmt.Errorf("WebhookSubscription still exists")
		}
	}

	return nil
}

func testAccResourceWebhookSubscriptionConfig() string {
	return `
resource "jumpcloud_webhook" "test" {
  name = "test-webhook"
  url  = "https://example.com/webhook"
}

resource "jumpcloud_webhook_subscription" "test" {
  name        = "test-webhook-subscription"
  webhook_id  = jumpcloud_webhook.test.id
  enabled     = true
  event_types = ["user.created"]
}
`
}

func testAccResourceWebhookSubscriptionConfigUpdated() string {
	return `
resource "jumpcloud_webhook" "test" {
  name = "test-webhook"
  url  = "https://example.com/webhook"
}

resource "jumpcloud_webhook_subscription" "test" {
  name        = "test-webhook-subscription-updated"
  webhook_id  = jumpcloud_webhook.test.id
  enabled     = false
  description = "Updated description"
  event_types = ["user.created"]
}
`
}

func testAccResourceWebhookSubscriptionConfigMultipleEventTypes() string {
	return `
resource "jumpcloud_webhook" "test" {
  name = "test-webhook"
  url  = "https://example.com/webhook"
}

resource "jumpcloud_webhook_subscription" "test" {
  name        = "test-webhook-subscription"
  webhook_id  = jumpcloud_webhook.test.id
  enabled     = true
  event_types = ["user.created", "system.connected"]
}
`
}
*/
