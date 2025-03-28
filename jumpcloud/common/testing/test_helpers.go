package testing

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestCheckResourceAttrSet is a helper that checks if a resource's attribute is set
func TestCheckResourceAttrSet(name, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		if rs.Primary.Attributes[key] == "" {
			return fmt.Errorf("attribute '%s' not set", key)
		}

		return nil
	}
}

// TestCheckResourceAttrEqual is a helper that checks if two resources' attributes are equal
func TestCheckResourceAttrEqual(name1, key1, name2, key2 string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs1, ok := s.RootModule().Resources[name1]
		if !ok {
			return fmt.Errorf("resource not found: %s", name1)
		}
		if rs1.Primary.ID == "" {
			return fmt.Errorf("resource ID not set for %s", name1)
		}

		rs2, ok := s.RootModule().Resources[name2]
		if !ok {
			return fmt.Errorf("resource not found: %s", name2)
		}
		if rs2.Primary.ID == "" {
			return fmt.Errorf("resource ID not set for %s", name2)
		}

		attr1 := rs1.Primary.Attributes[key1]
		attr2 := rs2.Primary.Attributes[key2]

		if attr1 != attr2 {
			return fmt.Errorf("attributes not equal: %s.%s = %s, %s.%s = %s", name1, key1, attr1, name2, key2, attr2)
		}

		return nil
	}
}

// RandomName generates a random name for testing resources
func RandomName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, secureRandomInt(100000))
}

// RandomEmail generates a random email for testing
func RandomEmail(prefix string) string {
	return fmt.Sprintf("%s-%d@example.com", prefix, secureRandomInt(100000))
}

// RandomString generates a random string of the specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		// Generate a random index in the charset
		charIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[charIndex.Int64()]
	}
	return string(result)
}

// secureRandomInt generates a random integer between 0 and max using crypto/rand
func secureRandomInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

// SkipIfEnvNotSet skips the test if the specified environment variable is not set
func SkipIfEnvNotSet(t *testing.T, env string) {
	t.Helper()
	if value := strings.TrimSpace(GetEnv(env, "")); value == "" {
		t.Skipf("Skipping test: environment variable %s not set", env)
	}
}

// GetEnv gets the value of an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value, exists := GetEnvOk(key); exists {
		return value
	}
	return defaultValue
}

// GetEnvOk gets the value of an environment variable and a boolean indicating if it was found
func GetEnvOk(key string) (string, bool) {
	value, exists := findEnv(key)
	return value, exists
}

// Internal helper to find environment variable
func findEnv(key string) (string, bool) {
	// Use go standard library to get env
	// This is just a placeholder - the actual implementation would depend
	// on how the test is run and environment variables are accessed
	value, exists := getOsEnv(key)
	return value, exists
}

// getOsEnv is a wrapper for os.LookupEnv to make testing easier
var getOsEnv = func(key string) (string, bool) {
	return os.LookupEnv(key)
}
