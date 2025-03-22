package scim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ScimSchemaAttribute represents a SCIM schema attribute
type ScimSchemaAttribute struct {
	Name          string                `json:"name"`
	Type          string                `json:"type"`
	MultiValued   bool                  `json:"multiValued"`
	Required      bool                  `json:"required"`
	CaseExact     bool                  `json:"caseExact"`
	Mutable       bool                  `json:"mutable"`
	Returned      string                `json:"returned"`
	Uniqueness    string                `json:"uniqueness"`
	Description   string                `json:"description,omitempty"`
	SubAttributes []ScimSchemaAttribute `json:"subAttributes,omitempty"`
}

// ScimSchema represents a SCIM schema in JumpCloud
type ScimSchema struct {
	ID          string                `json:"_id"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	URI         string                `json:"uri"`
	Type        string                `json:"type"` // core, extension, custom
	Attributes  []ScimSchemaAttribute `json:"attributes"`
	Standard    bool                  `json:"standard"` // indicates if it's a standard schema
	OrgID       string                `json:"orgId,omitempty"`
	Created     string                `json:"created"`
	Updated     string                `json:"updated"`
}

// DataSourceSchema returns a schema resource for the SCIM schema data source
func DataSourceSchema() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSchemaRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name", "uri"},
				Description:   "ID of the SCIM schema",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id", "uri"},
				Description:   "Name of the SCIM schema",
			},
			"uri": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id", "name"},
				Description:   "URI of the SCIM schema",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the SCIM schema",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the schema (core, extension, custom)",
			},
			"standard": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if it's a standard schema",
			},
			"attributes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of schema attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the attribute",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the attribute",
						},
						"multi_valued": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates if the attribute accepts multiple values",
						},
						"required": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates if the attribute is required",
						},
						"case_exact": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates if the attribute is case-sensitive",
						},
						"mutable": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates if the attribute can be modified",
						},
						"returned": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Indicates when the attribute is returned (always, never, default, request)",
						},
						"uniqueness": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Indicates the uniqueness of the attribute (none, server, global)",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the attribute",
						},
						"sub_attributes": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of sub-attributes (for complex attributes)",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Name of the sub-attribute",
									},
									"type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Type of the sub-attribute",
									},
									"multi_valued": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Indicates if the sub-attribute accepts multiple values",
									},
									"required": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Indicates if the sub-attribute is required",
									},
									"case_exact": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Indicates if the sub-attribute is case-sensitive",
									},
									"mutable": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Indicates if the sub-attribute can be modified",
									},
									"returned": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Indicates when the sub-attribute is returned",
									},
									"uniqueness": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Indicates the uniqueness of the sub-attribute",
									},
									"description": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Description of the sub-attribute",
									},
								},
							},
						},
					},
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the schema",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date of the schema",
			},
		},
	}
}

func dataSourceSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get orgId parameter if available
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Check which identifier was provided and fetch the schema accordingly
	var schemaID, schemaName, schemaURI string
	var url string

	if v, ok := d.GetOk("id"); ok {
		schemaID = v.(string)
	}

	if v, ok := d.GetOk("name"); ok {
		schemaName = v.(string)
	}

	if v, ok := d.GetOk("uri"); ok {
		schemaURI = v.(string)
	}

	// Ensure at least one identifier is provided
	if schemaID == "" && schemaName == "" && schemaURI == "" {
		return diag.FromErr(fmt.Errorf("one of id, name, or uri must be specified"))
	}

	// If ID is provided, fetch directly by ID
	if schemaID != "" {
		url = fmt.Sprintf("/api/v2/scim/schemas/%s%s", schemaID, orgIDParam)
		tflog.Debug(ctx, fmt.Sprintf("Fetching SCIM schema by ID: %s", schemaID))
		resp, err := c.DoRequest(http.MethodGet, url, nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error fetching SCIM schema by ID: %v", err))
		}

		var schema ScimSchema
		if err := json.Unmarshal(resp, &schema); err != nil {
			return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
		}

		d.SetId(schema.ID)
		diags = setSchemaValues(d, schema)
		return diags
	}

	// Otherwise, we need to list all schemas and find the matching one
	listURL := fmt.Sprintf("/api/v2/scim/schemas%s", orgIDParam)
	resp, err := c.DoRequest(http.MethodGet, listURL, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing SCIM schemas: %v", err))
	}

	var schemasResp struct {
		Results []ScimSchema `json:"results"`
	}
	if err := json.Unmarshal(resp, &schemasResp); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Find schema by name or URI
	var foundSchema *ScimSchema
	for _, schema := range schemasResp.Results {
		if schemaName != "" && schema.Name == schemaName {
			foundSchema = &schema
			break
		}
		if schemaURI != "" && schema.URI == schemaURI {
			foundSchema = &schema
			break
		}
	}

	if foundSchema == nil {
		if schemaName != "" {
			return diag.FromErr(fmt.Errorf("SCIM schema with name '%s' not found", schemaName))
		}
		return diag.FromErr(fmt.Errorf("SCIM schema with URI '%s' not found", schemaURI))
	}

	// Set ID and other values
	d.SetId(foundSchema.ID)
	diags = setSchemaValues(d, *foundSchema)
	return diags
}

func setSchemaValues(d *schema.ResourceData, schema ScimSchema) diag.Diagnostics {
	var diags diag.Diagnostics

	d.Set("name", schema.Name)
	d.Set("description", schema.Description)
	d.Set("uri", schema.URI)
	d.Set("type", schema.Type)
	d.Set("standard", schema.Standard)
	d.Set("created", schema.Created)
	d.Set("updated", schema.Updated)

	// Set org_id if present
	if schema.OrgID != "" {
		d.Set("org_id", schema.OrgID)
	}

	// Set attributes
	attributes := flattenSchemaAttributes(schema.Attributes)
	if err := d.Set("attributes", attributes); err != nil {
		return diag.FromErr(fmt.Errorf("error setting attributes: %v", err))
	}

	return diags
}

func flattenSchemaAttributes(attributes []ScimSchemaAttribute) []map[string]interface{} {
	if len(attributes) == 0 {
		return make([]map[string]interface{}, 0)
	}

	items := make([]map[string]interface{}, len(attributes))
	for i, attr := range attributes {
		item := map[string]interface{}{
			"name":         attr.Name,
			"type":         attr.Type,
			"multi_valued": attr.MultiValued,
			"required":     attr.Required,
			"case_exact":   attr.CaseExact,
			"mutable":      attr.Mutable,
			"returned":     attr.Returned,
			"uniqueness":   attr.Uniqueness,
		}

		// Add description if present
		if attr.Description != "" {
			item["description"] = attr.Description
		}

		// Add sub-attributes if present
		if len(attr.SubAttributes) > 0 {
			subAttrs := flattenSchemaSubAttributes(attr.SubAttributes)
			item["sub_attributes"] = subAttrs
		}

		items[i] = item
	}

	return items
}

func flattenSchemaSubAttributes(subAttributes []ScimSchemaAttribute) []map[string]interface{} {
	if len(subAttributes) == 0 {
		return make([]map[string]interface{}, 0)
	}

	items := make([]map[string]interface{}, len(subAttributes))
	for i, attr := range subAttributes {
		item := map[string]interface{}{
			"name":         attr.Name,
			"type":         attr.Type,
			"multi_valued": attr.MultiValued,
			"required":     attr.Required,
			"case_exact":   attr.CaseExact,
			"mutable":      attr.Mutable,
			"returned":     attr.Returned,
			"uniqueness":   attr.Uniqueness,
		}

		// Add description if present
		if attr.Description != "" {
			item["description"] = attr.Description
		}

		items[i] = item
	}

	return items
}
