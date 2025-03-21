package iplist

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestDataSourceListsSchema testa o schema do data source de listas de IPs
func TestDataSourceListsSchema(t *testing.T) {
	s := DataSourceLists()

	// Verificar campo filter
	if s.Schema["filter"] == nil {
		t.Error("Expected filter in schema, but it does not exist")
	}
	if s.Schema["filter"].Type != schema.TypeList {
		t.Error("Expected filter to be of type list")
	}
	if s.Schema["filter"].Required {
		t.Error("Expected filter to be optional")
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

	// Verificar campo ip_lists
	if s.Schema["ip_lists"] == nil {
		t.Error("Expected ip_lists in schema, but it does not exist")
	}
	if s.Schema["ip_lists"].Type != schema.TypeList {
		t.Error("Expected ip_lists to be of type list")
	}
	if !s.Schema["ip_lists"].Computed {
		t.Error("Expected ip_lists to be computed")
	}

	// Verificar estrutura interna de ip_lists
	ipListsElem, ok := s.Schema["ip_lists"].Elem.(*schema.Resource)
	if !ok {
		t.Error("Expected ip_lists.Elem to be a *schema.Resource")
		return
	}

	for _, field := range []string{"id", "name", "description", "type", "created", "updated"} {
		if ipListsElem.Schema[field] == nil {
			t.Errorf("Expected %s in ip_lists schema, but it does not exist", field)
		}
		if ipListsElem.Schema[field].Type != schema.TypeString {
			t.Errorf("Expected %s to be of type string", field)
		}
		if !ipListsElem.Schema[field].Computed {
			t.Errorf("Expected %s to be computed", field)
		}
	}

	// Verificar campo ips dentro de ip_lists
	if ipListsElem.Schema["ips"] == nil {
		t.Error("Expected ips in ip_lists schema, but it does not exist")
	}
	if ipListsElem.Schema["ips"].Type != schema.TypeList {
		t.Error("Expected ips to be of type list")
	}
	if !ipListsElem.Schema["ips"].Computed {
		t.Error("Expected ips to be computed")
	}
}

// TestFilterIPLists testa a função de filtragem de listas de IPs
func TestFilterIPLists(t *testing.T) {
	ipLists := []IPList{
		{
			ID:   "list1",
			Name: "AllowList",
			Type: "allow",
		},
		{
			ID:   "list2",
			Name: "DenyList",
			Type: "deny",
		},
		{
			ID:   "list3",
			Name: "AllowList2",
			Type: "allow",
		},
	}

	// Teste de filtro por nome
	nameFilter := map[string]interface{}{
		"name": "AllowList",
	}
	filteredByName := filterIPLists(ipLists, nameFilter)
	if len(filteredByName) != 1 {
		t.Errorf("Expected 1 IP list filtered by name, got %d", len(filteredByName))
	}
	if filteredByName[0].ID != "list1" {
		t.Errorf("Expected filtered IP list ID to be list1, got %s", filteredByName[0].ID)
	}

	// Teste de filtro por tipo
	typeFilter := map[string]interface{}{
		"type": "allow",
	}
	filteredByType := filterIPLists(ipLists, typeFilter)
	if len(filteredByType) != 2 {
		t.Errorf("Expected 2 IP lists filtered by type, got %d", len(filteredByType))
	}

	// Teste sem filtro
	emptyFilter := map[string]interface{}{}
	noFilter := filterIPLists(ipLists, emptyFilter)
	if len(noFilter) != 3 {
		t.Errorf("Expected 3 IP lists with no filter, got %d", len(noFilter))
	}
}

// TestFlattenFunctions testa as funções de achatamento
func TestFlattenFunctions(t *testing.T) {
	// Test flattenIPLists
	ipLists := []IPList{
		{
			ID:          "list1",
			Name:        "TestList",
			Description: "Test description",
			Type:        "allow",
			Created:     "2023-01-01T00:00:00Z",
			Updated:     "2023-01-02T00:00:00Z",
			IPs: []IPAddressEntry{
				{
					Address:     "192.168.1.1",
					Description: "Test IP",
				},
			},
		},
	}

	flattened := flattenIPLists(ipLists)
	if len(flattened) != 1 {
		t.Errorf("Expected 1 flattened IP list, got %d", len(flattened))
	}

	flattenedList := flattened[0].(map[string]interface{})
	if flattenedList["id"] != "list1" {
		t.Errorf("Expected id list1, got %s", flattenedList["id"])
	}
	if flattenedList["name"] != "TestList" {
		t.Errorf("Expected name TestList, got %s", flattenedList["name"])
	}
	if flattenedList["description"] != "Test description" {
		t.Errorf("Expected description 'Test description', got %s", flattenedList["description"])
	}
	if flattenedList["type"] != "allow" {
		t.Errorf("Expected type allow, got %s", flattenedList["type"])
	}

	// Test empty lists
	var emptyList []IPList
	flattenedEmpty := flattenIPLists(emptyList)
	if len(flattenedEmpty) != 0 {
		t.Errorf("Expected 0 entries for empty list, got %d", len(flattenedEmpty))
	}
}
