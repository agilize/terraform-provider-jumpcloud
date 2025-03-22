package templates

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

	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
	"registry.terraform.io/agilize/jumpcloud/pkg/errors"
)

// ExampleDataObject represents the data structure returned by the API
type ExampleDataObject struct {
	ID          string   `json:"_id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags,omitempty"`
	OrgID       string   `json:"orgId,omitempty"`
	Status      string   `json:"status"`
	Created     string   `json:"created,omitempty"`
	Updated     string   `json:"updated,omitempty"`
}

// DataSourceExamples returns the schema for JumpCloud example data source (list)
func DataSourceExamples() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExamplesRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			// filter fields come first to emphasize their role in data filtering
			"filter": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Filter criteria for examples",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter by name (supports partial match)",
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"type1", "type2", "type3"}, false),
							Description:  "Filter by type (exact match)",
						},
						"status": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"active", "inactive", "pending"}, false),
							Description:  "Filter by status (exact match)",
						},
					},
				},
			},

			// output fields follow
			"examples": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of example resources",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier of the example",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the example",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the example",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the example",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the example",
						},
						"tags": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Tags associated with the example",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation timestamp of the example",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last update timestamp of the example",
						},
					},
				},
			},
		},
		Description: "Retrieves a list of JumpCloud examples with optional filtering",
	}
}

// DataSourceExample returns the schema for a single JumpCloud example
func DataSourceExample() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExampleRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			// Search criteria fields come first
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "The unique identifier of the example to retrieve",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "The name of the example to retrieve (must be exact match)",
			},

			// Output fields follow
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the example",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the example",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the example",
			},
			"tags": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags associated with the example",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp of the example",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update timestamp of the example",
			},
		},
		Description: "Retrieves a single JumpCloud example by ID or name",
	}
}

func dataSourceExampleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading JumpCloud example data source")

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	// Determine whether to search by ID or name
	var path string
	if idValue, ok := d.GetOk("id"); ok {
		path = fmt.Sprintf("/api/v2/example/resources/%s", idValue.(string))
	} else if name, ok := d.GetOk("name"); ok {
		// Use a query parameter for name search
		path = fmt.Sprintf("/api/v2/example/resources?name=%s", name.(string))
	} else {
		return diag.FromErr(errors.NewInternalError("either id or name must be provided"))
	}

	// Get resource via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to read example: %s", path))
	resp, err := client.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		if apiclient.IsNotFound(err) {
			return diag.FromErr(errors.NewNotFoundError("example not found"))
		}
		return diag.FromErr(errors.NewInternalError("error reading example: %v", err))
	}

	// Handle single resource or resource list based on the path
	var example ExampleDataObject
	var examples []ExampleDataObject

	if _, ok := d.GetOk("id"); ok {
		// Deserialize single response
		if err := json.Unmarshal(resp, &example); err != nil {
			return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
		}

		d.SetId(example.ID)

		// Set values in state
		if err := d.Set("name", example.Name); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting name: %v", err))
		}
		if err := d.Set("description", example.Description); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting description: %v", err))
		}
		if err := d.Set("type", example.Type); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting type: %v", err))
		}
		if err := d.Set("status", example.Status); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting status: %v", err))
		}
		if err := d.Set("tags", example.Tags); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting tags: %v", err))
		}
		if err := d.Set("created", example.Created); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting created: %v", err))
		}
		if err := d.Set("updated", example.Updated); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting updated: %v", err))
		}
	} else {
		// Deserialize list response
		if err := json.Unmarshal(resp, &examples); err != nil {
			return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
		}

		// Filter by name (exact match)
		name := d.Get("name").(string)
		var found bool

		for _, ex := range examples {
			if ex.Name == name {
				example = ex
				found = true
				break
			}
		}

		if !found {
			return diag.FromErr(errors.NewNotFoundError("example with name %s not found", name))
		}

		d.SetId(example.ID)

		// Set values in state
		if err := d.Set("description", example.Description); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting description: %v", err))
		}
		if err := d.Set("type", example.Type); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting type: %v", err))
		}
		if err := d.Set("status", example.Status); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting status: %v", err))
		}
		if err := d.Set("tags", example.Tags); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting tags: %v", err))
		}
		if err := d.Set("created", example.Created); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting created: %v", err))
		}
		if err := d.Set("updated", example.Updated); err != nil {
			return diag.FromErr(errors.NewInternalError("error setting updated: %v", err))
		}
	}

	tflog.Debug(ctx, "Successfully read JumpCloud example data source")
	return diag.Diagnostics{}
}

func dataSourceExamplesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Info(ctx, "Reading JumpCloud examples data source")

	client, ok := meta.(*apiclient.Client)
	if !ok {
		return diag.FromErr(errors.NewInternalError("invalid client configuration"))
	}

	// Construct query parameters based on filters
	query := ""
	if filters, ok := d.GetOk("filter"); ok && len(filters.([]interface{})) > 0 {
		filter := filters.([]interface{})[0].(map[string]interface{})

		if name, ok := filter["name"].(string); ok && name != "" {
			query += fmt.Sprintf("&name=%s", name)
		}

		if typeName, ok := filter["type"].(string); ok && typeName != "" {
			query += fmt.Sprintf("&type=%s", typeName)
		}

		if status, ok := filter["status"].(string); ok && status != "" {
			query += fmt.Sprintf("&status=%s", status)
		}
	}

	// Remove leading & if present
	if len(query) > 0 && query[0] == '&' {
		query = "?" + query[1:]
	} else if len(query) > 0 {
		query = "?" + query
	}

	path := fmt.Sprintf("/api/v2/example/resources%s", query)

	// Get resources via API
	tflog.Debug(ctx, fmt.Sprintf("Calling JumpCloud API to list examples: %s", path))
	resp, err := client.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		return diag.FromErr(errors.NewInternalError("error listing examples: %v", err))
	}

	// Deserialize list response
	var examples []ExampleDataObject
	if err := json.Unmarshal(resp, &examples); err != nil {
		return diag.FromErr(errors.NewInternalError("error deserializing response: %v", err))
	}

	// Generate a consistent ID for the data source
	d.SetId(fmt.Sprintf("examples-%d", time.Now().Unix()))

	// Transform examples to schema format
	examplesOutput := make([]map[string]interface{}, 0, len(examples))
	for _, example := range examples {
		ex := map[string]interface{}{
			"id":          example.ID,
			"name":        example.Name,
			"description": example.Description,
			"type":        example.Type,
			"status":      example.Status,
			"tags":        example.Tags,
			"created":     example.Created,
			"updated":     example.Updated,
		}
		examplesOutput = append(examplesOutput, ex)
	}

	if err := d.Set("examples", examplesOutput); err != nil {
		return diag.FromErr(errors.NewInternalError("error setting examples: %v", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully read %d JumpCloud examples", len(examples)))
	return diag.Diagnostics{}
}
