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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// AttributeMapping represents a mapping between a source attribute and a target attribute
type AttributeMapping struct {
	SourcePath  string `json:"sourcePath"`
	TargetPath  string `json:"targetPath"`
	Constant    string `json:"constant,omitempty"`
	Expression  string `json:"expression,omitempty"`
	Transform   string `json:"transform,omitempty"`
	Required    bool   `json:"required"`
	Multivalued bool   `json:"multivalued"`
}

// ScimAttributeMapping represents a SCIM attribute mapping in JumpCloud
type ScimAttributeMapping struct {
	ID           string             `json:"_id,omitempty"`
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	ServerID     string             `json:"serverId"`
	SchemaID     string             `json:"schemaId"`
	Mappings     []AttributeMapping `json:"mappings"`
	Direction    string             `json:"direction"` // inbound, outbound, bidirectional
	ObjectClass  string             `json:"objectClass,omitempty"`
	AutoGenerate bool               `json:"autoGenerate,omitempty"`
	OrgID        string             `json:"orgId,omitempty"`
	Created      string             `json:"created,omitempty"`
	Updated      string             `json:"updated,omitempty"`
}

// ResourceAttributeMapping returns a schema resource for managing SCIM attribute mappings in JumpCloud
func ResourceAttributeMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAttributeMappingCreate,
		ReadContext:   resourceAttributeMappingRead,
		UpdateContext: resourceAttributeMappingUpdate,
		DeleteContext: resourceAttributeMappingDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Name of the SCIM attribute mapping",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the SCIM attribute mapping",
			},
			"server_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the associated SCIM server",
			},
			"schema_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the associated SCIM schema",
			},
			"direction": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"inbound", "outbound", "bidirectional"}, false),
				Description:  "Direction of the mapping (inbound, outbound, bidirectional)",
			},
			"object_class": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Object class for the mapping (e.g., User, Group)",
			},
			"auto_generate": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicates if the mapping should be automatically generated",
			},
			"mappings": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Path of the source attribute",
						},
						"target_path": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Path of the target attribute",
						},
						"constant": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Constant value to be used (if not mapped from a source value)",
						},
						"expression": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Custom transformation expression",
						},
						"transform": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Transformation to be applied to the value (e.g., toLowerCase, toUpperCase)",
						},
						"required": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Indicates if the attribute is required",
						},
						"multivalued": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Indicates if the attribute accepts multiple values",
						},
					},
				},
				Description: "List of attribute mappings",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the mapping",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date of the mapping",
			},
		},
	}
}

func resourceAttributeMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build ScimAttributeMapping object from terraform data
	mapping := &ScimAttributeMapping{
		Name:         d.Get("name").(string),
		ServerID:     d.Get("server_id").(string),
		SchemaID:     d.Get("schema_id").(string),
		Direction:    d.Get("direction").(string),
		AutoGenerate: d.Get("auto_generate").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		mapping.Description = v.(string)
	}

	if v, ok := d.GetOk("object_class"); ok {
		mapping.ObjectClass = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		mapping.OrgID = v.(string)
	}

	// Process attribute mappings
	if v, ok := d.GetOk("mappings"); ok {
		mappingsList := v.([]interface{})
		attributeMappings := make([]AttributeMapping, len(mappingsList))

		for i, meta := range mappingsList {
			mapData := meta.(map[string]interface{})

			// Validate that constant and expression are not used together
			constant, hasConstant := mapData["constant"].(string)
			expression, hasExpression := mapData["expression"].(string)

			if hasConstant && constant != "" && hasExpression && expression != "" {
				return diag.FromErr(fmt.Errorf("error in mapping %d: 'constant' and 'expression' cannot be used together", i))
			}

			attributeMapping := AttributeMapping{
				SourcePath:  mapData["source_path"].(string),
				TargetPath:  mapData["target_path"].(string),
				Required:    mapData["required"].(bool),
				Multivalued: mapData["multivalued"].(bool),
			}

			// Optional mapping fields
			if v, ok := mapData["constant"]; ok && v.(string) != "" {
				attributeMapping.Constant = v.(string)
			}

			if v, ok := mapData["expression"]; ok && v.(string) != "" {
				attributeMapping.Expression = v.(string)
			}

			if v, ok := mapData["transform"]; ok && v.(string) != "" {
				attributeMapping.Transform = v.(string)
			}

			attributeMappings[i] = attributeMapping
		}

		mapping.Mappings = attributeMappings
	}

	// Serialize to JSON
	reqBody, err := json.Marshal(mapping)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing SCIM attribute mapping: %v", err))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/scim/servers/%s/mappings", mapping.ServerID)
	if mapping.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, mapping.OrgID)
	}

	// Make request to create mapping
	tflog.Debug(ctx, fmt.Sprintf("Creating SCIM attribute mapping for server: %s", mapping.ServerID))
	resp, err := c.DoRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating SCIM attribute mapping: %v", err))
	}

	// Deserialize response
	var createdMapping ScimAttributeMapping
	if err := json.Unmarshal(resp, &createdMapping); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdMapping.ID == "" {
		return diag.FromErr(fmt.Errorf("created SCIM attribute mapping returned without an ID"))
	}

	// Set ID in state
	d.SetId(createdMapping.ID)

	// Read the resource to update state with all computed fields
	return resourceAttributeMappingRead(ctx, d, meta)
}

func resourceAttributeMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get mapping ID
	mappingID := d.Id()
	if mappingID == "" {
		return diag.FromErr(fmt.Errorf("SCIM attribute mapping ID is required"))
	}

	// Get server ID from state
	var serverID string
	if v, ok := d.GetOk("server_id"); ok {
		serverID = v.(string)
	} else {
		// If we don't have the server_id in state (possibly during import),
		// we need to fetch the mapping by ID to discover the server_id
		url := fmt.Sprintf("/api/v2/scim/mappings/%s", mappingID)
		if v, ok := d.GetOk("org_id"); ok {
			url = fmt.Sprintf("%s?orgId=%s", url, v.(string))
		}

		resp, err := c.DoRequest(http.MethodGet, url, nil)
		if err != nil {
			if common.IsNotFoundError(err) {
				tflog.Warn(ctx, fmt.Sprintf("SCIM attribute mapping %s not found, removing from state", mappingID))
				d.SetId("")
				return diags
			}
			return diag.FromErr(fmt.Errorf("error fetching SCIM mapping: %v", err))
		}

		var mapping ScimAttributeMapping
		if err := json.Unmarshal(resp, &mapping); err != nil {
			return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
		}

		serverID = mapping.ServerID
		d.Set("server_id", serverID)
	}

	// Get orgId parameter if available
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/scim/servers/%s/mappings/%s%s", serverID, mappingID, orgIDParam)

	// Make request to get mapping details
	tflog.Debug(ctx, fmt.Sprintf("Reading SCIM attribute mapping: %s", mappingID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("SCIM attribute mapping %s not found, removing from state", mappingID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading SCIM attribute mapping: %v", err))
	}

	// Deserialize response
	var mapping ScimAttributeMapping
	if err := json.Unmarshal(resp, &mapping); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing SCIM attribute mapping: %v", err))
	}

	// Set values in state
	d.Set("name", mapping.Name)
	d.Set("description", mapping.Description)
	d.Set("server_id", mapping.ServerID)
	d.Set("schema_id", mapping.SchemaID)
	d.Set("direction", mapping.Direction)
	d.Set("object_class", mapping.ObjectClass)
	d.Set("auto_generate", mapping.AutoGenerate)
	d.Set("created", mapping.Created)
	d.Set("updated", mapping.Updated)

	// Process attribute mappings
	if mapping.Mappings != nil {
		attributeMappings := make([]map[string]interface{}, len(mapping.Mappings))
		for i, meta := range mapping.Mappings {
			attributeMapping := map[string]interface{}{
				"source_path": meta.SourcePath,
				"target_path": meta.TargetPath,
				"required":    meta.Required,
				"multivalued": meta.Multivalued,
			}

			if meta.Constant != "" {
				attributeMapping["constant"] = meta.Constant
			}

			if meta.Expression != "" {
				attributeMapping["expression"] = meta.Expression
			}

			if meta.Transform != "" {
				attributeMapping["transform"] = meta.Transform
			}

			attributeMappings[i] = attributeMapping
		}

		if err := d.Set("mappings", attributeMappings); err != nil {
			return diag.FromErr(fmt.Errorf("error setting mappings: %v", err))
		}
	}

	// Set OrgID if present
	if mapping.OrgID != "" {
		d.Set("org_id", mapping.OrgID)
	}

	return diags
}

func resourceAttributeMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get mapping ID
	mappingID := d.Id()
	if mappingID == "" {
		return diag.FromErr(fmt.Errorf("SCIM attribute mapping ID is required"))
	}

	// Build ScimAttributeMapping object from terraform data
	mapping := &ScimAttributeMapping{
		ID:           mappingID,
		Name:         d.Get("name").(string),
		ServerID:     d.Get("server_id").(string),
		SchemaID:     d.Get("schema_id").(string),
		Direction:    d.Get("direction").(string),
		AutoGenerate: d.Get("auto_generate").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		mapping.Description = v.(string)
	}

	if v, ok := d.GetOk("object_class"); ok {
		mapping.ObjectClass = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		mapping.OrgID = v.(string)
	}

	// Process attribute mappings
	if v, ok := d.GetOk("mappings"); ok {
		mappingsList := v.([]interface{})
		attributeMappings := make([]AttributeMapping, len(mappingsList))

		for i, meta := range mappingsList {
			mapData := meta.(map[string]interface{})

			// Validate that constant and expression are not used together
			constant, hasConstant := mapData["constant"].(string)
			expression, hasExpression := mapData["expression"].(string)

			if hasConstant && constant != "" && hasExpression && expression != "" {
				return diag.FromErr(fmt.Errorf("error in mapping %d: 'constant' and 'expression' cannot be used together", i))
			}

			attributeMapping := AttributeMapping{
				SourcePath:  mapData["source_path"].(string),
				TargetPath:  mapData["target_path"].(string),
				Required:    mapData["required"].(bool),
				Multivalued: mapData["multivalued"].(bool),
			}

			// Optional mapping fields
			if v, ok := mapData["constant"]; ok && v.(string) != "" {
				attributeMapping.Constant = v.(string)
			}

			if v, ok := mapData["expression"]; ok && v.(string) != "" {
				attributeMapping.Expression = v.(string)
			}

			if v, ok := mapData["transform"]; ok && v.(string) != "" {
				attributeMapping.Transform = v.(string)
			}

			attributeMappings[i] = attributeMapping
		}

		mapping.Mappings = attributeMappings
	}

	// Serialize to JSON
	reqBody, err := json.Marshal(mapping)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing SCIM attribute mapping: %v", err))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/scim/servers/%s/mappings/%s", mapping.ServerID, mappingID)
	if mapping.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, mapping.OrgID)
	}

	// Make request to update mapping
	tflog.Debug(ctx, fmt.Sprintf("Updating SCIM attribute mapping: %s", mappingID))
	_, err = c.DoRequest(http.MethodPut, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating SCIM attribute mapping: %v", err))
	}

	// Read the resource to update state with all computed fields
	return resourceAttributeMappingRead(ctx, d, meta)
}

func resourceAttributeMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get mapping ID
	mappingID := d.Id()
	if mappingID == "" {
		return diag.FromErr(fmt.Errorf("SCIM attribute mapping ID is required"))
	}

	// Get server ID
	serverID := d.Get("server_id").(string)
	if serverID == "" {
		return diag.FromErr(fmt.Errorf("SCIM server ID is required"))
	}

	// Get orgId parameter if available
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/scim/servers/%s/mappings/%s%s", serverID, mappingID, orgIDParam)

	// Make request to delete mapping
	tflog.Debug(ctx, fmt.Sprintf("Deleting SCIM attribute mapping: %s", mappingID))
	_, err := c.DoRequest(http.MethodDelete, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting SCIM attribute mapping: %v", err))
	}

	// Clear ID from state
	d.SetId("")

	return diags
}
