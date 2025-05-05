package users_directory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// DataSourceUser returns a schema for the JumpCloud user data source
func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"username", "email"},
				Description:   "The ID of the user to retrieve",
			},
			"username": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_id", "email"},
				Description:   "The username of the user to retrieve",
			},
			"email": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"user_id", "username"},
				Description:   "The email of the user to retrieve",
			},
			// Output fields
			"firstname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The first name of the user",
			},
			"lastname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last name of the user",
			},
			"middlename": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The middle name of the user",
			},
			"displayname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The display name of the user",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the user",
			},
			"attributes": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Custom attributes for the user",
			},
			"mfa_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether MFA is enabled for the user",
			},
			"password_never_expires": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the password never expires",
			},
			"alternate_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An alternate email address for the user",
			},
			"company": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The company the user belongs to",
			},
			"cost_center": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The cost center the user is associated with",
			},
			"department": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The department the user belongs to",
			},
			"employee_identifier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An identifier for the employee",
			},
			"employee_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of employee",
			},
			"job_title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The job title of the user",
			},
			"location": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The location of the user",
			},
			"enable_managed_uid": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether managed UID is enabled for the user",
			},
			"enable_user_portal_multifactor": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether multifactor authentication is enabled for the user portal",
			},
			"externally_managed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user is externally managed",
			},
			"ldap_binding_user": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user is an LDAP binding user",
			},
			"passwordless_sudo": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether passwordless sudo is enabled for the user",
			},
			"global_passwordless_sudo": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether global passwordless sudo is enabled for the user",
			},
			"public_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public SSH key for the user",
			},
			"allow_public_key": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether public key authentication is allowed for the user",
			},
			"samba_service_user": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user is a Samba service user",
			},
			"sudo": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether sudo access is granted to the user",
			},
			"suspended": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user is suspended",
			},
			"unix_uid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The Unix UID for the user",
			},
			"unix_guid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The Unix GUID for the user",
			},
			"disable_device_max_login_attempts": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether maximum login attempts are disabled for the user's devices",
			},
			"password_recovery_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email address used for password recovery",
			},
			"enforce_uid_gid_consistency": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether UID/GID consistency is enforced",
			},
			"delegated_authority": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The delegated authority for the user",
			},
			"password_authority": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The password authority for the user",
			},
			"managed_apple_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The managed Apple ID for the user",
			},
			"bypass_managed_device_lockout": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether managed device lockout is bypassed for the user",
			},
			"manager_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the user's manager in JumpCloud",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date the user was created",
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	var path string
	var searchType string

	// Determine search method based on provided parameters
	if userID, ok := d.GetOk("user_id"); ok {
		// Direct lookup by ID
		path = fmt.Sprintf("/systemusers/%s", userID.(string))
		searchType = "ID"
	} else if _, ok := d.GetOk("username"); ok {
		// For username, we'll get all users and filter client-side
		// This is more reliable than using the search endpoint
		path = "/systemusers"
		searchType = "username"
	} else if _, ok := d.GetOk("email"); ok {
		// For email, we'll get all users and filter client-side
		path = "/systemusers"
		searchType = "email"
	} else {
		return diag.FromErr(fmt.Errorf("one of user_id, username, or email must be provided"))
	}

	tflog.Debug(ctx, fmt.Sprintf("Reading JumpCloud user by %s", searchType))

	// Make API request
	resp, err := c.DoRequest(http.MethodGet, path, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading user: %v", err))
	}

	// Handle search results vs direct ID lookup
	var user User
	if searchType == "ID" {
		// Direct lookup by ID returns a single user object
		if err := json.Unmarshal(resp, &user); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing user response: %v", err))
		}
	} else {
		// For username or email, we get all users and filter client-side
		// The API returns a response with a "results" array
		var response struct {
			Results    []User `json:"results"`
			TotalCount int    `json:"totalCount"`
		}

		if err := json.Unmarshal(resp, &response); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing users response: %v", err))
		}

		// Filter users based on search type
		var matchedUsers []User
		searchValue := d.Get(searchType).(string)

		for _, u := range response.Results {
			if searchType == "username" && u.Username == searchValue {
				matchedUsers = append(matchedUsers, u)
			} else if searchType == "email" && u.Email == searchValue {
				matchedUsers = append(matchedUsers, u)
			}
		}

		if len(matchedUsers) == 0 {
			return diag.FromErr(fmt.Errorf("no user found with %s: %s", searchType, searchValue))
		}

		if len(matchedUsers) > 1 {
			tflog.Warn(ctx, fmt.Sprintf("Multiple users found with %s: %s, using the first one", searchType, searchValue))
		}

		user = matchedUsers[0]
	}

	// Set the ID
	d.SetId(user.ID)

	// Set all the computed fields
	if err := setUserFields(d, &user); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// setUserFields sets all the user fields in the schema
func setUserFields(d *schema.ResourceData, user *User) error {
	if err := d.Set("username", user.Username); err != nil {
		return fmt.Errorf("error setting username: %v", err)
	}
	if err := d.Set("email", user.Email); err != nil {
		return fmt.Errorf("error setting email: %v", err)
	}
	if err := d.Set("firstname", user.FirstName); err != nil {
		return fmt.Errorf("error setting firstname: %v", err)
	}
	if err := d.Set("lastname", user.LastName); err != nil {
		return fmt.Errorf("error setting lastname: %v", err)
	}
	if err := d.Set("middlename", user.MiddleName); err != nil {
		return fmt.Errorf("error setting middlename: %v", err)
	}
	if err := d.Set("displayname", user.DisplayName); err != nil {
		return fmt.Errorf("error setting displayname: %v", err)
	}
	if err := d.Set("description", user.Description); err != nil {
		return fmt.Errorf("error setting description: %v", err)
	}
	if err := d.Set("mfa_enabled", user.MFAEnabled); err != nil {
		return fmt.Errorf("error setting mfa_enabled: %v", err)
	}
	if err := d.Set("password_never_expires", user.PasswordNeverExpires); err != nil {
		return fmt.Errorf("error setting password_never_expires: %v", err)
	}
	if err := d.Set("alternate_email", user.AlternateEmail); err != nil {
		return fmt.Errorf("error setting alternate_email: %v", err)
	}
	if err := d.Set("company", user.Company); err != nil {
		return fmt.Errorf("error setting company: %v", err)
	}
	if err := d.Set("cost_center", user.CostCenter); err != nil {
		return fmt.Errorf("error setting cost_center: %v", err)
	}
	if err := d.Set("department", user.Department); err != nil {
		return fmt.Errorf("error setting department: %v", err)
	}
	if err := d.Set("employee_identifier", user.EmployeeIdentifier); err != nil {
		return fmt.Errorf("error setting employee_identifier: %v", err)
	}
	if err := d.Set("employee_type", user.EmployeeType); err != nil {
		return fmt.Errorf("error setting employee_type: %v", err)
	}
	if err := d.Set("job_title", user.JobTitle); err != nil {
		return fmt.Errorf("error setting job_title: %v", err)
	}
	if err := d.Set("location", user.Location); err != nil {
		return fmt.Errorf("error setting location: %v", err)
	}
	if err := d.Set("enable_managed_uid", user.EnableManagedUID); err != nil {
		return fmt.Errorf("error setting enable_managed_uid: %v", err)
	}
	if err := d.Set("enable_user_portal_multifactor", user.EnableUserPortalMultifactor); err != nil {
		return fmt.Errorf("error setting enable_user_portal_multifactor: %v", err)
	}
	if err := d.Set("externally_managed", user.ExternallyManaged); err != nil {
		return fmt.Errorf("error setting externally_managed: %v", err)
	}
	if err := d.Set("ldap_binding_user", user.LDAPBindingUser); err != nil {
		return fmt.Errorf("error setting ldap_binding_user: %v", err)
	}
	if err := d.Set("passwordless_sudo", user.PasswordlessSudo); err != nil {
		return fmt.Errorf("error setting passwordless_sudo: %v", err)
	}
	if err := d.Set("global_passwordless_sudo", user.GlobalPasswordlessSudo); err != nil {
		return fmt.Errorf("error setting global_passwordless_sudo: %v", err)
	}
	if err := d.Set("public_key", user.PublicKey); err != nil {
		return fmt.Errorf("error setting public_key: %v", err)
	}
	if err := d.Set("allow_public_key", user.AllowPublicKey); err != nil {
		return fmt.Errorf("error setting allow_public_key: %v", err)
	}
	if err := d.Set("samba_service_user", user.SambaServiceUser); err != nil {
		return fmt.Errorf("error setting samba_service_user: %v", err)
	}
	if err := d.Set("sudo", user.Sudo); err != nil {
		return fmt.Errorf("error setting sudo: %v", err)
	}
	if err := d.Set("suspended", user.Suspended); err != nil {
		return fmt.Errorf("error setting suspended: %v", err)
	}
	if err := d.Set("unix_uid", user.UnixUID); err != nil {
		return fmt.Errorf("error setting unix_uid: %v", err)
	}
	if err := d.Set("unix_guid", user.UnixGUID); err != nil {
		return fmt.Errorf("error setting unix_guid: %v", err)
	}
	if err := d.Set("disable_device_max_login_attempts", user.DisableDeviceMaxLoginAttempts); err != nil {
		return fmt.Errorf("error setting disable_device_max_login_attempts: %v", err)
	}
	if err := d.Set("password_recovery_email", user.PasswordRecoveryEmail); err != nil {
		return fmt.Errorf("error setting password_recovery_email: %v", err)
	}
	if err := d.Set("enforce_uid_gid_consistency", user.EnforceUIDGIDConsistency); err != nil {
		return fmt.Errorf("error setting enforce_uid_gid_consistency: %v", err)
	}
	if err := d.Set("delegated_authority", user.DelegatedAuthority); err != nil {
		return fmt.Errorf("error setting delegated_authority: %v", err)
	}
	if err := d.Set("password_authority", user.PasswordAuthority); err != nil {
		return fmt.Errorf("error setting password_authority: %v", err)
	}
	if err := d.Set("managed_apple_id", user.ManagedAppleID); err != nil {
		return fmt.Errorf("error setting managed_apple_id: %v", err)
	}
	if err := d.Set("bypass_managed_device_lockout", user.BypassManagedDeviceLockout); err != nil {
		return fmt.Errorf("error setting bypass_managed_device_lockout: %v", err)
	}
	if err := d.Set("created", user.Created); err != nil {
		return fmt.Errorf("error setting created: %v", err)
	}

	// Set manager ID if present
	if user.Manager != nil {
		if err := d.Set("manager_id", user.Manager.ID); err != nil {
			return fmt.Errorf("error setting manager_id: %v", err)
		}
	}

	// Convert attributes to map
	if len(user.Attributes) > 0 {
		attrs := make(map[string]string)
		for _, attr := range user.Attributes {
			attrs[attr.Name] = attr.Value
		}
		if err := d.Set("attributes", attrs); err != nil {
			return fmt.Errorf("error setting attributes: %v", err)
		}
	}

	return nil
}
