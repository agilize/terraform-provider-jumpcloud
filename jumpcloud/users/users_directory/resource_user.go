package users_directory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// UserAttribute represents a single user attribute
type UserAttribute struct {
	ID    string `json:"_id,omitempty"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// PhoneNumber represents a user's phone number
type PhoneNumber struct {
	ID     string `json:"_id,omitempty"`
	Type   string `json:"type"`
	Number string `json:"number"`
}

// Address represents a user's address
type Address struct {
	ID              string `json:"_id,omitempty"`
	Type            string `json:"type"`
	PoBox           string `json:"poBox,omitempty"`
	ExtendedAddress string `json:"extendedAddress,omitempty"`
	StreetAddress   string `json:"streetAddress,omitempty"`
	Locality        string `json:"locality,omitempty"`
	Region          string `json:"region,omitempty"`
	PostalCode      string `json:"postalCode,omitempty"`
	Country         string `json:"country,omitempty"`
}

// SSHKey represents a user's SSH key
type SSHKey struct {
	ID         string `json:"_id,omitempty"`
	Name       string `json:"name"`
	PublicKey  string `json:"public_key"`
	CreateDate string `json:"create_date,omitempty"`
}

// MFAConfig represents a user's MFA configuration
type MFAConfig struct {
	Exclusion      bool   `json:"exclusion,omitempty"`
	ExclusionUntil string `json:"exclusionUntil,omitempty"`
	Configured     bool   `json:"configured,omitempty"`
}

// MFAEnrollment represents a user's MFA enrollment status
type MFAEnrollment struct {
	TotpStatus     string `json:"totpStatus,omitempty"`
	WebAuthnStatus string `json:"webAuthnStatus,omitempty"`
	PushStatus     string `json:"pushStatus,omitempty"`
	OverallStatus  string `json:"overallStatus,omitempty"`
}

// Manager represents a user's manager
type Manager struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// SecurityKey represents a WebAuthn security key
type SecurityKey struct {
	ID         string `json:"_id,omitempty"`
	Name       string `json:"name"`
	CreateDate string `json:"create_date,omitempty"`
}

// Helper functions for data sanitization

// sanitizeAttributeName ensures attribute names only contain letters and numbers
// by replacing special characters with their alphanumeric equivalents
func sanitizeAttributeName(name string) string {
	// Replace hyphens, underscores, and spaces with empty string
	reg := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return reg.ReplaceAllString(name, "")
}

// ensureInt converts a value to int if it's not already
func ensureInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 0
}

// sanitizePhoneNumber removes non-numeric characters from phone numbers
func sanitizePhoneNumber(phone string) string {
	reg := regexp.MustCompile(`[^0-9+]`)
	return reg.ReplaceAllString(phone, "")
}

// formatManagerID ensures the manager ID is in the correct format
func formatManagerID(id string) string {
	// Remove any surrounding quotes or braces
	id = strings.Trim(id, "\"' {}")

	// If it contains a colon, extract just the ID part
	if strings.Contains(id, ":") {
		parts := strings.Split(id, ":")
		id = strings.TrimSpace(parts[len(parts)-1])
	}

	return id
}

// User represents a JumpCloud user
type User struct {
	ID                            string          `json:"_id,omitempty"`
	Username                      string          `json:"username"`
	Email                         string          `json:"email"`
	FirstName                     string          `json:"firstname,omitempty"`
	LastName                      string          `json:"lastname,omitempty"`
	MiddleName                    string          `json:"middlename,omitempty"`
	Password                      string          `json:"password,omitempty"`
	Description                   string          `json:"description,omitempty"`
	DisplayName                   string          `json:"displayname,omitempty"`
	Attributes                    []UserAttribute `json:"attributes,omitempty"`
	MFAEnabled                    bool            `json:"mfa_enabled,omitempty"`
	PasswordNeverExpires          bool            `json:"password_never_expires,omitempty"`
	Activated                     bool            `json:"activated,omitempty"`
	AccountLocked                 bool            `json:"account_locked,omitempty"`
	AccountLockedDate             string          `json:"account_locked_date,omitempty"`
	AlternateEmail                string          `json:"alternateEmail,omitempty"`
	Company                       string          `json:"company,omitempty"`
	CostCenter                    string          `json:"costCenter,omitempty"`
	Department                    string          `json:"department,omitempty"`
	EmployeeIdentifier            string          `json:"employeeIdentifier,omitempty"`
	EmployeeType                  string          `json:"employeeType,omitempty"`
	JobTitle                      string          `json:"jobTitle,omitempty"`
	Location                      string          `json:"location,omitempty"`
	ManagedAppleID                string          `json:"managedAppleId,omitempty"`
	EnableManagedUID              bool            `json:"enable_managed_uid,omitempty"`
	EnableUserPortalMultifactor   bool            `json:"enable_user_portal_multifactor,omitempty"`
	ExternalDN                    string          `json:"external_dn,omitempty"`
	ExternalSourceType            string          `json:"external_source_type,omitempty"`
	ExternallyManaged             bool            `json:"externally_managed,omitempty"`
	LDAPBindingUser               bool            `json:"ldap_binding_user,omitempty"`
	PasswordlessSudo              bool            `json:"passwordless_sudo,omitempty"`
	PublicKey                     string          `json:"public_key,omitempty"`
	SambaServiceUser              bool            `json:"samba_service_user,omitempty"`
	State                         string          `json:"state,omitempty"`
	Sudo                          bool            `json:"sudo,omitempty"`
	Suspended                     bool            `json:"suspended,omitempty"`
	SystemUsername                string          `json:"systemUsername,omitempty"`
	UnixGUID                      int             `json:"unix_guid,omitempty"`
	UnixUID                       int             `json:"unix_uid,omitempty"`
	DisableDeviceMaxLoginAttempts bool            `json:"disableDeviceMaxLoginAttempts,omitempty"`
	BypassManagedDeviceLockout    bool            `json:"bypassManagedDeviceLockout,omitempty"`
	AllowPublicKey                bool            `json:"allow_public_key,omitempty"`
	PasswordExpired               bool            `json:"password_expired,omitempty"`
	TOTPEnabled                   bool            `json:"totp_enabled,omitempty"`
	Addresses                     []Address       `json:"addresses,omitempty"`
	PhoneNumbers                  []PhoneNumber   `json:"phoneNumbers,omitempty"`
	SSHKeys                       []SSHKey        `json:"ssh_keys,omitempty"`
	SecurityKeys                  []SecurityKey   `json:"security_keys,omitempty"`
	MFA                           MFAConfig       `json:"mfa,omitempty"`
	MFAEnrollment                 MFAEnrollment   `json:"mfaEnrollment,omitempty"`
	Created                       string          `json:"created,omitempty"`
	PasswordDate                  string          `json:"password_date,omitempty"`
	PasswordExpirationDate        string          `json:"password_expiration_date,omitempty"`
	Manager                       *Manager        `json:"manager,omitempty"`
	PasswordRecoveryEmail         string          `json:"password_recovery_email,omitempty"`
	EnforceUIDGIDConsistency      bool            `json:"enforce_uid_gid_consistency,omitempty"`
	GlobalPasswordlessSudo        bool            `json:"global_passwordless_sudo,omitempty"`
	DelegatedAuthority            string          `json:"delegated_authority,omitempty"`
	PasswordAuthority             string          `json:"password_authority,omitempty"`
}

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"firstname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"lastname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"middlename": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"displayname": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Custom attributes for the user (key-value pairs)",
			},
			"mfa_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"password_never_expires": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"activated": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"account_locked": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"alternate_email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"company": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cost_center": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"department": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"employee_identifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"employee_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"job_title": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_managed_uid": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_user_portal_multifactor": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"externally_managed": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ldap_binding_user": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"passwordless_sudo": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"public_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"samba_service_user": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"sudo": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"suspended": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"unix_guid": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"unix_uid": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"disable_device_max_login_attempts": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allow_public_key": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"addresses": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"po_box": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"extended_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"street_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"locality": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"postal_code": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"country": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"phone_numbers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"number": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"ssh_keys": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"public_key": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"mfa": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclusion": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"exclusion_until": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"configured": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"totp_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"password_expired": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password_expiration_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"manager_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password_recovery_email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enforce_uid_gid_consistency": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"global_passwordless_sudo": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"delegated_authority": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password_authority": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"security_keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"managed_apple_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"bypass_managed_device_lockout": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build user object from resource data
	user := &User{
		Username:                      d.Get("username").(string),
		Email:                         d.Get("email").(string),
		FirstName:                     d.Get("firstname").(string),
		LastName:                      d.Get("lastname").(string),
		MiddleName:                    d.Get("middlename").(string),
		Password:                      d.Get("password").(string),
		Description:                   d.Get("description").(string),
		DisplayName:                   d.Get("displayname").(string),
		MFAEnabled:                    d.Get("mfa_enabled").(bool),
		PasswordNeverExpires:          d.Get("password_never_expires").(bool),
		AlternateEmail:                d.Get("alternate_email").(string),
		Company:                       d.Get("company").(string),
		CostCenter:                    d.Get("cost_center").(string),
		Department:                    d.Get("department").(string),
		EmployeeIdentifier:            d.Get("employee_identifier").(string),
		EmployeeType:                  d.Get("employee_type").(string),
		JobTitle:                      d.Get("job_title").(string),
		Location:                      d.Get("location").(string),
		EnableManagedUID:              d.Get("enable_managed_uid").(bool),
		EnableUserPortalMultifactor:   d.Get("enable_user_portal_multifactor").(bool),
		ExternallyManaged:             d.Get("externally_managed").(bool),
		LDAPBindingUser:               d.Get("ldap_binding_user").(bool),
		PasswordlessSudo:              d.Get("passwordless_sudo").(bool),
		PublicKey:                     d.Get("public_key").(string),
		SambaServiceUser:              d.Get("samba_service_user").(bool),
		Sudo:                          d.Get("sudo").(bool),
		Suspended:                     d.Get("suspended").(bool),
		DisableDeviceMaxLoginAttempts: d.Get("disable_device_max_login_attempts").(bool),
		AllowPublicKey:                d.Get("allow_public_key").(bool),
		PasswordRecoveryEmail:         d.Get("password_recovery_email").(string),
		EnforceUIDGIDConsistency:      d.Get("enforce_uid_gid_consistency").(bool),
		GlobalPasswordlessSudo:        d.Get("global_passwordless_sudo").(bool),
		DelegatedAuthority:            d.Get("delegated_authority").(string),
		PasswordAuthority:             d.Get("password_authority").(string),
		ManagedAppleID:                d.Get("managed_apple_id").(string),
		BypassManagedDeviceLockout:    d.Get("bypass_managed_device_lockout").(bool),
	}

	// Handle unix_uid and unix_guid to ensure they are integers
	if v, ok := d.GetOk("unix_uid"); ok {
		// Convert to int regardless of input type
		user.UnixUID = ensureInt(v)
	}

	if v, ok := d.GetOk("unix_guid"); ok {
		// Convert to int regardless of input type
		user.UnixGUID = ensureInt(v)
	}

	// Set manager if present
	if v, ok := d.GetOk("manager_id"); ok {
		managerId := v.(string)
		if managerId != "" {
			// Format the manager ID to ensure it's in the correct format
			formattedManagerId := formatManagerID(managerId)
			user.Manager = &Manager{
				ID: formattedManagerId,
			}
		}
	}

	// Set custom attributes if present
	if v, ok := d.GetOk("attributes"); ok {
		attrMap := v.(map[string]any)
		var attributes []UserAttribute
		for name, value := range attrMap {
			// Sanitize the attribute name for the API
			sanitizedName := sanitizeAttributeName(name)

			// Convert value to string if it's not already
			var strValue string
			switch v := value.(type) {
			case string:
				strValue = v
			case int, int64, float64:
				strValue = fmt.Sprintf("%v", v)
			case bool:
				strValue = fmt.Sprintf("%v", v)
			default:
				strValue = fmt.Sprintf("%v", v)
			}

			attributes = append(attributes, UserAttribute{
				Name:  sanitizedName, // Use sanitized name for API
				Value: strValue,
			})
		}
		user.Attributes = attributes
	}

	// Set addresses if present
	if v, ok := d.GetOk("addresses"); ok {
		addresses := v.([]interface{})
		userAddresses := make([]Address, 0, len(addresses))

		for _, addr := range addresses {
			addrMap := addr.(map[string]interface{})
			userAddresses = append(userAddresses, Address{
				Type:            addrMap["type"].(string),
				PoBox:           addrMap["po_box"].(string),
				ExtendedAddress: addrMap["extended_address"].(string),
				StreetAddress:   addrMap["street_address"].(string),
				Locality:        addrMap["locality"].(string),
				Region:          addrMap["region"].(string),
				PostalCode:      addrMap["postal_code"].(string),
				Country:         addrMap["country"].(string),
			})
		}

		user.Addresses = userAddresses
	}

	// Set phone numbers if present
	if v, ok := d.GetOk("phone_numbers"); ok {
		phones := v.([]interface{})
		userPhones := make([]PhoneNumber, 0, len(phones))

		for _, phone := range phones {
			phoneMap := phone.(map[string]interface{})
			// Use the original phone number format in the state
			userPhones = append(userPhones, PhoneNumber{
				Type:   phoneMap["type"].(string),
				Number: phoneMap["number"].(string),
			})
		}

		user.PhoneNumbers = userPhones
	}

	// Set SSH keys if present
	if v, ok := d.GetOk("ssh_keys"); ok {
		keys := v.([]interface{})
		userKeys := make([]SSHKey, 0, len(keys))

		for _, key := range keys {
			keyMap := key.(map[string]interface{})
			userKeys = append(userKeys, SSHKey{
				Name:      keyMap["name"].(string),
				PublicKey: keyMap["public_key"].(string),
			})
		}

		user.SSHKeys = userKeys
	}

	// Set MFA configuration if present
	if v, ok := d.GetOk("mfa"); ok {
		mfaList := v.([]interface{})
		if len(mfaList) > 0 {
			mfaMap := mfaList[0].(map[string]interface{})
			user.MFA = MFAConfig{
				Exclusion:      mfaMap["exclusion"].(bool),
				ExclusionUntil: mfaMap["exclusion_until"].(string),
			}
		}
	}

	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing user: %v", err))
	}

	// Create user via API
	// The correct endpoint for creating users is /systemusers
	resp, err := c.DoRequest(http.MethodPost, "/systemusers", userJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating user: %v", err))
	}

	// Decode response
	var newUser User
	if err := json.Unmarshal(resp, &newUser); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing user response: %v", err))
	}

	// Set ID in resource data
	d.SetId(newUser.ID)

	// Read the user to set all the computed fields
	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	userID := d.Id()

	// Get user via API
	// The correct endpoint for getting users is /systemusers/{id}
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/systemusers/%s", userID), nil)
	if err != nil {
		// Handle 404 specifically to mark the resource as removed
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("User %s not found, removing from state", userID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading user %s: %v", userID, err))
	}

	// Decode response
	var user User
	if err := json.Unmarshal(resp, &user); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing user response: %v", err))
	}

	// Set fields in resource data
	d.Set("username", user.Username)
	d.Set("email", user.Email)
	d.Set("firstname", user.FirstName)
	d.Set("lastname", user.LastName)
	d.Set("middlename", user.MiddleName)
	d.Set("description", user.Description)
	d.Set("displayname", user.DisplayName)
	// Set all fields, using the configuration values for boolean fields
	// This ensures we don't get into a loop with boolean fields

	// For critical boolean fields that are causing loops, use the configuration value
	// For mfa_enabled
	configMfaEnabled := d.Get("mfa_enabled").(bool)
	d.Set("mfa_enabled", configMfaEnabled)

	// For password_never_expires
	configPasswordNeverExpires := d.Get("password_never_expires").(bool)
	d.Set("password_never_expires", configPasswordNeverExpires)

	// For bypass_managed_device_lockout
	configBypassManagedDeviceLockout := d.Get("bypass_managed_device_lockout").(bool)
	d.Set("bypass_managed_device_lockout", configBypassManagedDeviceLockout)

	// Set other fields normally
	d.Set("activated", user.Activated)
	d.Set("account_locked", user.AccountLocked)
	d.Set("alternate_email", user.AlternateEmail)
	d.Set("company", user.Company)
	d.Set("cost_center", user.CostCenter)
	d.Set("department", user.Department)
	d.Set("employee_identifier", user.EmployeeIdentifier)
	d.Set("employee_type", user.EmployeeType)
	d.Set("job_title", user.JobTitle)
	d.Set("location", user.Location)

	// More boolean fields
	configEnableManagedUID := d.Get("enable_managed_uid").(bool)
	d.Set("enable_managed_uid", configEnableManagedUID)

	configEnableUserPortalMultifactor := d.Get("enable_user_portal_multifactor").(bool)
	d.Set("enable_user_portal_multifactor", configEnableUserPortalMultifactor)

	configExternallyManaged := d.Get("externally_managed").(bool)
	d.Set("externally_managed", configExternallyManaged)

	configLDAPBindingUser := d.Get("ldap_binding_user").(bool)
	d.Set("ldap_binding_user", configLDAPBindingUser)

	configPasswordlessSudo := d.Get("passwordless_sudo").(bool)
	d.Set("passwordless_sudo", configPasswordlessSudo)

	d.Set("public_key", user.PublicKey)

	configSambaServiceUser := d.Get("samba_service_user").(bool)
	d.Set("samba_service_user", configSambaServiceUser)

	configSudo := d.Get("sudo").(bool)
	d.Set("sudo", configSudo)

	configSuspended := d.Get("suspended").(bool)
	d.Set("suspended", configSuspended)

	d.Set("unix_guid", user.UnixGUID)
	d.Set("unix_uid", user.UnixUID)

	configDisableDeviceMaxLoginAttempts := d.Get("disable_device_max_login_attempts").(bool)
	d.Set("disable_device_max_login_attempts", configDisableDeviceMaxLoginAttempts)

	configAllowPublicKey := d.Get("allow_public_key").(bool)
	d.Set("allow_public_key", configAllowPublicKey)

	d.Set("password_expired", user.PasswordExpired)
	d.Set("totp_enabled", user.TOTPEnabled)
	d.Set("state", user.State)
	d.Set("created", user.Created)
	d.Set("password_date", user.PasswordDate)
	d.Set("password_expiration_date", user.PasswordExpirationDate)
	// password_recovery_email is handled below

	configEnforceUIDGIDConsistency := d.Get("enforce_uid_gid_consistency").(bool)
	d.Set("enforce_uid_gid_consistency", configEnforceUIDGIDConsistency)

	configGlobalPasswordlessSudo := d.Get("global_passwordless_sudo").(bool)
	d.Set("global_passwordless_sudo", configGlobalPasswordlessSudo)

	// Handle string fields that might be causing issues
	// For delegated_authority
	if v, ok := d.GetOk("delegated_authority"); ok {
		d.Set("delegated_authority", v.(string))
	} else {
		d.Set("delegated_authority", user.DelegatedAuthority)
	}

	// For password_authority
	if v, ok := d.GetOk("password_authority"); ok {
		d.Set("password_authority", v.(string))
	} else {
		d.Set("password_authority", user.PasswordAuthority)
	}

	// For password_recovery_email
	if v, ok := d.GetOk("password_recovery_email"); ok {
		d.Set("password_recovery_email", v.(string))
	} else {
		d.Set("password_recovery_email", user.PasswordRecoveryEmail)
	}

	d.Set("managed_apple_id", user.ManagedAppleID)

	// Set manager ID if present
	if user.Manager != nil {
		d.Set("manager_id", user.Manager.ID)
	}

	// Set custom attributes if present
	if len(user.Attributes) > 0 {
		// Get the original attributes from the configuration
		oldAttrs := d.Get("attributes").(map[string]interface{})

		// Create a map of sanitized name -> original name
		sanitizedToOriginal := make(map[string]string)
		for origName := range oldAttrs {
			sanitizedToOriginal[sanitizeAttributeName(origName)] = origName
		}

		// Create new attribute map preserving original names where possible
		attrMap := make(map[string]any)
		for _, attr := range user.Attributes {
			// Check if we have this attribute in the old configuration
			if origName, exists := sanitizedToOriginal[attr.Name]; exists {
				// Use the original name
				attrMap[origName] = attr.Value
			} else {
				// Use the name from the API
				attrMap[attr.Name] = attr.Value
			}
		}
		d.Set("attributes", attrMap)
	}

	// Set addresses if present
	if len(user.Addresses) > 0 {
		addresses := make([]map[string]interface{}, 0, len(user.Addresses))
		for _, addr := range user.Addresses {
			addrMap := map[string]interface{}{
				"type":             addr.Type,
				"po_box":           addr.PoBox,
				"extended_address": addr.ExtendedAddress,
				"street_address":   addr.StreetAddress,
				"locality":         addr.Locality,
				"region":           addr.Region,
				"postal_code":      addr.PostalCode,
				"country":          addr.Country,
			}
			addresses = append(addresses, addrMap)
		}
		d.Set("addresses", addresses)
	}

	// Set phone numbers if present
	if len(user.PhoneNumbers) > 0 {
		// Get the original phone numbers from the configuration
		oldPhones := d.Get("phone_numbers").([]interface{})
		oldPhoneMap := make(map[string]string)

		// Create a map of type -> number from the old configuration
		for _, oldPhone := range oldPhones {
			oldPhoneData := oldPhone.(map[string]interface{})
			oldPhoneMap[oldPhoneData["type"].(string)] = oldPhoneData["number"].(string)
		}

		// Create new phone list preserving original formatting where possible
		phones := make([]map[string]interface{}, 0, len(user.PhoneNumbers))
		for _, phone := range user.PhoneNumbers {
			// Check if we have this phone type in the old configuration
			originalNumber, exists := oldPhoneMap[phone.Type]

			phoneMap := map[string]interface{}{
				"type": phone.Type,
			}

			// Use the original formatted number if it exists and the digits match
			if exists && sanitizePhoneNumber(originalNumber) == sanitizePhoneNumber(phone.Number) {
				phoneMap["number"] = originalNumber
			} else {
				phoneMap["number"] = phone.Number
			}

			phones = append(phones, phoneMap)
		}
		d.Set("phone_numbers", phones)
	}

	// Set SSH keys if present
	if len(user.SSHKeys) > 0 {
		keys := make([]map[string]interface{}, 0, len(user.SSHKeys))
		for _, key := range user.SSHKeys {
			keyMap := map[string]interface{}{
				"name":       key.Name,
				"public_key": key.PublicKey,
			}
			keys = append(keys, keyMap)
		}
		d.Set("ssh_keys", keys)
	}

	// Set MFA configuration if present
	if user.MFA.Configured || user.MFA.Exclusion {
		mfaConfig := []map[string]interface{}{
			{
				"exclusion":       user.MFA.Exclusion,
				"exclusion_until": user.MFA.ExclusionUntil,
				"configured":      user.MFA.Configured,
			},
		}
		d.Set("mfa", mfaConfig)
	}

	// Set security keys if present
	if len(user.SecurityKeys) > 0 {
		keys := make([]map[string]interface{}, 0, len(user.SecurityKeys))
		for _, key := range user.SecurityKeys {
			keyMap := map[string]interface{}{
				"name": key.Name,
			}
			keys = append(keys, keyMap)
		}
		d.Set("security_keys", keys)
	} else {
		// Set an empty list to prevent "(known after apply)" in plans
		d.Set("security_keys", []map[string]interface{}{})
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	userID := d.Id()

	// Build user object from resource data
	user := &User{
		Username:                      d.Get("username").(string), // Include username in updates
		Email:                         d.Get("email").(string),
		FirstName:                     d.Get("firstname").(string),
		LastName:                      d.Get("lastname").(string),
		MiddleName:                    d.Get("middlename").(string),
		Description:                   d.Get("description").(string),
		DisplayName:                   d.Get("displayname").(string),
		MFAEnabled:                    d.Get("mfa_enabled").(bool),
		PasswordNeverExpires:          d.Get("password_never_expires").(bool),
		AlternateEmail:                d.Get("alternate_email").(string),
		Company:                       d.Get("company").(string),
		CostCenter:                    d.Get("cost_center").(string),
		Department:                    d.Get("department").(string),
		EmployeeIdentifier:            d.Get("employee_identifier").(string),
		EmployeeType:                  d.Get("employee_type").(string),
		JobTitle:                      d.Get("job_title").(string),
		Location:                      d.Get("location").(string),
		EnableManagedUID:              d.Get("enable_managed_uid").(bool),
		EnableUserPortalMultifactor:   d.Get("enable_user_portal_multifactor").(bool),
		ExternallyManaged:             d.Get("externally_managed").(bool),
		LDAPBindingUser:               d.Get("ldap_binding_user").(bool),
		PasswordlessSudo:              d.Get("passwordless_sudo").(bool),
		PublicKey:                     d.Get("public_key").(string),
		SambaServiceUser:              d.Get("samba_service_user").(bool),
		Sudo:                          d.Get("sudo").(bool),
		Suspended:                     d.Get("suspended").(bool),
		DisableDeviceMaxLoginAttempts: d.Get("disable_device_max_login_attempts").(bool),
		AllowPublicKey:                d.Get("allow_public_key").(bool),
		PasswordRecoveryEmail:         d.Get("password_recovery_email").(string),
		EnforceUIDGIDConsistency:      d.Get("enforce_uid_gid_consistency").(bool),
		GlobalPasswordlessSudo:        d.Get("global_passwordless_sudo").(bool),
		DelegatedAuthority:            d.Get("delegated_authority").(string),
		PasswordAuthority:             d.Get("password_authority").(string),
		ManagedAppleID:                d.Get("managed_apple_id").(string),
		BypassManagedDeviceLockout:    d.Get("bypass_managed_device_lockout").(bool),
	}

	// Handle unix_uid and unix_guid to ensure they are integers
	if v, ok := d.GetOk("unix_uid"); ok {
		// Convert to int regardless of input type
		user.UnixUID = ensureInt(v)
	}

	if v, ok := d.GetOk("unix_guid"); ok {
		// Convert to int regardless of input type
		user.UnixGUID = ensureInt(v)
	}

	// Set manager if present
	if v, ok := d.GetOk("manager_id"); ok {
		managerId := v.(string)
		if managerId != "" {
			// Format the manager ID to ensure it's in the correct format
			formattedManagerId := formatManagerID(managerId)
			user.Manager = &Manager{
				ID: formattedManagerId,
			}
		}
	}

	// Only set password if it's been changed
	if d.HasChange("password") {
		user.Password = d.Get("password").(string)
	}

	// Set custom attributes if present
	if v, ok := d.GetOk("attributes"); ok {
		attrMap := v.(map[string]any)
		var attributes []UserAttribute
		for name, value := range attrMap {
			// Sanitize the attribute name for the API
			sanitizedName := sanitizeAttributeName(name)

			// Convert value to string if it's not already
			var strValue string
			switch v := value.(type) {
			case string:
				strValue = v
			case int, int64, float64:
				strValue = fmt.Sprintf("%v", v)
			case bool:
				strValue = fmt.Sprintf("%v", v)
			default:
				strValue = fmt.Sprintf("%v", v)
			}

			attributes = append(attributes, UserAttribute{
				Name:  sanitizedName, // Use sanitized name for API
				Value: strValue,
			})
		}
		user.Attributes = attributes
	}

	// Set addresses if present
	if v, ok := d.GetOk("addresses"); ok {
		addresses := v.([]interface{})
		userAddresses := make([]Address, 0, len(addresses))

		for _, addr := range addresses {
			addrMap := addr.(map[string]interface{})
			userAddresses = append(userAddresses, Address{
				Type:            addrMap["type"].(string),
				PoBox:           addrMap["po_box"].(string),
				ExtendedAddress: addrMap["extended_address"].(string),
				StreetAddress:   addrMap["street_address"].(string),
				Locality:        addrMap["locality"].(string),
				Region:          addrMap["region"].(string),
				PostalCode:      addrMap["postal_code"].(string),
				Country:         addrMap["country"].(string),
			})
		}

		user.Addresses = userAddresses
	}

	// Set phone numbers if present
	if v, ok := d.GetOk("phone_numbers"); ok {
		phones := v.([]interface{})
		userPhones := make([]PhoneNumber, 0, len(phones))

		for _, phone := range phones {
			phoneMap := phone.(map[string]interface{})
			// Use the original phone number format in the state
			userPhones = append(userPhones, PhoneNumber{
				Type:   phoneMap["type"].(string),
				Number: phoneMap["number"].(string),
			})
		}

		user.PhoneNumbers = userPhones
	}

	// Set SSH keys if present
	if v, ok := d.GetOk("ssh_keys"); ok {
		keys := v.([]interface{})
		userKeys := make([]SSHKey, 0, len(keys))

		for _, key := range keys {
			keyMap := key.(map[string]interface{})
			userKeys = append(userKeys, SSHKey{
				Name:      keyMap["name"].(string),
				PublicKey: keyMap["public_key"].(string),
			})
		}

		user.SSHKeys = userKeys
	}

	// Set MFA configuration if present
	if v, ok := d.GetOk("mfa"); ok {
		mfaList := v.([]interface{})
		if len(mfaList) > 0 {
			mfaMap := mfaList[0].(map[string]interface{})
			user.MFA = MFAConfig{
				Exclusion:      mfaMap["exclusion"].(bool),
				ExclusionUntil: mfaMap["exclusion_until"].(string),
			}
		}
	}

	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing user: %v", err))
	}

	// Update user via API
	// The correct endpoint for updating users is /systemusers/{id}
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/systemusers/%s", userID), userJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating user %s: %v", userID, err))
	}

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	userID := d.Id()

	// Delete user via API
	// The correct endpoint for deleting users is /systemusers/{id}
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/systemusers/%s", userID), nil)
	if err != nil {
		// If the resource is already gone, don't return an error
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("User %s not found during delete, removing from state", userID))
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting user %s: %v", userID, err))
	}

	// Set ID to empty to signify resource has been removed
	d.SetId("")

	return nil
}
