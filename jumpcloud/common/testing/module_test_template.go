package testing

import (
	"testing"
)

// ModuleTestHelper provides standard test helpers for module acceptance tests
type ModuleTestHelper struct {
	// Module-specific resources
	Resources ProviderResources

	// Module-specific data sources
	DataSources ProviderDataSources

	// Provider factories for test cases
	ProviderFactories ProviderFactories

	// Any additional setup needed for this module's tests
	CustomSetup func(t *testing.T)
}

// NewModuleTestHelper creates a new module test helper
func NewModuleTestHelper(resources ProviderResources, dataSources ProviderDataSources) *ModuleTestHelper {
	factories := CreateModuleTestProviderFactories(resources, dataSources)

	return &ModuleTestHelper{
		Resources:         resources,
		DataSources:       dataSources,
		ProviderFactories: factories,
		CustomSetup:       func(t *testing.T) {},
	}
}

// WithCustomSetup adds custom setup logic for the module's tests
func (m *ModuleTestHelper) WithCustomSetup(setup func(t *testing.T)) *ModuleTestHelper {
	m.CustomSetup = setup
	return m
}

// PreCheck provides a standard pre-check function for acceptance tests
func (m *ModuleTestHelper) PreCheck(t *testing.T) {
	// Run standard pre-check
	AccPreCheck(t)

	// Run any module-specific setup
	m.CustomSetup(t)
}

// RunTestCase runs a test case with the module's provider factories
func (m *ModuleTestHelper) RunTestCase(t *testing.T, testCase func(t *testing.T)) {
	t.Helper()
	SetupTestCase(t, m.ProviderFactories)
	testCase(t)
}

// ExampleModuleHelper shows how to use this template in a specific module
func ExampleModuleHelper() {
	// In each module's test_utils.go file:

	/*
		package mymodule

		import (
			"testing"

			commontest "registry.terraform.io/agilize/jumpcloud/jumpcloud/common/testing"
		)

		// Define module-specific resources
		var moduleResources = commontest.ProviderResources{
			"jumpcloud_my_resource": ResourceMyResource(),
		}

		// Define module-specific data sources
		var moduleDataSources = commontest.ProviderDataSources{
			"jumpcloud_my_data_source": DataSourceMyDataSource(),
		}

		// Create a module test helper
		var TestHelper = commontest.NewModuleTestHelper(
			moduleResources,
			moduleDataSources,
		).WithCustomSetup(func(t *testing.T) {
			// Add any module-specific setup logic here
		})

		// Helper for your module's acceptance tests
		func testAccPreCheck(t *testing.T) {
			TestHelper.PreCheck(t)
		}
	*/
}
