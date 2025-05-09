package app_catalog

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

// DataSourceCategories returns a data source for JumpCloud app catalog categories
func DataSourceCategories() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCategoriesRead,
		Schema: map[string]*schema.Schema{
			"categories": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"display_order": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"parent_category": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"icon_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"applications": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

// dataSourceCategoriesRead reads the app catalog categories from JumpCloud
func dataSourceCategoriesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	tflog.Debug(ctx, "Reading app catalog categories")

	url := "/api/v2/appcatalog/categories"

	resp, err := client.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading app catalog categories: %v", err))
	}

	var categoriesResp common.AppCatalogCategoriesResponse
	if err := json.Unmarshal(resp, &categoriesResp); err != nil {
		return diag.FromErr(fmt.Errorf("error unmarshalling app catalog categories response: %v", err))
	}

	d.SetId(fmt.Sprintf("app-catalog-categories-%d", time.Now().Unix()))

	categories := flattenCategories(categoriesResp.Results)
	if err := d.Set("categories", categories); err != nil {
		return diag.FromErr(fmt.Errorf("error setting categories: %v", err))
	}

	return nil
}

// flattenCategories converts API category objects to a format suitable for Terraform state
func flattenCategories(categories []common.AppCatalogCategory) []map[string]interface{} {
	var result []map[string]interface{}

	for _, category := range categories {
		categoryMap := map[string]interface{}{
			"id":              category.ID,
			"name":            category.Name,
			"description":     category.Description,
			"display_order":   category.DisplayOrder,
			"parent_category": category.ParentCategory,
			"icon_url":        category.IconURL,
			"applications":    category.Applications,
		}

		result = append(result, categoryMap)
	}

	return result
}

// JumpCloudClient is an interface for interaction with the JumpCloud API
type JumpCloudClient interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
	GetApiKey() string
	GetOrgID() string
}

// nolint:unused
func dataSourceAppCatalogCategories() *schema.Resource {
	return &schema.Resource{
		// ... existing code ...
	}
}
