package testing

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ProviderResources is a map of resource name to resource
type ProviderResources map[string]*schema.Resource

// ProviderDataSources is a map of data source name to data source
type ProviderDataSources map[string]*schema.Resource

// ProviderFactories is a map of provider name to provider factory function
type ProviderFactories map[string]func() (*schema.Provider, error)

// TestAccProviders holds the providers for acceptance testing
var TestAccProviders map[string]*schema.Provider

// Initialize test providers
func init() {
	// This will be populated by the provider in actual use
	// Here we're just initializing the map
	TestAccProviders = make(map[string]*schema.Provider)
}

// TestAccPreCheck validates the necessary environment variables exist for acceptance tests
func TestAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("JUMPCLOUD_API_KEY"); v == "" {
		t.Fatal("JUMPCLOUD_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("JUMPCLOUD_ORG_ID"); v == "" {
		t.Fatal("JUMPCLOUD_ORG_ID must be set for acceptance tests")
	}
}

// AccPreCheck validates the necessary environment variables exist for acceptance tests
func AccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("JUMPCLOUD_API_KEY"); v == "" {
		t.Fatal("JUMPCLOUD_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("JUMPCLOUD_ORG_ID"); v == "" {
		t.Fatal("JUMPCLOUD_ORG_ID must be set for acceptance tests")
	}
}

// CreateTestStep creates a standard test step configuration for acceptance tests
func CreateTestStep(name, configText string, checkFunc resource.TestCheckFunc) resource.TestStep {
	return resource.TestStep{
		Config: configText,
		Check:  checkFunc,
	}
}

// RegisterTestResources is a helper to register multiple resources at once
func RegisterTestResources(provider *schema.Provider, resources ProviderResources) {
	for name, resource := range resources {
		provider.ResourcesMap[name] = resource
	}
}

// RegisterTestDataSources is a helper to register multiple data sources at once
func RegisterTestDataSources(provider *schema.Provider, dataSources ProviderDataSources) {
	for name, dataSource := range dataSources {
		provider.DataSourcesMap[name] = dataSource
	}
}

// SetupTestCase sets up a standard test case with the provided provider factories
func SetupTestCase(t *testing.T, factories ProviderFactories) {
	t.Helper()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { AccPreCheck(t) },
		ProviderFactories: factories,
	})
}

// NewTestProvider creates a new provider with the given resources and data sources
func NewTestProvider(resources ProviderResources, dataSources ProviderDataSources) *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_API_KEY", nil),
				Description: "JumpCloud API key",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("JUMPCLOUD_ORG_ID", nil),
				Description: "JumpCloud organization ID",
			},
		},
		ResourcesMap:   make(map[string]*schema.Resource),
		DataSourcesMap: make(map[string]*schema.Resource),
	}

	RegisterTestResources(provider, resources)
	RegisterTestDataSources(provider, dataSources)

	return provider
}

// NewProviderFactories creates a map of provider factories with the given provider
func NewProviderFactories(p *schema.Provider) ProviderFactories {
	return ProviderFactories{
		"jumpcloud": func() (*schema.Provider, error) {
			return p, nil
		},
	}
}

// CreateModuleTestProviderFactories creates provider factories for a specific module
// with the given resources and data sources
func CreateModuleTestProviderFactories(resources ProviderResources, dataSources ProviderDataSources) ProviderFactories {
	provider := NewTestProvider(resources, dataSources)
	return NewProviderFactories(provider)
}

// GetProviderFactories retorna as provider factories padrão para uso em todos os testes
func GetProviderFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"jumpcloud": func() (*schema.Provider, error) {
			return TestAccProviders["jumpcloud"], nil
		},
	}
}

// GenerateTestResourceConfig generates a basic configuration for a resource
func GenerateTestResourceConfig(resourceType, resourceName string, attributes map[string]string) string {
	config := "resource \"" + resourceType + "\" \"" + resourceName + "\" {\n"

	for key, value := range attributes {
		config += "  " + key + " = \"" + value + "\"\n"
	}

	config += "}\n"
	return config
}

// GenerateTestDataSourceConfig generates a basic configuration for a data source
func GenerateTestDataSourceConfig(dataSourceType, dataSourceName string, attributes map[string]string) string {
	config := "data \"" + dataSourceType + "\" \"" + dataSourceName + "\" {\n"

	for key, value := range attributes {
		config += "  " + key + " = \"" + value + "\"\n"
	}

	config += "}\n"
	return config
}
