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

// TestResourceUserCreate tests the Create method of the user resource
func TestResourceUserCreate(t *testing.T) {
	// Create a mock client
	mockClient := new(MockClient)

	// Mock response for creating a user
	userResponse := User{
		ID:       "test-id",
		Username: "test.user",
		Email:    "test.user@example.com",
	}
	responseBody, _ := json.Marshal(userResponse)

	// Set up expectations
	mockClient.On("DoRequest", http.MethodPost, "/api/systemusers", mock.Anything).Return(responseBody, nil)
	mockClient.On("DoRequest", http.MethodGet, "/api/systemusers/test-id", []byte(nil)).Return(responseBody, nil)

	// Create test schema data
	d := schema.TestResourceDataRaw(t, resourceUser().Schema, map[string]interface{}{
		"username":  "test.user",
		"email":     "test.user@example.com",
		"password":  "securePassword123!",
		"firstname": "Test",
		"lastname":  "User",
	})

	// Call the function being tested
	diags := resourceUserCreate(context.Background(), d, mockClient)

	// Check results
	assert.False(t, diags.HasError(), "resourceUserCreate returned an error")
	assert.Equal(t, "test-id", d.Id(), "User ID was not set correctly")

	// Verify expectations were met
	mockClient.AssertExpectations(t)
}

// TestResourceUserRead tests the Read method of the user resource
func TestResourceUserRead(t *testing.T) {
	// Create a mock client
	mockClient := new(MockClient)

	// Mock response for reading a user
	userResponse := User{
		ID:          "test-id",
		Username:    "test.user",
		Email:       "test.user@example.com",
		FirstName:   "Test",
		LastName:    "User",
		Description: "Test user",
	}
	responseBody, _ := json.Marshal(userResponse)

	// Set up expectations
	mockClient.On("DoRequest", http.MethodGet, "/api/systemusers/test-id", []byte(nil)).Return(responseBody, nil)

	// Create test schema data
	d := schema.TestResourceDataRaw(t, resourceUser().Schema, map[string]interface{}{})
	d.SetId("test-id")

	// Call the function being tested
	diags := resourceUserRead(context.Background(), d, mockClient)

	// Check results
	assert.False(t, diags.HasError(), "resourceUserRead returned an error")
	assert.Equal(t, "test.user", d.Get("username").(string), "Username was not set correctly")
	assert.Equal(t, "test.user@example.com", d.Get("email").(string), "Email was not set correctly")
	assert.Equal(t, "Test", d.Get("firstname").(string), "Firstname was not set correctly")
	assert.Equal(t, "User", d.Get("lastname").(string), "Lastname was not set correctly")
	assert.Equal(t, "Test user", d.Get("description").(string), "Description was not set correctly")

	// Verify expectations were met
	mockClient.AssertExpectations(t)
}

// TestResourceUserUpdate tests the Update method of the user resource
func TestResourceUserUpdate(t *testing.T) {
	// Create a mock client
	mockClient := new(MockClient)

	// Mock response for updating a user
	userResponse := User{
		ID:          "test-id",
		Username:    "test.user",
		Email:       "updated.email@example.com",
		FirstName:   "Updated",
		LastName:    "User",
		Description: "Updated description",
	}
	responseBody, _ := json.Marshal(userResponse)

	// Set up expectations
	mockClient.On("DoRequest", http.MethodPut, "/api/systemusers/test-id", mock.Anything).Return([]byte("{}"), nil)
	mockClient.On("DoRequest", http.MethodGet, "/api/systemusers/test-id", []byte(nil)).Return(responseBody, nil)

	// Create test schema data with original values
	oldData := map[string]interface{}{
		"username":    "test.user",
		"email":       "test.user@example.com",
		"firstname":   "Test",
		"lastname":    "User",
		"description": "Test user",
	}
	d := schema.TestResourceDataRaw(t, resourceUser().Schema, oldData)
	d.SetId("test-id")

	// Set new values to simulate changes
	d.Set("email", "updated.email@example.com")
	d.Set("firstname", "Updated")
	d.Set("description", "Updated description")

	// Call the function being tested
	diags := resourceUserUpdate(context.Background(), d, mockClient)

	// Check results
	assert.False(t, diags.HasError(), "resourceUserUpdate returned an error")

	// Verify expectations were met
	mockClient.AssertExpectations(t)
}

// TestResourceUserDelete tests the Delete method of the user resource
func TestResourceUserDelete(t *testing.T) {
	// Create a mock client
	mockClient := new(MockClient)

	// Set up expectations
	mockClient.On("DoRequest", http.MethodDelete, "/api/systemusers/test-id", []byte(nil)).Return([]byte("{}"), nil)

	// Create test schema data
	d := schema.TestResourceDataRaw(t, resourceUser().Schema, map[string]interface{}{})
	d.SetId("test-id")

	// Call the function being tested
	diags := resourceUserDelete(context.Background(), d, mockClient)

	// Check results
	assert.False(t, diags.HasError(), "resourceUserDelete returned an error")
	assert.Equal(t, "", d.Id(), "ID was not cleared after deletion")

	// Verify expectations were met
	mockClient.AssertExpectations(t)
}

// TestAccJumpCloudUser_basic tests the user resource through an acceptance test
// This is an acceptance test that would make real API calls.
// It should be executed with TF_ACC=1 environment variable.
func TestAccJumpCloudUser_basic(t *testing.T) {
	// Skip if not running acceptance tests
	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(),
		CheckDestroy:      testAccCheckJumpCloudUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJumpCloudUserConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserExists("jumpcloud_user.test"),
					resource.TestCheckResourceAttr("jumpcloud_user.test", "username", "test.user"),
					resource.TestCheckResourceAttr("jumpcloud_user.test", "email", "test.user@example.com"),
					resource.TestCheckResourceAttr("jumpcloud_user.test", "firstname", "Test"),
					resource.TestCheckResourceAttr("jumpcloud_user.test", "lastname", "User"),
					resource.TestCheckResourceAttr("jumpcloud_user.test", "description", "Created by Terraform acceptance test"),
				),
			},
			{
				Config: testAccJumpCloudUserConfig_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJumpCloudUserExists("jumpcloud_user.test"),
					resource.TestCheckResourceAttr("jumpcloud_user.test", "email", "updated.user@example.com"),
					resource.TestCheckResourceAttr("jumpcloud_user.test", "firstname", "Updated"),
					resource.TestCheckResourceAttr("jumpcloud_user.test", "description", "Updated by Terraform acceptance test"),
				),
			},
		},
	})
}

// This function checks if the user was properly destroyed
func testAccCheckJumpCloudUserDestroy(s *terraform.State) error {
	// We would implement this to make an API call to verify the resource is gone
	// For now, we'll just return nil since this is for illustration
	return nil
}

// This function checks if the user exists
func testAccCheckJumpCloudUserExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// We would implement this to make an API call to verify the resource exists
		// For now, we'll just return nil since this is for illustration
		return nil
	}
}

// Test configurations for the acceptance tests
const testAccJumpCloudUserConfig_basic = `
resource "jumpcloud_user" "test" {
  username    = "test.user"
  email       = "test.user@example.com"
  firstname   = "Test"
  lastname    = "User"
  password    = "securePassword123!"
  description = "Created by Terraform acceptance test"
}
`

const testAccJumpCloudUserConfig_update = `
resource "jumpcloud_user" "test" {
  username    = "test.user"
  email       = "updated.user@example.com"
  firstname   = "Updated"
  lastname    = "User"
  password    = "securePassword123!"
  description = "Updated by Terraform acceptance test"
}
`
