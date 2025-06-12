package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// sanitizeAttributeName ensures attribute names only contain letters and numbers
// as required by the JumpCloud API
func sanitizeAttributeName(name string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(name, "")
}

// convertFieldForAPI converts field names for API usage, automatically detecting custom attributes
func convertFieldForAPI(field string) string {
	// Define standard JumpCloud user fields and their API mappings
	standardFields := map[string]string{
		"company":      "company",
		"costCenter":   "costCenter",
		"department":   "department",
		"description":  "description",
		"employeeType": "employeeType",
		"jobTitle":     "jobTitle",
		"location":     "location",
		"userState":    "state", // userState maps to state in API
		"state":        "state", // state is the correct API field
	}

	// If field already has attributes. prefix, handle it
	if strings.HasPrefix(field, "attributes.") {
		// Convert "attributes.area" to "attributes[name=area].value"
		attrName := strings.TrimPrefix(field, "attributes.")
		return fmt.Sprintf("attributes[name=%s].value", attrName)
	}

	// If it's a standard field, return the mapped API field
	if apiField, exists := standardFields[field]; exists {
		return apiField
	}

	// Otherwise, treat as custom attribute
	return fmt.Sprintf("attributes[name=%s].value", field)
}

// convertFieldFromAPI converts field names from API format back to Terraform format
func convertFieldFromAPI(field string) string {
	// Define reverse mapping from API fields to Terraform fields
	apiToTerraform := map[string]string{
		"company":      "company",
		"costCenter":   "costCenter",
		"department":   "department",
		"description":  "description",
		"employeeType": "employeeType",
		"jobTitle":     "jobTitle",
		"location":     "location",
		"state":        "state", // Keep as state (not userState)
	}

	// If it's a standard field, return the Terraform field name
	if terraformField, exists := apiToTerraform[field]; exists {
		return terraformField
	}

	// Convert "attributes[name=area].value" back to just "area"
	if strings.Contains(field, "attributes[name=") && strings.HasSuffix(field, "].value") {
		start := strings.Index(field, "attributes[name=") + len("attributes[name=")
		end := strings.Index(field, "].value")
		if start < end {
			attrName := field[start:end]
			return attrName // Return just the attribute name, not "attributes.area"
		}
	}

	return field
}

// ResourceUserGroup returns the resource for JumpCloud user groups
func ResourceUserGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserGroupCreate,
		ReadContext:   resourceUserGroupRead,
		UpdateContext: resourceUserGroupUpdate,
		DeleteContext: resourceUserGroupDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the user group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the user group",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "user_group",
				Description: "Type of the group. Default is 'user_group'",
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Custom attributes for the group (key-value pairs)",
			},
			"membership_method": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "STATIC",
				Description: "Method for determining group membership. Valid values are STATIC, DYNAMIC_REVIEW_REQUIRED, or DYNAMIC_AUTOMATED",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					valid := map[string]bool{
						"STATIC":                  true,
						"DYNAMIC_REVIEW_REQUIRED": true,
						"DYNAMIC_AUTOMATED":       true,
					}
					if !valid[v] {
						errs = append(errs, fmt.Errorf("%q must be one of STATIC, DYNAMIC_REVIEW_REQUIRED, or DYNAMIC_AUTOMATED, got: %s", key, v))
					}
					return
				},
			},
			"member_query": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Query for determining dynamic group membership",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "FilterQuery",
							Description: "Type of query. Valid values are 'FilterQuery' (for standard fields) and 'Search' (for custom attributes)",
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								valid := map[string]bool{
									"FilterQuery": true,
									"Search":      true,
								}
								if !valid[v] {
									errs = append(errs, fmt.Errorf("%q must be either 'FilterQuery' or 'Search', got: %s", key, v))
								}
								return
							},
						},
						"filter": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "Filters for the query",
							Set: func(v interface{}) int {
								// Create a hash based on field, operator, and value
								// This ensures that filters with the same content are treated as identical
								// regardless of their order in the list
								m := v.(map[string]interface{})
								field := m["field"].(string)
								operator := m["operator"].(string)
								value := m["value"].(string)
								return schema.HashString(fmt.Sprintf("%s:%s:%s", field, operator, value))
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Field to filter on. Standard fields: company, costCenter, department, description, employeeType, jobTitle, location, userState. Custom attributes: use the attribute name directly (e.g., 'area', 'tribe'). The provider automatically detects if it's a custom attribute.",
									},
									"operator": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Operator for the filter. Valid operators include: eq, ne, in",
										ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
											v := val.(string)
											valid := map[string]bool{
												"eq": true,
												"ne": true,
												"in": true,
												"gt": true,
												"ge": true,
												"lt": true,
												"le": true,
											}
											if !valid[v] {
												errs = append(errs, fmt.Errorf("%q must be one of eq, ne, in, gt, ge, lt, le, got: %s", key, v))
											}
											return
										},
									},
									"value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Value for the filter. For 'in' operator, use pipe-delimited values (e.g., 'value1|value2|value3')",
									},
								},
							},
						},
					},
				},
			},
			"member_query_exemptions": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Users exempted from the dynamic group query",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the user to exempt",
						},
						"type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "USER",
							Description: "Type of the exemption. Currently only USER is supported",
						},
					},
				},
			},
			"member_suggestions_notify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to send email notifications for membership suggestions. Only applicable for DYNAMIC_REVIEW_REQUIRED groups",
			},
			"member_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of users in the group",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date when the group was created",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date when the group was last updated",
			},
		},
	}
}

func resourceUserGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build user group based on resource data
	group := &common.UserGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),
	}

	// Process custom attributes
	if attr, ok := d.GetOk("attributes"); ok {
		attributes := make(map[string]interface{})
		for k, v := range attr.(map[string]interface{}) {
			// Sanitize attribute name for API
			sanitizedName := sanitizeAttributeName(k)
			attributes[sanitizedName] = v
		}
		group.Attributes = attributes
	}

	// Set membership method if provided
	if v, ok := d.GetOk("membership_method"); ok {
		group.MembershipMethod = v.(string)
	}

	// Set member suggestions notify if provided
	if v, ok := d.GetOk("member_suggestions_notify"); ok {
		group.MemberSuggestionsNotify = v.(bool)
	}

	// Process member query if provided
	if v, ok := d.GetOk("member_query"); ok {
		memberQuery, err := expandMemberQuery(v.([]interface{}))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error expanding member query: %v", err))
		}
		group.MemberQuery = memberQuery
	}

	// Process member query exemptions if provided
	if v, ok := d.GetOk("member_query_exemptions"); ok {
		exemptions, err := expandMemberQueryExemptions(v.([]interface{}))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error expanding member query exemptions: %v", err))
		}
		group.MemberQueryExemptions = exemptions
	}

	// Convert to JSON
	groupJSON, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing user group: %v", err))
	}

	// Check if group already exists by name
	name := d.Get("name").(string)
	existingGroups, err := getUserGroupsByName(ctx, c, name)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error checking for existing user group: %v", err))
	}

	var resp []byte

	if len(existingGroups) > 0 {
		// Group already exists, use the first one
		tflog.Info(ctx, fmt.Sprintf("User group with name '%s' already exists, using existing group", name))
		group := existingGroups[0]
		d.SetId(group.ID)

		// Update the existing group
		url := fmt.Sprintf("/api/v2/usergroups/%s", group.ID)
		tflog.Debug(ctx, fmt.Sprintf("Updating existing user group with URL: %s", url))
		tflog.Debug(ctx, fmt.Sprintf("Request body: %s", string(groupJSON)))

		resp, err = c.DoRequest(http.MethodPut, url, groupJSON)
		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Error updating existing user group: %v", err))
			return diag.FromErr(fmt.Errorf("error updating existing user group: %v", err))
		}
	} else {
		// Create new group
		url := "/api/v2/usergroups"
		tflog.Debug(ctx, fmt.Sprintf("Creating user group with URL: %s", url))
		tflog.Debug(ctx, fmt.Sprintf("Request body: %s", string(groupJSON)))

		resp, err = c.DoRequest(http.MethodPost, url, groupJSON)
		if err != nil {
			// Check if the error is "Already Exists"
			if strings.Contains(err.Error(), "Already Exists") {
				// Try to get the group by name again
				name := d.Get("name").(string)
				tflog.Info(ctx, fmt.Sprintf("Group already exists, trying to get group by name: %s", name))

				existingGroups, err2 := getUserGroupsByName(ctx, c, name)
				if err2 != nil || len(existingGroups) == 0 {
					return diag.FromErr(fmt.Errorf("error creating user group and failed to find by name: %v, %v", err, err2))
				}

				// Use the existing group
				group := existingGroups[0]
				d.SetId(group.ID)

				// Return early with success
				return resourceUserGroupRead(ctx, d, meta)
			}

			tflog.Error(ctx, fmt.Sprintf("Error creating user group: %v", err))
			return diag.FromErr(fmt.Errorf("error creating user group: %v", err))
		}
	}

	// Try to decode response as a single object first
	var newGroup common.UserGroup
	if err := json.Unmarshal(resp, &newGroup); err != nil {
		// If that fails, try to decode as an array
		var groups []common.UserGroup
		if err2 := json.Unmarshal(resp, &groups); err2 != nil {
			return diag.FromErr(fmt.Errorf("error deserializing group response: %v, %v", err, err2))
		}

		// Use the first group in the array
		if len(groups) == 0 {
			return diag.FromErr(fmt.Errorf("no groups returned from API"))
		}
		newGroup = groups[0]
	}

	// Set ID in terraform state
	d.SetId(newGroup.ID)

	// Read the group to ensure all computed fields are set
	return resourceUserGroupRead(ctx, d, meta)
}

func resourceUserGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	// Get group via API
	url := fmt.Sprintf("/api/v2/usergroups/%s", groupID)
	tflog.Debug(ctx, fmt.Sprintf("Reading user group with URL: %s", url))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		// Handle 404 specifically to mark the resource as removed
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("User group %s not found, removing from state", groupID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading user group %s: %v", groupID, err))
	}

	// Try to decode response as a single object first
	var group common.UserGroup
	if err := json.Unmarshal(resp, &group); err != nil {
		// If that fails, try to decode as an array
		var groups []common.UserGroup
		if err2 := json.Unmarshal(resp, &groups); err2 != nil {
			// If both fail, try to get the group by name
			name := d.Get("name").(string)
			tflog.Info(ctx, fmt.Sprintf("Failed to unmarshal response, trying to get group by name: %s", name))

			existingGroups, err3 := getUserGroupsByName(ctx, c, name)
			if err3 != nil || len(existingGroups) == 0 {
				return diag.FromErr(fmt.Errorf("error deserializing group response and failed to find by name: %v, %v, %v", err, err2, err3))
			}

			group = existingGroups[0]
		} else {
			// Use the first group in the array
			if len(groups) == 0 {
				// If no groups returned, try to get the group by name
				name := d.Get("name").(string)
				tflog.Info(ctx, fmt.Sprintf("No groups returned, trying to get group by name: %s", name))

				existingGroups, err3 := getUserGroupsByName(ctx, c, name)
				if err3 != nil || len(existingGroups) == 0 {
					return diag.FromErr(fmt.Errorf("no groups returned from API and failed to find by name: %v", err3))
				}

				group = existingGroups[0]
			} else {
				group = groups[0]
			}
		}
	}

	// Debug log the group data
	groupJSON, _ := json.Marshal(group)
	tflog.Debug(ctx, fmt.Sprintf("Group data: %s", string(groupJSON)))

	// Set ID in terraform state if not already set
	if d.Id() == "" {
		d.SetId(group.ID)
	}

	// Set fields in terraform state
	if err := d.Set("name", group.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %v", err))
	}

	if err := d.Set("description", group.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %v", err))
	}

	if err := d.Set("type", group.Type); err != nil {
		return diag.FromErr(fmt.Errorf("error setting type: %v", err))
	}

	// Set dynamic group fields
	if err := d.Set("membership_method", group.MembershipMethod); err != nil {
		return diag.FromErr(fmt.Errorf("error setting membership_method: %v", err))
	}

	if err := d.Set("member_suggestions_notify", group.MemberSuggestionsNotify); err != nil {
		return diag.FromErr(fmt.Errorf("error setting member_suggestions_notify: %v", err))
	}

	// Set member query if present
	if group.MemberQuery != nil {
		flattened := flattenMemberQuery(group.MemberQuery)
		fmt.Printf("DEBUG: Setting member_query to: %+v\n", flattened)
		if err := d.Set("member_query", flattened); err != nil {
			return diag.FromErr(fmt.Errorf("error setting member_query: %v", err))
		}
	}

	// Set member query exemptions if present
	if len(group.MemberQueryExemptions) > 0 {
		if err := d.Set("member_query_exemptions", flattenMemberQueryExemptions(group.MemberQueryExemptions)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting member_query_exemptions: %v", err))
		}
	}

	// Process attributes, preserving original names
	if group.Attributes != nil {
		// Get the original attributes from the configuration
		oldAttrs := d.Get("attributes").(map[string]interface{})

		// Create a map of sanitized name -> original name
		sanitizedToOriginal := make(map[string]string)
		for origName := range oldAttrs {
			sanitizedToOriginal[sanitizeAttributeName(origName)] = origName
		}

		// Create new attribute map preserving original names
		attributes := make(map[string]interface{})
		for attrName, attrValue := range group.Attributes {
			// Skip the ldapGroups attribute as it's managed by JumpCloud
			if attrName == "ldapGroups" {
				continue
			}

			// Check if we have this attribute in the old configuration
			if origName, exists := sanitizedToOriginal[attrName]; exists {
				// Use the original name
				attributes[origName] = fmt.Sprintf("%v", attrValue)
			} else {
				// Use the name from the API
				attributes[attrName] = fmt.Sprintf("%v", attrValue)
			}
		}

		// Only set attributes if we have any (avoid setting empty map)
		if len(attributes) > 0 {
			if err := d.Set("attributes", attributes); err != nil {
				return diag.FromErr(fmt.Errorf("error setting attributes: %v", err))
			}
		} else if len(oldAttrs) > 0 {
			// If we had attributes before but now they're all filtered out,
			// keep the original attributes to avoid unnecessary changes
			if err := d.Set("attributes", oldAttrs); err != nil {
				return diag.FromErr(fmt.Errorf("error setting attributes: %v", err))
			}
		}
	}

	return diags
}

func resourceUserGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	// Build user group based on resource data
	group := &common.UserGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Type:        d.Get("type").(string),
	}

	// Process custom attributes
	if attr, ok := d.GetOk("attributes"); ok {
		attributes := make(map[string]interface{})
		for k, v := range attr.(map[string]interface{}) {
			// Sanitize attribute name for API
			sanitizedName := sanitizeAttributeName(k)
			attributes[sanitizedName] = v
		}
		group.Attributes = attributes
	}

	// Set membership method if provided
	if v, ok := d.GetOk("membership_method"); ok {
		group.MembershipMethod = v.(string)
	}

	// Set member suggestions notify if provided
	if v, ok := d.GetOk("member_suggestions_notify"); ok {
		group.MemberSuggestionsNotify = v.(bool)
	}

	// Process member query if provided
	if v, ok := d.GetOk("member_query"); ok {
		memberQuery, err := expandMemberQuery(v.([]interface{}))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error expanding member query: %v", err))
		}
		group.MemberQuery = memberQuery
	}

	// Process member query exemptions if provided
	if v, ok := d.GetOk("member_query_exemptions"); ok {
		exemptions, err := expandMemberQueryExemptions(v.([]interface{}))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error expanding member query exemptions: %v", err))
		}
		group.MemberQueryExemptions = exemptions
	}

	// Convert to JSON
	groupJSON, err := json.Marshal(group)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing user group: %v", err))
	}

	// Update group via API
	url := fmt.Sprintf("/api/v2/usergroups/%s", groupID)
	tflog.Debug(ctx, fmt.Sprintf("Updating user group with URL: %s", url))
	tflog.Debug(ctx, fmt.Sprintf("Request body: %s", string(groupJSON)))
	_, err = c.DoRequest(http.MethodPut, url, groupJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating user group %s: %v", groupID, err))
	}

	return resourceUserGroupRead(ctx, d, meta)
}

func resourceUserGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	groupID := d.Id()

	// Delete group via API
	url := fmt.Sprintf("/api/v2/usergroups/%s", groupID)
	tflog.Debug(ctx, fmt.Sprintf("Deleting user group with URL: %s", url))
	_, err := c.DoRequest(http.MethodDelete, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting user group %s: %v", groupID, err))
	}

	// Set ID to empty to indicate resource has been removed
	d.SetId("")

	return nil
}

// Helper functions for dynamic groups

// expandMemberQuery converts the Terraform schema representation of a member query to the API format
func expandMemberQuery(input []interface{}) (interface{}, error) {
	if len(input) == 0 || input[0] == nil {
		return nil, nil
	}

	queryMap := input[0].(map[string]interface{})
	queryType := queryMap["query_type"].(string)

	// Handle both TypeSet and TypeList for backward compatibility
	var filters []interface{}
	if filterSet, ok := queryMap["filter"].(*schema.Set); ok {
		filters = filterSet.List()
	} else if filterList, ok := queryMap["filter"].([]interface{}); ok {
		filters = filterList
	}

	if len(filters) > 0 {
		switch queryType {
		case "FilterQuery":
			// Use FilterQuery specific type
			query := &common.UserGroupFilterQuery{
				QueryType: queryType,
				Filters:   make([]common.UserGroupFilter, 0, len(filters)),
			}

			for _, f := range filters {
				filterMap := f.(map[string]interface{})
				filter := common.UserGroupFilter{
					Field:    filterMap["field"].(string),
					Operator: filterMap["operator"].(string),
					Value:    filterMap["value"].(string),
				}
				query.Filters = append(query.Filters, filter)
			}

			return query, nil
		case "Search":
			// Convert filters to searchFilters format for Search query
			var searchFilters []string
			for _, f := range filters {
				filterMap := f.(map[string]interface{})
				field := filterMap["field"].(string)
				operator := filterMap["operator"].(string)
				value := filterMap["value"].(string)

				// Auto-detect if field is a custom attribute
				field = convertFieldForAPI(field)

				// Convert operator format and handle special cases
				if operator == "in" {
					// For 'in' operator, split values and create separate filters with $eq
					values := strings.Split(value, "|")
					for _, val := range values {
						val = strings.TrimSpace(val)
						if val != "" {
							searchFilter := fmt.Sprintf("%s:$eq:%s", field, val)
							searchFilters = append(searchFilters, searchFilter)
						}
					}
				} else {
					var searchOp string
					switch operator {
					case "eq":
						searchOp = "$eq"
					case "ne":
						searchOp = "$ne"
					default:
						searchOp = "$eq" // default to equals
					}

					searchFilter := fmt.Sprintf("%s:%s:%s", field, searchOp, value)
					searchFilters = append(searchFilters, searchFilter)
				}
			}

			// Create SearchQuery specific type
			if len(searchFilters) > 0 {
				// Create searchFilters as JSON string (as expected by JumpCloud API)
				searchFiltersObj := map[string]interface{}{
					"filter": searchFilters,
				}
				searchFiltersJSON, _ := json.Marshal(searchFiltersObj)

				query := &common.UserGroupSearchQuery{
					QueryType:     queryType,
					SearchFilters: string(searchFiltersJSON), // Send as string, not object
				}

				return query, nil
			}
		}
	}

	// Return basic query if no filters
	return &common.UserGroupQuery{QueryType: queryType}, nil
}

// expandMemberQueryExemptions converts the Terraform schema representation of member query exemptions to the API format
func expandMemberQueryExemptions(input []interface{}) ([]common.UserGroupExemption, error) {
	if len(input) == 0 {
		return nil, nil
	}

	exemptions := make([]common.UserGroupExemption, 0, len(input))
	for _, item := range input {
		exemptionMap := item.(map[string]interface{})
		exemption := common.UserGroupExemption{
			ID:   exemptionMap["id"].(string),
			Type: exemptionMap["type"].(string),
		}
		exemptions = append(exemptions, exemption)
	}

	return exemptions, nil
}

// flattenMemberQuery converts the API representation of a member query to the Terraform schema format
func flattenMemberQuery(query interface{}) []interface{} {
	if query == nil {
		return []interface{}{}
	}

	result := make(map[string]interface{})
	var filters []interface{}

	switch q := query.(type) {
	case *common.UserGroupFilterQuery:
		result["query_type"] = q.QueryType
		if len(q.Filters) > 0 {
			// Sort filters by field name for consistent ordering
			sortedFilters := make([]common.UserGroupFilter, len(q.Filters))
			copy(sortedFilters, q.Filters)
			sort.Slice(sortedFilters, func(i, j int) bool {
				return sortedFilters[i].Field < sortedFilters[j].Field
			})

			filters = make([]interface{}, 0, len(sortedFilters))
			for _, filter := range sortedFilters {
				filterMap := make(map[string]interface{})
				filterMap["field"] = filter.Field
				filterMap["operator"] = filter.Operator
				filterMap["value"] = filter.Value
				filters = append(filters, filterMap)
			}
		}
	case *common.UserGroupSearchQuery:
		result["query_type"] = q.QueryType
		if q.SearchFilters != "" {
			filters = parseSearchFilters(q.SearchFilters)
		}
	case *common.UserGroupQuery:
		result["query_type"] = q.QueryType
		if q.QueryType == "FilterQuery" && len(q.Filters) > 0 {
			filters = make([]interface{}, 0, len(q.Filters))
			for _, filter := range q.Filters {
				filterMap := make(map[string]interface{})
				filterMap["field"] = filter.Field
				filterMap["operator"] = filter.Operator
				filterMap["value"] = filter.Value
				filters = append(filters, filterMap)
			}
		} else if q.QueryType == "Search" && q.SearchFilters != nil {
			filters = parseSearchFiltersFromObject(q.SearchFilters)
		}
	case map[string]interface{}:
		// Handle the case when JSON is deserialized as generic map
		if queryType, ok := q["queryType"].(string); ok {
			result["query_type"] = queryType

			switch queryType {
			case "FilterQuery":
				if filtersArray, ok := q["filters"].([]interface{}); ok {
					// Create a temporary slice to sort filters
					var tempFilters []map[string]interface{}
					for _, f := range filtersArray {
						if filterMap, ok := f.(map[string]interface{}); ok {
							newFilter := make(map[string]interface{})
							if field, ok := filterMap["field"].(string); ok {
								newFilter["field"] = field
							}
							if operator, ok := filterMap["operator"].(string); ok {
								newFilter["operator"] = operator
							}
							if value, ok := filterMap["value"].(string); ok {
								newFilter["value"] = value
							}
							tempFilters = append(tempFilters, newFilter)
						}
					}

					// Sort filters by field name for consistent ordering
					sort.Slice(tempFilters, func(i, j int) bool {
						return tempFilters[i]["field"].(string) < tempFilters[j]["field"].(string)
					})

					// Convert to interface{} slice
					filters = make([]interface{}, 0, len(tempFilters))
					for _, filter := range tempFilters {
						filters = append(filters, filter)
					}
				}
			case "Search":
				if searchFilters, ok := q["searchFilters"].(string); ok {
					filters = parseSearchFilters(searchFilters)
				}
			}
		}
	default:
		return []interface{}{}
	}

	result["filter"] = filters
	return []interface{}{result}
}

// parseSearchFilters converts searchFilters JSON string back to filter array format
func parseSearchFilters(searchFilters string) []interface{} {
	var filters []interface{}

	// Parse the JSON string to extract filters
	// Example: {"filter":["company:$eq:Agilize","department:$eq:Administration","state:$eq:ACTIVATED","attributes[name=area].value:$eq:asd"]}

	var searchFiltersObj map[string]interface{}
	if err := json.Unmarshal([]byte(searchFilters), &searchFiltersObj); err != nil {
		return filters
	}

	if filterArray, ok := searchFiltersObj["filter"].([]interface{}); ok {
		// Group filters by field to handle multiple values (like jobTitle)
		fieldGroups := make(map[string][]string)

		for _, filterItem := range filterArray {
			if filterStr, ok := filterItem.(string); ok {
				parts := strings.Split(filterStr, ":")
				if len(parts) >= 3 {
					field := parts[0]
					value := strings.Join(parts[2:], ":") // In case value contains colons

					// Convert back from API format
					field = convertFieldFromAPI(field)

					// Group values by field
					if _, exists := fieldGroups[field]; !exists {
						fieldGroups[field] = []string{}
					}
					fieldGroups[field] = append(fieldGroups[field], value)
				}
			}
		}

		// Convert grouped filters back to Terraform format
		// Sort fields to ensure consistent ordering
		var sortedFields []string
		for field := range fieldGroups {
			sortedFields = append(sortedFields, field)
		}
		sort.Strings(sortedFields)

		for _, field := range sortedFields {
			values := fieldGroups[field]
			filterMap := make(map[string]interface{})
			filterMap["field"] = field

			if len(values) > 1 {
				// Multiple values = "in" operator
				filterMap["operator"] = "in"
				filterMap["value"] = strings.Join(values, "|")
			} else {
				// Single value = "eq" operator
				filterMap["operator"] = "eq"
				filterMap["value"] = values[0]
			}

			filters = append(filters, filterMap)
		}
	}

	return filters
}

// parseSearchFiltersFromObject converts searchFilters object back to filter array format
func parseSearchFiltersFromObject(searchFilters map[string]interface{}) []interface{} {
	var filters []interface{}

	if filterArray, ok := searchFilters["filter"].([]interface{}); ok {
		for _, filterItem := range filterArray {
			if filterStr, ok := filterItem.(string); ok {
				parts := strings.Split(filterStr, ":")
				if len(parts) >= 3 {
					field := parts[0]
					operator := parts[1]
					value := strings.Join(parts[2:], ":") // In case value contains colons

					// Convert back from API format
					field = convertFieldFromAPI(field)

					// Convert operator format
					switch operator {
					case "$eq":
						operator = "eq"
					case "$ne":
						operator = "ne"
					case "$in":
						operator = "in"
					}

					filterMap := make(map[string]interface{})
					filterMap["field"] = field
					filterMap["operator"] = operator
					filterMap["value"] = value
					filters = append(filters, filterMap)
				}
			}
		}
	}

	return filters
}

// flattenMemberQueryExemptions converts the API representation of member query exemptions to the Terraform schema format
func flattenMemberQueryExemptions(exemptions []common.UserGroupExemption) []interface{} {
	if len(exemptions) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, 0, len(exemptions))
	for _, exemption := range exemptions {
		exemptionMap := make(map[string]interface{})
		exemptionMap["id"] = exemption.ID
		exemptionMap["type"] = exemption.Type
		result = append(result, exemptionMap)
	}

	return result
}

// getUserGroupsByName retrieves user groups by name
func getUserGroupsByName(ctx context.Context, c common.ClientInterface, name string) ([]common.UserGroup, error) {
	// Get all user groups
	resp, err := c.DoRequest(http.MethodGet, "/api/v2/usergroups", nil)
	if err != nil {
		return nil, fmt.Errorf("error listing user groups: %v", err)
	}

	// Parse response
	var groups []common.UserGroup
	if err := json.Unmarshal(resp, &groups); err != nil {
		return nil, fmt.Errorf("error parsing user groups response: %v", err)
	}

	// Filter by name
	var matchingGroups []common.UserGroup
	for _, group := range groups {
		if group.Name == name {
			matchingGroups = append(matchingGroups, group)
		}
	}

	return matchingGroups, nil
}
