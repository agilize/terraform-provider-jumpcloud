package iplist

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestDataSourceLocationsSchema testa o schema do data source de localizações de IP
func TestDataSourceLocationsSchema(t *testing.T) {
	s := DataSourceLocations()

	// Verificar campo ip_addresses
	if s.Schema["ip_addresses"] == nil {
		t.Error("Expected ip_addresses in schema, but it does not exist")
	}
	if s.Schema["ip_addresses"].Type != schema.TypeList {
		t.Error("Expected ip_addresses to be of type list")
	}
	if !s.Schema["ip_addresses"].Required {
		t.Error("Expected ip_addresses to be required")
	}

	// Verificar campo org_id
	if s.Schema["org_id"] == nil {
		t.Error("Expected org_id in schema, but it does not exist")
	}
	if s.Schema["org_id"].Type != schema.TypeString {
		t.Error("Expected org_id to be of type string")
	}
	if s.Schema["org_id"].Required {
		t.Error("Expected org_id to be optional")
	}

	// Verificar campo locations
	if s.Schema["locations"] == nil {
		t.Error("Expected locations in schema, but it does not exist")
	}
	if s.Schema["locations"].Type != schema.TypeList {
		t.Error("Expected locations to be of type list")
	}
	if !s.Schema["locations"].Computed {
		t.Error("Expected locations to be computed")
	}

	// Verificar estrutura interna de locations
	locationsElem, ok := s.Schema["locations"].Elem.(*schema.Resource)
	if !ok {
		t.Error("Expected locations.Elem to be a *schema.Resource")
		return
	}

	// Lista de campos do tipo string
	stringFields := []string{
		"ip", "country_code", "country_name", "region_code", "region_name",
		"city", "time_zone", "continent_code", "postal_code", "isp",
		"domain", "as", "as_name", "threat_level", "threat_types", "threat_classifiers",
	}
	for _, field := range stringFields {
		if locationsElem.Schema[field] == nil {
			t.Errorf("Expected %s in locations schema, but it does not exist", field)
		}
		if locationsElem.Schema[field].Type != schema.TypeString {
			t.Errorf("Expected %s to be of type string", field)
		}
		if !locationsElem.Schema[field].Computed {
			t.Errorf("Expected %s to be computed", field)
		}
	}

	// Lista de campos do tipo float
	floatFields := []string{"latitude", "longitude"}
	for _, field := range floatFields {
		if locationsElem.Schema[field] == nil {
			t.Errorf("Expected %s in locations schema, but it does not exist", field)
		}
		if locationsElem.Schema[field].Type != schema.TypeFloat {
			t.Errorf("Expected %s to be of type float", field)
		}
		if !locationsElem.Schema[field].Computed {
			t.Errorf("Expected %s to be computed", field)
		}
	}

	// Lista de campos do tipo int
	intFields := []string{"metro_code", "area_code"}
	for _, field := range intFields {
		if locationsElem.Schema[field] == nil {
			t.Errorf("Expected %s in locations schema, but it does not exist", field)
		}
		if locationsElem.Schema[field].Type != schema.TypeInt {
			t.Errorf("Expected %s to be of type int", field)
		}
		if !locationsElem.Schema[field].Computed {
			t.Errorf("Expected %s to be computed", field)
		}
	}

	// Lista de campos do tipo bool
	boolFields := []string{"proxy", "mobile"}
	for _, field := range boolFields {
		if locationsElem.Schema[field] == nil {
			t.Errorf("Expected %s in locations schema, but it does not exist", field)
		}
		if locationsElem.Schema[field].Type != schema.TypeBool {
			t.Errorf("Expected %s to be of type bool", field)
		}
		if !locationsElem.Schema[field].Computed {
			t.Errorf("Expected %s to be computed", field)
		}
	}
}

// TestFlattenIPLocations testa a função flattenIPLocations
func TestFlattenIPLocations(t *testing.T) {
	// Criar dados de teste
	locations := []IPLocationInfo{
		{
			IP:                "192.168.1.1",
			CountryCode:       "US",
			CountryName:       "United States",
			RegionCode:        "CA",
			RegionName:        "California",
			City:              "San Francisco",
			Latitude:          37.7749,
			Longitude:         -122.4194,
			MetroCode:         807,
			AreaCode:          415,
			TimeZone:          "America/Los_Angeles",
			ContinentCode:     "NA",
			PostalCode:        "94105",
			ISP:               "Example ISP",
			Domain:            "example.com",
			AS:                "AS12345",
			ASName:            "Example AS",
			Proxy:             false,
			Mobile:            false,
			ThreatLevel:       "low",
			ThreatTypes:       "",
			ThreatClassifiers: "",
		},
	}

	flattened := flattenIPLocations(locations)
	if len(flattened) != 1 {
		t.Errorf("Expected 1 flattened location, got %d", len(flattened))
	}

	flattenedLoc := flattened[0].(map[string]interface{})

	// Verificar campos
	if flattenedLoc["ip"] != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", flattenedLoc["ip"])
	}
	if flattenedLoc["country_code"] != "US" {
		t.Errorf("Expected country_code US, got %s", flattenedLoc["country_code"])
	}
	if flattenedLoc["city"] != "San Francisco" {
		t.Errorf("Expected city San Francisco, got %s", flattenedLoc["city"])
	}
	if flattenedLoc["latitude"] != 37.7749 {
		t.Errorf("Expected latitude 37.7749, got %f", flattenedLoc["latitude"])
	}
	if flattenedLoc["longitude"] != -122.4194 {
		t.Errorf("Expected longitude -122.4194, got %f", flattenedLoc["longitude"])
	}
	if flattenedLoc["threat_level"] != "low" {
		t.Errorf("Expected threat_level low, got %s", flattenedLoc["threat_level"])
	}

	// Teste com lista vazia
	var emptyList []IPLocationInfo
	flattenedEmpty := flattenIPLocations(emptyList)
	if len(flattenedEmpty) != 0 {
		t.Errorf("Expected 0 entries for empty list, got %d", len(flattenedEmpty))
	}
}
