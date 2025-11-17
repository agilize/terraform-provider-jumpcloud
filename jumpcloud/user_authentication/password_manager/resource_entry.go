package password_manager

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

// ResourceEntry returns a schema resource for managing JumpCloud password entries
func ResourceEntry() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEntryCreate,
		ReadContext:   resourceEntryRead,
		UpdateContext: resourceEntryUpdate,
		DeleteContext: resourceEntryDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"safe_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the password safe where the entry will be stored",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the password entry",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the password entry",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"site", "application", "database", "ssh", "server",
					"email", "note", "creditcard", "identity", "file",
					"wifi", "custom",
				}, false),
				Description: "Type of entry (site, application, database, ssh, etc)",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Username associated with the entry",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Password stored",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL associated with the entry (for sites or applications)",
			},
			"notes": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Notes or additional information about the entry",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Tags to categorize the entry",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Additional data specific to the entry type",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"folder": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Folder where the entry will be organized within the safe",
			},
			"favorite": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicates if the entry is marked as a favorite",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the entry",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date of the entry",
			},
			"last_used": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date the entry was last used",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Read:   schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

func resourceEntryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build password entry
	entry := &Entry{
		SafeID: d.Get("safe_id").(string),
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		entry.Description = v.(string)
	}

	if v, ok := d.GetOk("username"); ok {
		entry.Username = v.(string)
	}

	if v, ok := d.GetOk("password"); ok {
		entry.Password = v.(string)
	}

	if v, ok := d.GetOk("url"); ok {
		entry.Url = v.(string)
	}

	if v, ok := d.GetOk("notes"); ok {
		entry.Notes = v.(string)
	}

	if v, ok := d.GetOk("folder"); ok {
		entry.Folder = v.(string)
	}

	if v, ok := d.GetOk("favorite"); ok {
		entry.Favorite = v.(bool)
	}

	// Process tags
	if v, ok := d.GetOk("tags"); ok {
		tagSet := v.(*schema.Set).List()
		tags := make([]string, len(tagSet))
		for i, tag := range tagSet {
			tags[i] = tag.(string)
		}
		entry.Tags = tags
	}

	// Process metadata
	if v, ok := d.GetOk("metadata"); ok {
		metadataMap := v.(map[string]interface{})
		if len(metadataMap) > 0 {
			entry.Metadata = metadataMap
		}
	}

	// Serialize to JSON
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing password entry: %v", err))
	}

	// Create entry via API
	tflog.Debug(ctx, fmt.Sprintf("Creating password entry for safe: %s", entry.SafeID))
	resp, err := client.DoRequest(http.MethodPost, fmt.Sprintf("/api/v2/password-safes/%s/entries", entry.SafeID), entryJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating password entry: %v", err))
	}

	// Deserialize response
	var createdEntry Entry
	if err := json.Unmarshal(resp, &createdEntry); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdEntry.ID == "" {
		return diag.FromErr(fmt.Errorf("password entry created without an ID"))
	}

	d.SetId(createdEntry.ID)
	return resourceEntryRead(ctx, d, meta)
}

func resourceEntryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("password entry ID not provided"))
	}

	safeID := d.Get("safe_id").(string)
	if safeID == "" {
		return diag.FromErr(fmt.Errorf("safe_id not provided"))
	}

	// Get entry via API
	tflog.Debug(ctx, fmt.Sprintf("Reading password entry with ID: %s from safe: %s", id, safeID))
	resp, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/password-safes/%s/entries/%s", safeID, id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Password entry %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading password entry: %v", err))
	}

	// Deserialize response
	var entry Entry
	if err := json.Unmarshal(resp, &entry); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
	if err := d.Set("name", entry.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %v", err))
	}

	if err := d.Set("description", entry.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %v", err))
	}

	if err := d.Set("type", entry.Type); err != nil {
		return diag.FromErr(fmt.Errorf("error setting type: %v", err))
	}

	if err := d.Set("username", entry.Username); err != nil {
		return diag.FromErr(fmt.Errorf("error setting username: %v", err))
	}

	// Don't set password in state unless it changed (sensitive value)

	if err := d.Set("url", entry.Url); err != nil {
		return diag.FromErr(fmt.Errorf("error setting url: %v", err))
	}

	if err := d.Set("notes", entry.Notes); err != nil {
		return diag.FromErr(fmt.Errorf("error setting notes: %v", err))
	}

	if err := d.Set("folder", entry.Folder); err != nil {
		return diag.FromErr(fmt.Errorf("error setting folder: %v", err))
	}

	if err := d.Set("favorite", entry.Favorite); err != nil {
		return diag.FromErr(fmt.Errorf("error setting favorite: %v", err))
	}

	if err := d.Set("created", entry.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %v", err))
	}

	if err := d.Set("updated", entry.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
	}

	if err := d.Set("last_used", entry.LastUsed); err != nil {
		return diag.FromErr(fmt.Errorf("error setting last_used: %v", err))
	}

	if entry.Tags != nil {
		if err := d.Set("tags", entry.Tags); err != nil {
			return diag.FromErr(fmt.Errorf("error setting tags: %v", err))
		}
	}

	if entry.Metadata != nil {
		if err := d.Set("metadata", flattenMetadata(entry.Metadata)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting metadata: %v", err))
		}
	}

	return diags
}

func resourceEntryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("password entry ID not provided"))
	}

	safeID := d.Get("safe_id").(string)
	if safeID == "" {
		return diag.FromErr(fmt.Errorf("safe_id not provided"))
	}

	// Check if anything changed
	if !d.HasChanges("name", "description", "type", "username", "password", "url", "notes",
		"tags", "metadata", "folder", "favorite") {
		return resourceEntryRead(ctx, d, meta)
	}

	// Build updated password entry
	entry := &Entry{
		ID:     id,
		SafeID: safeID,
		Name:   d.Get("name").(string),
		Type:   d.Get("type").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		entry.Description = v.(string)
	}

	if v, ok := d.GetOk("username"); ok {
		entry.Username = v.(string)
	}

	if v, ok := d.GetOk("password"); ok {
		entry.Password = v.(string)
	}

	if v, ok := d.GetOk("url"); ok {
		entry.Url = v.(string)
	}

	if v, ok := d.GetOk("notes"); ok {
		entry.Notes = v.(string)
	}

	if v, ok := d.GetOk("folder"); ok {
		entry.Folder = v.(string)
	}

	if v, ok := d.GetOk("favorite"); ok {
		entry.Favorite = v.(bool)
	}

	// Process tags
	if v, ok := d.GetOk("tags"); ok {
		tagSet := v.(*schema.Set).List()
		tags := make([]string, len(tagSet))
		for i, tag := range tagSet {
			tags[i] = tag.(string)
		}
		entry.Tags = tags
	}

	// Process metadata
	if v, ok := d.GetOk("metadata"); ok {
		metadataMap := v.(map[string]interface{})
		if len(metadataMap) > 0 {
			entry.Metadata = metadataMap
		}
	}

	// Serialize to JSON
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing password entry: %v", err))
	}

	// Update entry via API
	tflog.Debug(ctx, fmt.Sprintf("Updating password entry with ID: %s", id))
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/password-safes/%s/entries/%s", safeID, id), entryJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating password entry: %v", err))
	}

	return resourceEntryRead(ctx, d, meta)
}

func resourceEntryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("password entry ID not provided"))
	}

	safeID := d.Get("safe_id").(string)
	if safeID == "" {
		return diag.FromErr(fmt.Errorf("safe_id not provided"))
	}

	// Delete entry via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting password entry with ID: %s from safe: %s", id, safeID))
	_, err = client.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/password-safes/%s/entries/%s", safeID, id), nil)
	if err != nil {
		if !common.IsNotFoundError(err) {
			return diag.FromErr(fmt.Errorf("error deleting password entry: %v", err))
		}
		// If it's already gone, that's fine
		tflog.Warn(ctx, fmt.Sprintf("Password entry %s was already deleted", id))
	}

	d.SetId("")
	return diags
}

// flattenMetadata converts the metadata map to a format suitable for Terraform state
func flattenMetadata(metadata map[string]interface{}) map[string]interface{} {
	flattenedMetadata := make(map[string]interface{})

	for k, v := range metadata {
		switch v := v.(type) {
		case string, bool, int, float64:
			flattenedMetadata[k] = fmt.Sprintf("%v", v)
		case map[string]interface{}:
			// For nested maps, convert to JSON string
			if nestedJSON, err := json.Marshal(v); err == nil {
				flattenedMetadata[k] = string(nestedJSON)
			}
		case []interface{}:
			// For arrays, convert to JSON string
			if arrayJSON, err := json.Marshal(v); err == nil {
				flattenedMetadata[k] = string(arrayJSON)
			}
		default:
			// For other types, use string representation
			flattenedMetadata[k] = fmt.Sprintf("%v", v)
		}
	}

	return flattenedMetadata
}
