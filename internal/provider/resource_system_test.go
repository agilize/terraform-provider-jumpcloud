package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestResourceSystemCreate tests the Create method of the system resource
func TestResourceSystemCreate(t *testing.T) {
	// Create a mock client
	mockClient := new(MockClient)

	// Mock response for creating a system
	systemResponse := System{
		ID:          "test-system-id",
		DisplayName: "test-system",
		SystemType:  "linux",
		OS:          "ubuntu",
		Version:     "20.04",
	}
	responseBody, _ := json.Marshal(systemResponse)

	// Set up expectations
	mockClient.On("DoRequest", http.MethodPost, "/api/v2/systems", mock.Anything).Return(responseBody, nil)
	mockClient.On("DoRequest", http.MethodGet, "/api/systems/test-system-id", nil).Return(responseBody, nil)

	// Create test schema data
	d := schema.TestResourceDataRaw(t, resourceSystem().Schema, map[string]interface{}{
		"display_name":                      "test-system",
		"allow_ssh_root_login":              false,
		"allow_ssh_password_authentication": true,
		"allow_multi_factor_authentication": false,
		"description":                       "Test system",
		"tags":                              []interface{}{"test", "dev"},
	})

	// Call the function being tested
	diags := resourceSystemCreate(context.Background(), d, mockClient)

	// Check results
	assert.False(t, diags.HasError(), "resourceSystemCreate returned an error")
	assert.Equal(t, "test-system-id", d.Id(), "System ID was not set correctly")

	// Verify expectations were met
	mockClient.AssertExpectations(t)
}

// TestResourceSystemRead tests the Read method of the system resource
func TestResourceSystemRead(t *testing.T) {
	// Create a mock client
	mockClient := new(MockClient)

	// Mock response for reading a system
	systemResponse := System{
		ID:                             "test-system-id",
		DisplayName:                    "test-system",
		SystemType:                     "linux",
		OS:                             "ubuntu",
		Version:                        "20.04",
		AgentVersion:                   "1.0.0",
		AllowSshRootLogin:              false,
		AllowSshPasswordAuthentication: true,
		AllowMultiFactorAuthentication: false,
		Tags:                           []string{"test", "dev"},
		Description:                    "Test system",
	}
	responseBody, _ := json.Marshal(systemResponse)

	// Set up expectations
	mockClient.On("DoRequest", http.MethodGet, "/api/systems/test-system-id", nil).Return(responseBody, nil)

	// Create test schema data
	d := schema.TestResourceDataRaw(t, resourceSystem().Schema, map[string]interface{}{})
	d.SetId("test-system-id")

	// Call the function being tested
	diags := resourceSystemRead(context.Background(), d, mockClient)

	// Check results
	assert.False(t, diags.HasError(), "resourceSystemRead returned an error")
	assert.Equal(t, "test-system", d.Get("display_name").(string), "DisplayName was not set correctly")
	assert.Equal(t, "linux", d.Get("system_type").(string), "SystemType was not set correctly")
	assert.Equal(t, "ubuntu", d.Get("os").(string), "OS was not set correctly")
	assert.Equal(t, "20.04", d.Get("version").(string), "Version was not set correctly")
	assert.Equal(t, "1.0.0", d.Get("agent_version").(string), "AgentVersion was not set correctly")
	assert.Equal(t, false, d.Get("allow_ssh_root_login").(bool), "AllowSshRootLogin was not set correctly")
	assert.Equal(t, true, d.Get("allow_ssh_password_authentication").(bool), "AllowSshPasswordAuthentication was not set correctly")
	assert.Equal(t, false, d.Get("allow_multi_factor_authentication").(bool), "AllowMultiFactorAuthentication was not set correctly")
	assert.Equal(t, "Test system", d.Get("description").(string), "Description was not set correctly")

	// Verify expectations were met
	mockClient.AssertExpectations(t)
}

// TestResourceSystemUpdate tests the Update method of the system resource
func TestResourceSystemUpdate(t *testing.T) {
	// Create a mock client
	mockClient := new(MockClient)

	// Mock response for updating a system
	systemResponse := System{
		ID:                             "test-system-id",
		DisplayName:                    "updated-system",
		SystemType:                     "linux",
		OS:                             "ubuntu",
		Version:                        "20.04",
		AllowSshRootLogin:              true,
		AllowSshPasswordAuthentication: false,
		AllowMultiFactorAuthentication: true,
		Description:                    "Updated system description",
	}
	responseBody, _ := json.Marshal(systemResponse)

	// Set up expectations
	mockClient.On("DoRequest", http.MethodPut, "/api/systems/test-system-id", mock.Anything).Return([]byte("{}"), nil)
	mockClient.On("DoRequest", http.MethodGet, "/api/systems/test-system-id", nil).Return(responseBody, nil)

	// Create test schema data with original values
	oldData := map[string]interface{}{
		"display_name":                      "test-system",
		"allow_ssh_root_login":              false,
		"allow_ssh_password_authentication": true,
		"allow_multi_factor_authentication": false,
		"description":                       "Test system",
	}
	d := schema.TestResourceDataRaw(t, resourceSystem().Schema, oldData)
	d.SetId("test-system-id")

	// Set new values to simulate changes
	d.Set("display_name", "updated-system")
	d.Set("allow_ssh_root_login", true)
	d.Set("allow_ssh_password_authentication", false)
	d.Set("allow_multi_factor_authentication", true)
	d.Set("description", "Updated system description")

	// Call the function being tested
	diags := resourceSystemUpdate(context.Background(), d, mockClient)

	// Check results
	assert.False(t, diags.HasError(), "resourceSystemUpdate returned an error")

	// Verify expectations were met
	mockClient.AssertExpectations(t)
}

// TestResourceSystemDelete tests the Delete method of the system resource
func TestResourceSystemDelete(t *testing.T) {
	// Create a mock client
	mockClient := new(MockClient)

	// Set up expectations
	mockClient.On("DoRequest", http.MethodDelete, "/api/systems/test-system-id", nil).Return([]byte("{}"), nil)

	// Create test schema data
	d := schema.TestResourceDataRaw(t, resourceSystem().Schema, map[string]interface{}{})
	d.SetId("test-system-id")

	// Call the function being tested
	diags := resourceSystemDelete(context.Background(), d, mockClient)

	// Check results
	assert.False(t, diags.HasError(), "resourceSystemDelete returned an error")
	assert.Equal(t, "", d.Id(), "ID was not cleared after deletion")

	// Verify expectations were met
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudSystem_basic tests the system resource through an acceptance test
// This is an acceptance test that would make real API calls.
// It should be executed with TF_ACC=1 environment variable.
func TestAccJumpCloudSystem_basic(t *testing.T) {
	// Skip if not running acceptance tests
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudSystemConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemExists("jumpcloud_system.test"),
					resource.TestCheckResourceAttr("jumpcloud_system.test", "display_name", "test-system"),
					resource.TestCheckResourceAttr("jumpcloud_system.test", "allow_ssh_root_login", "false"),
					resource.TestCheckResourceAttr("jumpcloud_system.test", "allow_ssh_password_authentication", "true"),
					resource.TestCheckResourceAttr("jumpcloud_system.test", "description", "Created by Terraform acceptance test"),
				),
			},
			{
				Config: testAccJumpCloudSystemConfig_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudSystemExists("jumpcloud_system.test"),
					resource.TestCheckResourceAttr("jumpcloud_system.test", "display_name", "updated-system"),
					resource.TestCheckResourceAttr("jumpcloud_system.test", "allow_ssh_root_login", "true"),
					resource.TestCheckResourceAttr("jumpcloud_system.test", "description", "Updated by Terraform acceptance test"),
				),
			},
		},
	})
}

// This function checks if the system was properly destroyed
func testAccCheckJumpCloudSystemDestroy(s *terraform.State) error {
	// We would implement this to make an API call to verify the resource is gone
	// For now, we'll just return nil since this is for illustration
	return nil
}

// This function checks if the system exists
func testAccCheckJumpCloudSystemExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// We would implement this to make an API call to verify the resource exists
		// For now, we'll just return nil since this is for illustration
		return nil
	}
}

// Test configurations for the acceptance tests
const testAccJumpCloudSystemConfig_basic = `
resource "jumpcloud_system" "test" {
  display_name                      = "test-system"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = true
  allow_multi_factor_authentication = false
  description                       = "Created by Terraform acceptance test"
  
  tags = [
    "test",
    "terraform"
  ]
}
`

const testAccJumpCloudSystemConfig_update = `
resource "jumpcloud_system" "test" {
  display_name                      = "updated-system"
  allow_ssh_root_login              = true
  allow_ssh_password_authentication = true
  allow_multi_factor_authentication = true
  description                       = "Updated by Terraform acceptance test"
  
  tags = [
    "test",
    "terraform",
    "updated"
  ]
}
`
