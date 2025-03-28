package organization

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ResourceOrganization returns the resource schema for JumpCloud organizations
func ResourceOrganization() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationCreate,
		ReadContext:   resourceOrganizationRead,
		UpdateContext: resourceOrganizationUpdate,
		DeleteContext: resourceOrganizationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"logo_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"website": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"contact_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"contact_email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"contact_phone": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"settings": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"parent_org_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"allowed_domains": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceOrganizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Build the organization from the schema data
	org := buildOrganizationStruct(d)

	// Create the organization
	newOrg, err := createOrganization(meta, org)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the ID and other computed values
	d.SetId(newOrg.ID)
	if err := setOrganizationResourceData(d, newOrg); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOrganizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	// Get the organization
	org, err := getOrganization(meta, id)
	if err != nil {
		// If the organization was not found, return nil to remove from state
		if common.IsNotFound(404) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Set the data in the schema
	if err := setOrganizationResourceData(d, org); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOrganizationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	// Build the organization from the schema data
	org := buildOrganizationStruct(d)

	// Update the organization
	updatedOrg, err := updateOrganization(meta, id, org)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the updated data in the schema
	if err := setOrganizationResourceData(d, updatedOrg); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOrganizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := d.Id()

	// Delete the organization
	if err := deleteOrganization(meta, id); err != nil {
		return diag.FromErr(err)
	}

	// Clear the ID to complete removal from state
	d.SetId("")
	return nil
}

// Helper functions
func buildOrganizationStruct(d *schema.ResourceData) *Organization {
	org := &Organization{
		Name:         d.Get("name").(string),
		DisplayName:  d.Get("display_name").(string),
		LogoURL:      d.Get("logo_url").(string),
		Website:      d.Get("website").(string),
		ContactName:  d.Get("contact_name").(string),
		ContactEmail: d.Get("contact_email").(string),
		ContactPhone: d.Get("contact_phone").(string),
		ParentOrgID:  d.Get("parent_org_id").(string),
	}

	// Convert settings map
	if v, ok := d.GetOk("settings"); ok {
		settingsMap := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			settingsMap[k] = v.(string)
		}
		org.Settings = settingsMap
	}

	// Convert allowed domains list
	if v, ok := d.GetOk("allowed_domains"); ok {
		domains := make([]string, 0)
		for _, domain := range v.([]interface{}) {
			domains = append(domains, domain.(string))
		}
		org.AllowedDomains = domains
	}

	return org
}

func setOrganizationResourceData(d *schema.ResourceData, org *Organization) error {
	if err := d.Set("name", org.Name); err != nil {
		return err
	}
	if err := d.Set("display_name", org.DisplayName); err != nil {
		return err
	}
	if err := d.Set("logo_url", org.LogoURL); err != nil {
		return err
	}
	if err := d.Set("website", org.Website); err != nil {
		return err
	}
	if err := d.Set("contact_name", org.ContactName); err != nil {
		return err
	}
	if err := d.Set("contact_email", org.ContactEmail); err != nil {
		return err
	}
	if err := d.Set("contact_phone", org.ContactPhone); err != nil {
		return err
	}
	if err := d.Set("settings", org.Settings); err != nil {
		return err
	}
	if err := d.Set("parent_org_id", org.ParentOrgID); err != nil {
		return err
	}
	if err := d.Set("allowed_domains", org.AllowedDomains); err != nil {
		return err
	}
	if err := d.Set("created", org.Created); err != nil {
		return err
	}
	if err := d.Set("updated", org.Updated); err != nil {
		return err
	}

	return nil
}

// API functions for organizations
func createOrganization(client interface{}, org *Organization) (*Organization, error) {
	// Implementation depends on the actual JumpCloud API
	// This is a placeholder for the actual implementation
	apiClient, err := common.ConvertToClientInterface(client)
	if err != nil {
		return nil, fmt.Errorf("error converting client: %w", err)
	}

	// Convert our internal struct to JSON
	body := map[string]interface{}{
		"name":           org.Name,
		"displayName":    org.DisplayName,
		"logoUrl":        org.LogoURL,
		"website":        org.Website,
		"contactName":    org.ContactName,
		"contactEmail":   org.ContactEmail,
		"contactPhone":   org.ContactPhone,
		"settings":       org.Settings,
		"parentOrgId":    org.ParentOrgID,
		"allowedDomains": org.AllowedDomains,
	}

	// Call the API
	resp, err := apiClient.DoRequest("POST", "/organizations", body)
	if err != nil {
		return nil, fmt.Errorf("error creating organization: %w", err)
	}

	// Parse the response
	var result Organization
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing organization response: %w", err)
	}

	return &result, nil
}

func getOrganization(client interface{}, id string) (*Organization, error) {
	// Implementation depends on the actual JumpCloud API
	// This is a placeholder for the actual implementation
	apiClient, err := common.ConvertToClientInterface(client)
	if err != nil {
		return nil, fmt.Errorf("error converting client: %w", err)
	}

	// Call the API
	resp, err := apiClient.DoRequest("GET", fmt.Sprintf("/organizations/%s", id), nil)
	if err != nil {
		// Check if it's a 404 error
		if len(resp) == 0 {
			return nil, fmt.Errorf("error getting organization: %w", err)
		}
		return nil, fmt.Errorf("error getting organization: %w", err)
	}

	// Parse the response
	var result Organization
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing organization response: %w", err)
	}

	return &result, nil
}

func updateOrganization(client interface{}, id string, org *Organization) (*Organization, error) {
	// Implementation depends on the actual JumpCloud API
	// This is a placeholder for the actual implementation
	apiClient, err := common.ConvertToClientInterface(client)
	if err != nil {
		return nil, fmt.Errorf("error converting client: %w", err)
	}

	// Convert our internal struct to JSON
	body := map[string]interface{}{
		"name":           org.Name,
		"displayName":    org.DisplayName,
		"logoUrl":        org.LogoURL,
		"website":        org.Website,
		"contactName":    org.ContactName,
		"contactEmail":   org.ContactEmail,
		"contactPhone":   org.ContactPhone,
		"settings":       org.Settings,
		"parentOrgId":    org.ParentOrgID,
		"allowedDomains": org.AllowedDomains,
	}

	// Call the API
	resp, err := apiClient.DoRequest("PUT", fmt.Sprintf("/organizations/%s", id), body)
	if err != nil {
		return nil, fmt.Errorf("error updating organization: %w", err)
	}

	// Parse the response
	var result Organization
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing organization response: %w", err)
	}

	return &result, nil
}

func deleteOrganization(client interface{}, id string) error {
	// Implementation depends on the actual JumpCloud API
	// This is a placeholder for the actual implementation
	apiClient, err := common.ConvertToClientInterface(client)
	if err != nil {
		return fmt.Errorf("error converting client: %w", err)
	}

	// Call the API
	_, err = apiClient.DoRequest("DELETE", fmt.Sprintf("/organizations/%s", id), nil)
	if err != nil {
		return fmt.Errorf("error deleting organization: %w", err)
	}

	return nil
}
