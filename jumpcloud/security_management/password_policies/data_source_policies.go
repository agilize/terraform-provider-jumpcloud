package password_policies

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// JumpCloudClient is an interface for interaction with the JumpCloud API
type JumpCloudClient interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
	GetApiKey() string
	GetOrgID() string
}

// PasswordPolicyItem represents a JumpCloud password policy
type PasswordPolicyItem struct {
	ID                        string   `json:"_id"`
	Name                      string   `json:"name"`
	Description               string   `json:"description,omitempty"`
	Status                    string   `json:"status"`
	MinLength                 int      `json:"minLength"`
	MaxLength                 int      `json:"maxLength,omitempty"`
	RequireUppercase          bool     `json:"requireUppercase"`
	RequireLowercase          bool     `json:"requireLowercase"`
	RequireNumber             bool     `json:"requireNumber"`
	RequireSymbol             bool     `json:"requireSymbol"`
	MinimumAge                int      `json:"minimumAge,omitempty"`
	ExpirationTime            int      `json:"expirationTime,omitempty"`
	ExpirationWarningTime     int      `json:"expirationWarningTime,omitempty"`
	DisallowPreviousPasswords int      `json:"disallowPreviousPasswords,omitempty"`
	DisallowCommonPasswords   bool     `json:"disallowCommonPasswords"`
	DisallowUsername          bool     `json:"disallowUsername"`
	DisallowNameAndEmail      bool     `json:"disallowNameAndEmail"`
	DisallowPasswordsFromList bool     `json:"disallowPasswordsFromList"`
	Scope                     string   `json:"scope,omitempty"`
	TargetResources           []string `json:"targetResources,omitempty"`
	OrgID                     string   `json:"orgId,omitempty"`
	Created                   string   `json:"created"`
	Updated                   string   `json:"updated"`
}

// PasswordPoliciesResponse represents the API response for listing password policies
type PasswordPoliciesResponse struct {
	Results    []PasswordPolicyItem `json:"results"`
	TotalCount int                  `json:"totalCount"`
}

// DataSourcePolicies returns a data source for JumpCloud password policies
func DataSourcePolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePoliciesRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"operator": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "eq",
							ValidateFunc: validation.StringInSlice([]string{"eq", "ne", "lt", "gt", "le", "ge", "contains"}, false),
						},
					},
				},
				Description: "Filter criteria for the password policies",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Maximum number of password policies to return",
			},
			"skip": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Number of password policies to skip",
			},
			"sort": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeString,
							Required: true,
						},
						"direction": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
						},
					},
				},
				Description: "Sort criteria for the password policies",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID",
			},
			"policies": {
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
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"min_length": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_length": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"require_uppercase": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"require_lowercase": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"require_number": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"require_symbol": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"minimum_age": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"expiration_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"expiration_warning_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"disallow_previous_passwords": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"disallow_common_passwords": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"disallow_username": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"disallow_name_and_email": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"disallow_passwords_from_list": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"scope": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_resources": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"org_id": {
							Type:     schema.TypeString,
							Computed: true,
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
				},
				Description: "List of JumpCloud password policies",
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of password policies found",
			},
		},
	}
}

// dataSourcePoliciesRead reads password policies from JumpCloud
func dataSourcePoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(JumpCloudClient)
	orgID := d.Get("org_id").(string)

	// Build query parameters
	queryParams := url.Values{}

	// Add org_id if provided
	if orgID != "" {
		queryParams.Add("orgId", orgID)
	}

	// Add filters if specified
	if filters, ok := d.GetOk("filter"); ok {
		for i, filter := range filters.([]interface{}) {
			f := filter.(map[string]interface{})
			name := f["name"].(string)
			value := f["value"].(string)
			operator := f["operator"].(string)

			queryParams.Add(fmt.Sprintf("filter[%d].field", i), name)
			queryParams.Add(fmt.Sprintf("filter[%d].value", i), value)
			queryParams.Add(fmt.Sprintf("filter[%d].operator", i), operator)
		}
	}

	// Add sort if specified
	if sortList, ok := d.GetOk("sort"); ok && len(sortList.([]interface{})) > 0 {
		sort := sortList.([]interface{})[0].(map[string]interface{})
		field := sort["field"].(string)
		direction := sort["direction"].(string)

		queryParams.Add("sort", field)
		queryParams.Add("sortDirection", direction)
	}

	// Add limit and skip
	queryParams.Add("limit", strconv.Itoa(d.Get("limit").(int)))
	queryParams.Add("skip", strconv.Itoa(d.Get("skip").(int)))

	// Construct URL
	urlStr := "/api/v2/password-policies"
	if len(queryParams) > 0 {
		urlStr = fmt.Sprintf("%s?%s", urlStr, queryParams.Encode())
	}

	// Make the API request
	resp, err := client.DoRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading password policies: %v", err))
	}

	// Parse response
	var policiesResp PasswordPoliciesResponse
	if err := json.Unmarshal(resp, &policiesResp); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing password policies response: %v", err))
	}

	// Set ID for the data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	// Flatten the policies for the Terraform state
	policies := flattenPasswordPolicies(policiesResp.Results)

	// Set the policies in the state
	if err := d.Set("policies", policies); err != nil {
		return diag.FromErr(fmt.Errorf("error setting policies: %v", err))
	}

	// Set the total count
	if err := d.Set("total", policiesResp.TotalCount); err != nil {
		return diag.FromErr(fmt.Errorf("error setting total count: %v", err))
	}

	return nil
}

// flattenPasswordPolicies converts API policy objects to a format suitable for Terraform state
func flattenPasswordPolicies(policies []PasswordPolicyItem) []map[string]interface{} {
	var result []map[string]interface{}

	for _, policy := range policies {
		policyMap := map[string]interface{}{
			"id":                           policy.ID,
			"name":                         policy.Name,
			"description":                  policy.Description,
			"status":                       policy.Status,
			"min_length":                   policy.MinLength,
			"require_uppercase":            policy.RequireUppercase,
			"require_lowercase":            policy.RequireLowercase,
			"require_number":               policy.RequireNumber,
			"require_symbol":               policy.RequireSymbol,
			"disallow_common_passwords":    policy.DisallowCommonPasswords,
			"disallow_username":            policy.DisallowUsername,
			"disallow_name_and_email":      policy.DisallowNameAndEmail,
			"disallow_passwords_from_list": policy.DisallowPasswordsFromList,
			"scope":                        policy.Scope,
			"target_resources":             policy.TargetResources,
			"org_id":                       policy.OrgID,
			"created":                      policy.Created,
			"updated":                      policy.Updated,
		}

		// Set optional values only if they are present
		if policy.MaxLength > 0 {
			policyMap["max_length"] = policy.MaxLength
		}

		if policy.MinimumAge > 0 {
			policyMap["minimum_age"] = policy.MinimumAge
		}

		if policy.ExpirationTime > 0 {
			policyMap["expiration_time"] = policy.ExpirationTime
		}

		if policy.ExpirationWarningTime > 0 {
			policyMap["expiration_warning_time"] = policy.ExpirationWarningTime
		}

		if policy.DisallowPreviousPasswords > 0 {
			policyMap["disallow_previous_passwords"] = policy.DisallowPreviousPasswords
		}

		result = append(result, policyMap)
	}

	return result
}
