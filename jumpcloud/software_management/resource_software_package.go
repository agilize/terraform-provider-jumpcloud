package software_management

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// SoftwarePackage represents a software package in JumpCloud
type SoftwarePackage struct {
	ID          string                 `json:"_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Version     string                 `json:"version"`
	Type        string                 `json:"type"` // windows, macos, linux
	URL         string                 `json:"url,omitempty"`
	FilePath    string                 `json:"filePath,omitempty"`
	FileSize    int64                  `json:"fileSize,omitempty"`
	SHA256      string                 `json:"sha256,omitempty"`
	MD5         string                 `json:"md5,omitempty"`
	Status      string                 `json:"status,omitempty"` // active, inactive, processing, error
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	OrgID       string                 `json:"orgId,omitempty"`
	Created     string                 `json:"created,omitempty"`
	Updated     string                 `json:"updated,omitempty"`
}

// ResourceSoftwarePackage returns a schema resource for managing software packages
func ResourceSoftwarePackage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSoftwarePackageCreate,
		ReadContext:   resourceSoftwarePackageRead,
		UpdateContext: resourceSoftwarePackageUpdate,
		DeleteContext: resourceSoftwarePackageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Name of the software package",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the software package",
			},
			"version": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
				Description:  "Version of the software package",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"windows", "macos", "linux"}, false),
				Description:  "Type of the software package (windows, macos, linux)",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL where the software package can be downloaded",
			},
			"file_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path to the software package file",
			},
			"file_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Size of the software package file in bytes",
			},
			"sha256": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SHA256 hash of the software package file",
			},
			"md5": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "MD5 hash of the software package file",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the software package (active, inactive, processing, error)",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Additional metadata for the software package",
			},
			"parameters": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Parameters for the software package installation",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Tags associated with the software package",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for the software package",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update timestamp",
			},
		},
	}
}

func resourceSoftwarePackageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	// Create package object from resource data
	pkg := SoftwarePackage{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Version:     d.Get("version").(string),
		Type:        d.Get("type").(string),
		URL:         d.Get("url").(string),
		FilePath:    d.Get("file_path").(string),
		FileSize:    int64(d.Get("file_size").(int)),
		SHA256:      d.Get("sha256").(string),
		MD5:         d.Get("md5").(string),
	}

	// Handle maps
	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			metadata[k] = v
		}
		pkg.Metadata = metadata
	}

	if v, ok := d.GetOk("parameters"); ok {
		parameters := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			parameters[k] = v
		}
		pkg.Parameters = parameters
	}

	// Handle lists
	if v, ok := d.GetOk("tags"); ok {
		tagsList := v.([]interface{})
		tags := make([]string, len(tagsList))
		for i, v := range tagsList {
			tags[i] = v.(string)
		}
		pkg.Tags = tags
	}

	// Set org_id if provided
	if v, ok := d.GetOk("org_id"); ok {
		pkg.OrgID = v.(string)
	}

	// Convert to JSON
	reqBody, err := json.Marshal(pkg)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing software package: %v", err))
	}

	// Create software package via API
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/software/packages", reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating software package: %v", err))
	}

	// Parse response
	var createdPkg SoftwarePackage
	if err := json.Unmarshal(resp, &createdPkg); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing software package response: %v", err))
	}

	d.SetId(createdPkg.ID)
	tflog.Trace(ctx, "Created software package", map[string]interface{}{
		"id": d.Id(),
	})

	return resourceSoftwarePackageRead(ctx, d, meta)
}

func resourceSoftwarePackageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()

	// Get software package via API
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/software/packages/%s", id), nil)
	if err != nil {
		// Handle 404 specifically
		if err.Error() == "status code 404" {
			tflog.Warn(ctx, fmt.Sprintf("Software package %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading software package %s: %v", id, err))
	}

	// Decode response
	var pkg SoftwarePackage
	if err := json.Unmarshal(resp, &pkg); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing software package response: %v", err))
	}

	// Set the resource data
	d.Set("name", pkg.Name)
	d.Set("description", pkg.Description)
	d.Set("version", pkg.Version)
	d.Set("type", pkg.Type)
	d.Set("url", pkg.URL)
	d.Set("file_path", pkg.FilePath)
	d.Set("file_size", pkg.FileSize)
	d.Set("sha256", pkg.SHA256)
	d.Set("md5", pkg.MD5)
	d.Set("status", pkg.Status)
	d.Set("org_id", pkg.OrgID)
	d.Set("created", pkg.Created)
	d.Set("updated", pkg.Updated)

	// Handle maps
	if pkg.Metadata != nil {
		if err := d.Set("metadata", pkg.Metadata); err != nil {
			return diag.FromErr(err)
		}
	}

	if pkg.Parameters != nil {
		if err := d.Set("parameters", pkg.Parameters); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle lists
	if pkg.Tags != nil {
		if err := d.Set("tags", pkg.Tags); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceSoftwarePackageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()

	// Create package object from resource data
	pkg := SoftwarePackage{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Version:     d.Get("version").(string),
		Type:        d.Get("type").(string),
		URL:         d.Get("url").(string),
		FilePath:    d.Get("file_path").(string),
		FileSize:    int64(d.Get("file_size").(int)),
		SHA256:      d.Get("sha256").(string),
		MD5:         d.Get("md5").(string),
	}

	// Handle maps
	if v, ok := d.GetOk("metadata"); ok {
		metadata := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			metadata[k] = v
		}
		pkg.Metadata = metadata
	}

	if v, ok := d.GetOk("parameters"); ok {
		parameters := make(map[string]interface{})
		for k, v := range v.(map[string]interface{}) {
			parameters[k] = v
		}
		pkg.Parameters = parameters
	}

	// Handle lists
	if v, ok := d.GetOk("tags"); ok {
		tagsList := v.([]interface{})
		tags := make([]string, len(tagsList))
		for i, v := range tagsList {
			tags[i] = v.(string)
		}
		pkg.Tags = tags
	}

	// Set org_id if provided
	if v, ok := d.GetOk("org_id"); ok {
		pkg.OrgID = v.(string)
	}

	// Convert to JSON
	reqBody, err := json.Marshal(pkg)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing software package: %v", err))
	}

	// Update software package via API
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/software/packages/%s", id), reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating software package %s: %v", id, err))
	}

	tflog.Trace(ctx, "Updated software package", map[string]interface{}{
		"id": d.Id(),
	})

	return resourceSoftwarePackageRead(ctx, d, meta)
}

func resourceSoftwarePackageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := common.GetClientFromMeta(meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()

	// Delete software package via API
	_, err := client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/software/packages/%s", id), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting software package %s: %v", id, err))
	}

	// Set ID to empty to signify resource has been removed
	d.SetId("")
	tflog.Trace(ctx, "Deleted software package", map[string]interface{}{
		"id": id,
	})

	return diags
}
