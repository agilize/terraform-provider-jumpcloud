package users_directory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	ExclusionDays  int    `json:"exclusionDays,omitempty"`
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
	ID   string `json:"_id,omitempty"`
	Name string `json:"name,omitempty"`
}

// UnmarshalJSON implements the json.Unmarshaler interface for Manager
// This allows handling both string and object representations of the manager field
func (m *Manager) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string first (for data source responses)
	var managerID string
	if err := json.Unmarshal(data, &managerID); err == nil {
		m.ID = managerID
		return nil
	}

	// If that fails, try to unmarshal as an object (for resource responses)
	type ManagerAlias Manager
	var alias ManagerAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*m = Manager(alias)
	return nil
}

// SecurityKey represents a WebAuthn security key
type SecurityKey struct {
	ID         string `json:"_id,omitempty"`
	Name       string `json:"name"`
	CreateDate string `json:"create_date,omitempty"`
}

// RecoveryEmail represents a JumpCloud user's recovery email
type RecoveryEmail struct {
	Address string `json:"address,omitempty"`
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
func ensureInt(val any) int {
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

	// The JumpCloud API expects just the ID string, not an object
	return id
}

// formatStateForAPI converts the normalized state to the format expected by the API
func formatStateForAPI(state string) string {
	// Convert any case to uppercase format expected by JumpCloud API
	result := ""
	switch strings.ToUpper(state) {
	case "ACTIVATED":
		result = "ACTIVATED"
	case "STAGED":
		result = "STAGED"
	case "SUSPENDED":
		result = "SUSPENDED"
	default:
		// Return uppercase as fallback
		result = strings.ToUpper(state)
	}

	return result
}

// getFirstDefinedBool returns the first defined boolean value from the list of fields
// Useful for handling deprecated fields and their replacements
func getFirstDefinedBool(d *schema.ResourceData, fields []string) bool {
	for _, field := range fields {
		// Usar Get e verificar se o campo existe no estado
		if _, exists := d.GetOk(field); exists {
			return d.Get(field).(bool)
		}
	}
	return false
}

// formatAuthorityField formats the authority field to the format expected by the API
// If the value is "None" or empty, returns nil so it's omitted in JSON serialization
// For "ActiveDirectory", returns an object with the name
func formatAuthorityField(value string) any {
	if value == "" || value == "None" {
		return nil
	}

	if value == "ActiveDirectory" {
		return map[string]string{
			"name": value,
		}
	}

	// Para outros valores, retorna um objeto com o ID
	return map[string]string{
		"id": value,
	}
}

// formatPasswordAuthority formats the password_authority field to the format expected by the API
// If the value is "None" or empty, returns nil so it's omitted in JSON serialization
// For "Scim", configures the restrictedFields field with the password field
func formatPasswordAuthority(value string) any {
	if value == "" || value == "None" {
		return nil
	}

	if value == "Scim" {
		// For Scim, we don't send passwordAuthority, but restrictedFields instead
		// Return nil so the passwordAuthority field is omitted
		return nil
	}

	// For other values, use the standard format
	return map[string]string{
		"name": value,
	}
}

// User represents a JumpCloud user
type User struct {
	ID                          string          `json:"_id,omitempty"`
	Username                    string          `json:"username"`
	Email                       string          `json:"email"`
	FirstName                   string          `json:"firstname,omitempty"`
	LastName                    string          `json:"lastname,omitempty"`
	MiddleName                  string          `json:"middlename,omitempty"`
	Password                    string          `json:"password,omitempty"`
	Description                 string          `json:"description,omitempty"`
	DisplayName                 string          `json:"displayname,omitempty"`
	Attributes                  []UserAttribute `json:"attributes,omitempty"`
	MFAEnabled                  bool            `json:"mfa_enabled,omitempty"`
	PasswordNeverExpires        bool            `json:"password_never_expires,omitempty"`
	Activated                   bool            `json:"activated,omitempty"`
	AccountLocked               bool            `json:"account_locked,omitempty"`
	AccountLockedDate           string          `json:"account_locked_date,omitempty"`
	AlternateEmail              string          `json:"alternateEmail,omitempty"`
	Company                     string          `json:"company,omitempty"`
	CostCenter                  string          `json:"costCenter,omitempty"`
	Department                  string          `json:"department,omitempty"`
	EmployeeIdentifier          string          `json:"employeeIdentifier,omitempty"`
	EmployeeType                string          `json:"employeeType,omitempty"`
	JobTitle                    string          `json:"jobTitle,omitempty"`
	Location                    string          `json:"location,omitempty"`
	ManagedAppleID              string          `json:"managedAppleId,omitempty"`
	EnableManagedUID            bool            `json:"enable_managed_uid,omitempty"`
	EnableUserPortalMultifactor bool            `json:"enable_user_portal_multifactor,omitempty"`
	ExternalDN                  string          `json:"external_dn,omitempty"`
	ExternalSourceType          string          `json:"external_source_type,omitempty"`
	ExternallyManaged           bool            `json:"externally_managed,omitempty"`
	LDAPBindingUser             bool            `json:"ldap_binding_user,omitempty"`
	PasswordlessSudo            bool            `json:"passwordless_sudo,omitempty"`
	SambaServiceUser            bool            `json:"samba_service_user,omitempty"`
	State                       string          `json:"state,omitempty"`
	ActivationScheduled         bool            `json:"activation_scheduled,omitempty"`
	ScheduledActivationDate     string          `json:"scheduled_activation_date,omitempty"`
	Sudo                        bool            `json:"sudo,omitempty"`
	Suspended                   bool            `json:"suspended,omitempty"`
	SystemUsername              string          `json:"systemUsername,omitempty"`
	UnixGUID                    int             `json:"unix_guid,omitempty"`
	UnixUID                     int             `json:"unix_uid,omitempty"`
	// DisableDeviceMaxLoginAttempts is the real API name for BypassManagedDeviceLockout
	DisableDeviceMaxLoginAttempts bool `json:"disableDeviceMaxLoginAttempts,omitempty"`
	// BypassManagedDeviceLockout is the Terraform name, but not sent directly to the API
	BypassManagedDeviceLockout bool          `json:"-"`
	AllowPublicKey             bool          `json:"allow_public_key,omitempty"`
	PasswordExpired            bool          `json:"password_expired,omitempty"`
	TOTPEnabled                bool          `json:"totp_enabled,omitempty"`
	Addresses                  []Address     `json:"addresses,omitempty"`
	PhoneNumbers               []PhoneNumber `json:"phoneNumbers,omitempty"`
	SSHKeys                    []SSHKey      `json:"ssh_keys,omitempty"`
	SecurityKeys               []SecurityKey `json:"security_keys,omitempty"`
	MFA                        MFAConfig     `json:"mfa"`
	MFAEnrollment              MFAEnrollment `json:"mfaEnrollment"`
	Created                    string        `json:"created,omitempty"`
	PasswordDate               string        `json:"password_date,omitempty"`
	PasswordExpirationDate     string        `json:"password_expiration_date,omitempty"`
	Manager                    *Manager      `json:"manager,omitempty"`
	// PasswordRecoveryEmail is not used directly, we use RecoveryEmail instead
	PasswordRecoveryEmail string `json:"-"`
	// RecoveryEmail is the correct structure for recovery email
	RecoveryEmail *RecoveryEmail `json:"recoveryEmail,omitempty"`
	// EnforceUIDGIDConsistency maps to enable_managed_uid in the API, but not sent directly
	EnforceUIDGIDConsistency bool `json:"-"`
	// GlobalPasswordlessSudo maps to passwordless_sudo in the API, but not sent directly
	GlobalPasswordlessSudo bool `json:"-"`
	// These fields need to be treated as objects, not simple strings
	DelegatedAuthority any `json:"delegatedAuthority,omitempty"`
	PasswordAuthority  any `json:"passwordAuthority,omitempty"`
	// LocalUserAccount mapeia para systemUsername na API
	LocalUserAccount string `json:"-"`
}

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserImport,
		},
		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			// Validate allowed state transitions
			if diff.HasChange("state") {
				oldState, newState := diff.GetChange("state")
				oldStateUpper := strings.ToUpper(oldState.(string))
				newStateUpper := strings.ToUpper(newState.(string))

				// Skip validation if states are the same (no real change)
				if oldStateUpper == newStateUpper {
					return nil
				}

				// Define allowed transitions
				allowedTransitions := map[string][]string{
					"STAGED":    {"ACTIVATED", "SUSPENDED"},
					"ACTIVATED": {"SUSPENDED"},
					"SUSPENDED": {"ACTIVATED"},
				}

				// Check if transition is allowed
				if allowedStates, exists := allowedTransitions[oldStateUpper]; exists {
					transitionAllowed := false
					for _, allowedState := range allowedStates {
						if newStateUpper == allowedState {
							transitionAllowed = true
							break
						}
					}

					if !transitionAllowed {
						return fmt.Errorf("cannot change user state from %s to %s - allowed transitions from %s are: %v",
							oldStateUpper, newStateUpper, oldStateUpper, allowedStates)
					}
				} else if oldStateUpper != "" {
					// Unrecognized state or no transitions allowed
					return fmt.Errorf("cannot change user state from %s to %s - no transitions allowed from %s",
						oldStateUpper, newStateUpper, oldStateUpper)
				}
			}
			return nil
		},
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
			"require_mfa": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"mfa_enabled"},
				Description:   "Whether to require MFA for the user portal",
			},
			"mfa_enabled": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"require_mfa"},
				Deprecated:    "Use require_mfa instead",
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
				Type:       schema.TypeBool,
				Optional:   true,
				Default:    false,
				Deprecated: "This field is deprecated and will be removed in a future version.",
			},
			"enable_user_portal_multifactor": {
				Type:       schema.TypeBool,
				Optional:   true,
				Default:    false,
				Deprecated: "This field is deprecated and will be removed in a future version. Use mfa_enabled instead.",
			},
			"externally_managed": {
				Type:       schema.TypeBool,
				Optional:   true,
				Default:    false,
				Deprecated: "This field is deprecated and will be removed in a future version. Use password_authority instead.",
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
			"samba_service_user": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_global_admin_sudo": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"sudo"},
				Description:   "Enable as Global Administrator/Sudo on all device associations",
			},
			"sudo": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"enable_global_admin_sudo"},
				Deprecated:    "Use enable_global_admin_sudo instead",
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
				Type:       schema.TypeBool,
				Optional:   true,
				Default:    false,
				Deprecated: "This field is deprecated and will be removed in a future version. Use bypass_managed_device_lockout instead.",
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
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclusion": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether the user is excluded from MFA requirements",
						},
						"exclusion_days": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Number of days the user is excluded from MFA requirements (automatically converted based on user state)",
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(int)
								if v < 1 {
									errs = append(errs, fmt.Errorf("%q must be >= 1, got: %d", key, v))
								}
								return warns, errs
							},
						},
						"configured": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Whether MFA is configured for the user",
						},
					},
				},
				Description: "MFA configuration for the user (automatically handles STAGED vs ACTIVATED states)",
			},
			"state": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "User state (ACTIVATED, STAGED, SUSPENDED) - case insensitive",
				StateFunc: func(val interface{}) string {
					// Normalize to uppercase
					return strings.ToUpper(val.(string))
				},
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					// Validate allowed values (case insensitive)
					validStates := []string{"ACTIVATED", "STAGED", "SUSPENDED"}
					state := strings.ToUpper(val.(string))

					for _, validState := range validStates {
						if state == validState {
							return warns, errs
						}
					}

					errs = append(errs, fmt.Errorf("%q must be one of %v (case insensitive), got: %s", key, validStates, val.(string)))
					return warns, errs
				},
			},
			"activation_scheduled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether user activation is scheduled for a future date",
			},
			"scheduled_activation_date": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Date when user should be automatically activated (ISO 8601 format)",
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
			"local_user_account": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Local username for this user",
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
		Username:    d.Get("username").(string),
		Email:       d.Get("email").(string),
		FirstName:   d.Get("firstname").(string),
		LastName:    d.Get("lastname").(string),
		MiddleName:  d.Get("middlename").(string),
		Password:    d.Get("password").(string),
		Description: d.Get("description").(string),
		DisplayName: d.Get("displayname").(string),
		// Usar require_mfa se estiver definido, caso contrário usar mfa_enabled para compatibilidade
		MFAEnabled:           getFirstDefinedBool(d, []string{"require_mfa", "mfa_enabled"}),
		PasswordNeverExpires: d.Get("password_never_expires").(bool),
		AlternateEmail:       d.Get("alternate_email").(string),
		Company:              d.Get("company").(string),
		CostCenter:           d.Get("cost_center").(string),
		Department:           d.Get("department").(string),
		EmployeeIdentifier:   d.Get("employee_identifier").(string),
		EmployeeType:         d.Get("employee_type").(string),
		JobTitle:             d.Get("job_title").(string),
		Location:             d.Get("location").(string),
		// Mapeamento correto: enforce_uid_gid_consistency -> enable_managed_uid
		EnableManagedUID: d.Get("enforce_uid_gid_consistency").(bool),
		// Mapeamento correto: require_mfa -> enable_user_portal_multifactor
		EnableUserPortalMultifactor: getFirstDefinedBool(d, []string{"require_mfa", "mfa_enabled"}),
		ExternallyManaged:           d.Get("externally_managed").(bool),
		LDAPBindingUser:             d.Get("ldap_binding_user").(bool),
		// Mapeamento correto: global_passwordless_sudo -> passwordless_sudo
		PasswordlessSudo: d.Get("global_passwordless_sudo").(bool),
		SambaServiceUser: d.Get("samba_service_user").(bool),
		// Usar enable_global_admin_sudo se estiver definido, caso contrário usar sudo para compatibilidade
		Sudo:      getFirstDefinedBool(d, []string{"enable_global_admin_sudo", "sudo"}),
		Suspended: d.Get("suspended").(bool),
		// Campos para controle de estado e agendamento de ativação
		State:                   formatStateForAPI(d.Get("state").(string)),
		ActivationScheduled:     d.Get("activation_scheduled").(bool),
		ScheduledActivationDate: d.Get("scheduled_activation_date").(string),
		// Mapeamento correto: bypass_managed_device_lockout -> disableDeviceMaxLoginAttempts
		DisableDeviceMaxLoginAttempts: d.Get("bypass_managed_device_lockout").(bool),
		AllowPublicKey:                d.Get("allow_public_key").(bool),
		PasswordRecoveryEmail:         d.Get("password_recovery_email").(string),
		// Estes campos são mapeados para outros campos na API
		EnforceUIDGIDConsistency: d.Get("enforce_uid_gid_consistency").(bool),
		GlobalPasswordlessSudo:   d.Get("global_passwordless_sudo").(bool),
		// Usar as funções específicas para formatar corretamente os campos de autoridade
		DelegatedAuthority:         formatAuthorityField(d.Get("delegated_authority").(string)),
		PasswordAuthority:          formatPasswordAuthority(d.Get("password_authority").(string)),
		ManagedAppleID:             d.Get("managed_apple_id").(string),
		BypassManagedDeviceLockout: d.Get("bypass_managed_device_lockout").(bool),
		// Mapeamento correto: local_user_account -> systemUsername
		SystemUsername: d.Get("local_user_account").(string),
		// Mapeamento correto: password_recovery_email -> recoveryEmail
		RecoveryEmail: &RecoveryEmail{
			Address: d.Get("password_recovery_email").(string),
		},
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
			// The JumpCloud API expects the manager ID directly, not wrapped in an object
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
		addresses := v.([]any)
		userAddresses := make([]Address, 0, len(addresses))

		for _, addr := range addresses {
			addrMap := addr.(map[string]any)
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
		phones := v.([]any)
		userPhones := make([]PhoneNumber, 0, len(phones))

		for _, phone := range phones {
			phoneMap := phone.(map[string]any)
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
		keys := v.([]any)
		userKeys := make([]SSHKey, 0, len(keys))

		for _, key := range keys {
			keyMap := key.(map[string]any)
			userKeys = append(userKeys, SSHKey{
				Name:      keyMap["name"].(string),
				PublicKey: keyMap["public_key"].(string),
			})
		}

		user.SSHKeys = userKeys
	}

	// Set MFA configuration based on user state
	userState := d.Get("state").(string)
	if mfaConfig := buildMFAConfig(d, userState); mfaConfig != nil {
		user.MFA = *mfaConfig

		// Log da conversão para debug
		if v, ok := d.GetOk("mfa"); ok {
			mfaList := v.([]any)
			if len(mfaList) > 0 {
				mfaMap := mfaList[0].(map[string]any)
				if days, ok := mfaMap["exclusion_days"]; ok {
					logMFAConversion(ctx, userState, days.(int), mfaConfig)
				}
			}
		}
	}

	// Ensure these fields are explicitly set, even if they're empty
	// This is necessary because the JumpCloud API requires these fields to be present
	if user.LocalUserAccount == "" {
		user.LocalUserAccount = "" // Explicitly set to empty string
	}

	if user.PasswordRecoveryEmail == "" {
		user.PasswordRecoveryEmail = "" // Explicitly set to empty string
	}

	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing user: %v", err))
	}

	// Create user via API
	// Use the constant for the system users path
	tflog.Debug(ctx, fmt.Sprintf("Creating user with URL: %s", common.SystemUsersPath))
	tflog.Debug(ctx, fmt.Sprintf("Request body: %s", string(userJSON)))

	// Try with direct API path
	resp, err := c.DoRequest(http.MethodPost, "/api/systemusers", userJSON)
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

	// Store the values of problematic fields to ensure they're preserved in the state
	passwordRecoveryEmail := d.Get("password_recovery_email").(string)
	localUserAccount := d.Get("local_user_account").(string)
	passwordNeverExpires := d.Get("password_never_expires").(bool)
	bypassManagedDeviceLockout := d.Get("bypass_managed_device_lockout").(bool)

	// Always make a separate API call for problematic fields
	// These fields need special handling to ensure they're properly set
	{
		// Create a special update object with all the problematic fields
		// Make sure to use the exact field names expected by the JumpCloud API
		specialUpdate := map[string]any{
			// Mapeamento correto: password_recovery_email -> recoveryEmail
			"recoveryEmail": map[string]string{
				"address": passwordRecoveryEmail,
			},
			// Mapeamento correto: local_user_account -> systemUsername
			"systemUsername": localUserAccount,
			// Tentar vários formatos de nome para os mesmos campos
			// Nomes corretos da API
			"password_never_expires":        passwordNeverExpires,
			"disableDeviceMaxLoginAttempts": bypassManagedDeviceLockout,

			// Nomes corretos da API
			"enable_user_portal_multifactor": getFirstDefinedBool(d, []string{"require_mfa", "mfa_enabled"}),
			"sudo":                           getFirstDefinedBool(d, []string{"enable_global_admin_sudo", "sudo"}),
			"passwordless_sudo":              d.Get("global_passwordless_sudo").(bool),
			"ldap_binding_user":              d.Get("ldap_binding_user").(bool),
			"enable_managed_uid":             d.Get("enforce_uid_gid_consistency").(bool),

			// Adicionando campos de autoridade que estavam faltando
			// Tratar como objetos ou nulos, não como strings simples
			"delegatedAuthority": formatAuthorityField(d.Get("delegated_authority").(string)),

			// Tratamento especial para password_authority
			"passwordAuthority": formatPasswordAuthority(d.Get("password_authority").(string)),

			// Adicionar restrictedFields quando password_authority for "Scim"
			"restrictedFields": func() any {
				if d.Get("password_authority").(string) == "Scim" {
					return []map[string]any{
						{
							"field": "password",
							"type":  "scim",
							"id":    nil,
						},
					}
				}
				return nil
			}(),
		}

		// Convert to JSON
		specialUpdateJSON, err := json.Marshal(specialUpdate)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error serializing special update: %v", err))
		}

		// Make a PUT request to update just these fields
		// JumpCloud API doesn't support PATCH, so we need to use PUT
		tflog.Debug(ctx, fmt.Sprintf("Making special update for problematic fields for user ID: %s", newUser.ID))
		tflog.Debug(ctx, fmt.Sprintf("Special update request body: %s", string(specialUpdateJSON)))
		_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/systemusers/%s", newUser.ID), specialUpdateJSON)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error making special update for user %s: %v", newUser.ID, err))
		}
	}

	// Preserve os valores dos campos de recuperação de email e conta local
	// que não são retornados corretamente pela API
	if err := d.Set("password_recovery_email", passwordRecoveryEmail); err != nil {
		return diag.FromErr(fmt.Errorf("error setting password_recovery_email: %v", err))
	}

	if err := d.Set("local_user_account", localUserAccount); err != nil {
		return diag.FromErr(fmt.Errorf("error setting local_user_account: %v", err))
	}

	// Preserve os valores dos campos de autoridade
	if err := d.Set("password_authority", d.Get("password_authority").(string)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting password_authority: %v", err))
	}

	if err := d.Set("delegated_authority", d.Get("delegated_authority").(string)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting delegated_authority: %v", err))
	}

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
	// Use the same direct API path as in create
	tflog.Debug(ctx, fmt.Sprintf("Reading user with ID: %s", userID))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/systemusers/%s", userID), nil)
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
	if err := d.Set("username", user.Username); err != nil {
		return diag.FromErr(fmt.Errorf("error setting username: %v", err))
	}

	if err := d.Set("email", user.Email); err != nil {
		return diag.FromErr(fmt.Errorf("error setting email: %v", err))
	}

	if err := d.Set("firstname", user.FirstName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting firstname: %v", err))
	}

	if err := d.Set("lastname", user.LastName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting lastname: %v", err))
	}

	if err := d.Set("middlename", user.MiddleName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting middlename: %v", err))
	}

	if err := d.Set("description", user.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %v", err))
	}

	if err := d.Set("displayname", user.DisplayName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting displayname: %v", err))
	}

	// Set all fields, using the configuration values for boolean fields
	// This ensures we don't get into a loop with boolean fields

	// For all boolean fields that might cause loops, use the configuration value
	// This ensures we don't get into a loop with boolean fields

	// User Information
	if err := d.Set("activated", user.Activated); err != nil {
		return diag.FromErr(fmt.Errorf("error setting activated: %v", err))
	}

	if err := d.Set("account_locked", user.AccountLocked); err != nil {
		return diag.FromErr(fmt.Errorf("error setting account_locked: %v", err))
	}

	if err := d.Set("alternate_email", user.AlternateEmail); err != nil {
		return diag.FromErr(fmt.Errorf("error setting alternate_email: %v", err))
	}

	if err := d.Set("company", user.Company); err != nil {
		return diag.FromErr(fmt.Errorf("error setting company: %v", err))
	}

	if err := d.Set("cost_center", user.CostCenter); err != nil {
		return diag.FromErr(fmt.Errorf("error setting cost_center: %v", err))
	}

	if err := d.Set("department", user.Department); err != nil {
		return diag.FromErr(fmt.Errorf("error setting department: %v", err))
	}

	if err := d.Set("employee_identifier", user.EmployeeIdentifier); err != nil {
		return diag.FromErr(fmt.Errorf("error setting employee_identifier: %v", err))
	}

	if err := d.Set("employee_type", user.EmployeeType); err != nil {
		return diag.FromErr(fmt.Errorf("error setting employee_type: %v", err))
	}

	if err := d.Set("job_title", user.JobTitle); err != nil {
		return diag.FromErr(fmt.Errorf("error setting job_title: %v", err))
	}

	if err := d.Set("location", user.Location); err != nil {
		return diag.FromErr(fmt.Errorf("error setting location: %v", err))
	}

	if err := d.Set("unix_guid", user.UnixGUID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting unix_guid: %v", err))
	}

	if err := d.Set("unix_uid", user.UnixUID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting unix_uid: %v", err))
	}

	if err := d.Set("password_expired", user.PasswordExpired); err != nil {
		return diag.FromErr(fmt.Errorf("error setting password_expired: %v", err))
	}

	if err := d.Set("totp_enabled", user.TOTPEnabled); err != nil {
		return diag.FromErr(fmt.Errorf("error setting totp_enabled: %v", err))
	}

	if err := d.Set("state", user.State); err != nil {
		return diag.FromErr(fmt.Errorf("error setting state: %v", err))
	}

	if err := d.Set("activation_scheduled", user.ActivationScheduled); err != nil {
		return diag.FromErr(fmt.Errorf("error setting activation_scheduled: %v", err))
	}

	if err := d.Set("scheduled_activation_date", user.ScheduledActivationDate); err != nil {
		return diag.FromErr(fmt.Errorf("error setting scheduled_activation_date: %v", err))
	}

	if err := d.Set("created", user.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %v", err))
	}

	if err := d.Set("password_date", user.PasswordDate); err != nil {
		return diag.FromErr(fmt.Errorf("error setting password_date: %v", err))
	}

	if err := d.Set("password_expiration_date", user.PasswordExpirationDate); err != nil {
		return diag.FromErr(fmt.Errorf("error setting password_expiration_date: %v", err))
	}

	// Para campos que precisam ser lidos da API, mas que podem causar problemas de persistência
	// Vamos usar uma abordagem mais simples: preservar os valores do estado atual

	// Obter os valores atuais do estado
	currentState := map[string]bool{
		"password_never_expires":            d.Get("password_never_expires").(bool),
		"require_mfa":                       d.Get("require_mfa").(bool),
		"mfa_enabled":                       d.Get("mfa_enabled").(bool),
		"enable_global_admin_sudo":          d.Get("enable_global_admin_sudo").(bool),
		"sudo":                              d.Get("sudo").(bool),
		"global_passwordless_sudo":          d.Get("global_passwordless_sudo").(bool),
		"passwordless_sudo":                 d.Get("passwordless_sudo").(bool),
		"ldap_binding_user":                 d.Get("ldap_binding_user").(bool),
		"enforce_uid_gid_consistency":       d.Get("enforce_uid_gid_consistency").(bool),
		"enable_managed_uid":                d.Get("enable_managed_uid").(bool),
		"bypass_managed_device_lockout":     d.Get("bypass_managed_device_lockout").(bool),
		"disable_device_max_login_attempts": d.Get("disable_device_max_login_attempts").(bool),
		"externally_managed":                d.Get("externally_managed").(bool),
		"samba_service_user":                d.Get("samba_service_user").(bool),
		"suspended":                         d.Get("suspended").(bool),
		"allow_public_key":                  d.Get("allow_public_key").(bool),
	}

	// Valores da API
	apiState := map[string]bool{
		"password_never_expires":            user.PasswordNeverExpires,
		"require_mfa":                       user.EnableUserPortalMultifactor,
		"mfa_enabled":                       user.EnableUserPortalMultifactor,
		"enable_global_admin_sudo":          user.Sudo,
		"sudo":                              user.Sudo,
		"global_passwordless_sudo":          user.PasswordlessSudo,
		"passwordless_sudo":                 user.PasswordlessSudo,
		"ldap_binding_user":                 user.LDAPBindingUser,
		"enforce_uid_gid_consistency":       user.EnableManagedUID,
		"enable_managed_uid":                user.EnableManagedUID,
		"bypass_managed_device_lockout":     user.DisableDeviceMaxLoginAttempts,
		"disable_device_max_login_attempts": user.DisableDeviceMaxLoginAttempts,
		"externally_managed":                user.ExternallyManaged,
		"samba_service_user":                user.SambaServiceUser,
		"suspended":                         user.Suspended,
		"allow_public_key":                  user.AllowPublicKey,
	}

	// Definir os campos no estado
	// Se o valor no estado atual for diferente do valor na API, usar o valor do estado atual
	// Caso contrário, usar o valor da API
	for field, apiValue := range apiState {
		currentValue, exists := currentState[field]
		if exists && currentValue != apiValue {
			// Se o valor no estado atual for diferente do valor na API, usar o valor do estado atual
			if err := d.Set(field, currentValue); err != nil {
				return diag.FromErr(fmt.Errorf("error setting %s: %v", field, err))
			}
		} else {
			// Caso contrário, usar o valor da API
			if err := d.Set(field, apiValue); err != nil {
				return diag.FromErr(fmt.Errorf("error setting %s: %v", field, err))
			}
		}
	}

	// For password_recovery_email and local_user_account, only set them if they're not already set
	// This preserves the values we set in the update function
	if _, ok := d.GetOk("password_recovery_email"); !ok {
		if err := d.Set("password_recovery_email", user.PasswordRecoveryEmail); err != nil {
			return diag.FromErr(fmt.Errorf("error setting password_recovery_email: %v", err))
		}
	}

	if _, ok := d.GetOk("local_user_account"); !ok {
		if err := d.Set("local_user_account", user.LocalUserAccount); err != nil {
			return diag.FromErr(fmt.Errorf("error setting local_user_account: %v", err))
		}
	}

	// Handle string fields that might be causing issues
	// For delegated_authority
	if v, ok := d.GetOk("delegated_authority"); ok {
		if err := d.Set("delegated_authority", v.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting delegated_authority: %v", err))
		}
	} else {
		if err := d.Set("delegated_authority", user.DelegatedAuthority); err != nil {
			return diag.FromErr(fmt.Errorf("error setting delegated_authority: %v", err))
		}
	}

	// For password_authority
	if v, ok := d.GetOk("password_authority"); ok {
		if err := d.Set("password_authority", v.(string)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting password_authority: %v", err))
		}
	} else {
		if err := d.Set("password_authority", user.PasswordAuthority); err != nil {
			return diag.FromErr(fmt.Errorf("error setting password_authority: %v", err))
		}
	}

	// Set managed_apple_id from the API response
	if err := d.Set("managed_apple_id", user.ManagedAppleID); err != nil {
		return diag.FromErr(fmt.Errorf("error setting managed_apple_id: %v", err))
	}

	// Set manager ID if present
	if user.Manager != nil {
		if err := d.Set("manager_id", user.Manager.ID); err != nil {
			return diag.FromErr(fmt.Errorf("error setting manager_id: %v", err))
		}
	}

	// Set custom attributes if present
	if len(user.Attributes) > 0 {
		// Get the original attributes from the configuration
		oldAttrs := d.Get("attributes").(map[string]any)

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
		if err := d.Set("attributes", attrMap); err != nil {
			return diag.FromErr(fmt.Errorf("error setting attributes: %v", err))
		}
	}

	// Set addresses if present
	if len(user.Addresses) > 0 {
		addresses := make([]map[string]any, 0, len(user.Addresses))
		for _, addr := range user.Addresses {
			addrMap := map[string]any{
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
		if err := d.Set("addresses", addresses); err != nil {
			return diag.FromErr(fmt.Errorf("error setting addresses: %v", err))
		}
	}

	// Set phone numbers if present
	if len(user.PhoneNumbers) > 0 {
		// Get the original phone numbers from the configuration
		oldPhones := d.Get("phone_numbers").([]any)
		oldPhoneMap := make(map[string]string)

		// Create a map of type -> number from the old configuration
		for _, oldPhone := range oldPhones {
			oldPhoneData := oldPhone.(map[string]any)
			oldPhoneMap[oldPhoneData["type"].(string)] = oldPhoneData["number"].(string)
		}

		// Create new phone list preserving original formatting where possible
		phones := make([]map[string]any, 0, len(user.PhoneNumbers))
		for _, phone := range user.PhoneNumbers {
			// Check if we have this phone type in the old configuration
			originalNumber, exists := oldPhoneMap[phone.Type]

			phoneMap := map[string]any{
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
		if err := d.Set("phone_numbers", phones); err != nil {
			return diag.FromErr(fmt.Errorf("error setting phone_numbers: %v", err))
		}
	}

	// Set SSH keys if present
	if len(user.SSHKeys) > 0 {
		keys := make([]map[string]any, 0, len(user.SSHKeys))
		for _, key := range user.SSHKeys {
			keyMap := map[string]any{
				"name":       key.Name,
				"public_key": key.PublicKey,
			}
			keys = append(keys, keyMap)
		}
		if err := d.Set("ssh_keys", keys); err != nil {
			return diag.FromErr(fmt.Errorf("error setting ssh_keys: %v", err))
		}
	}

	// Set MFA configuration if present
	if user.MFA.Configured || user.MFA.Exclusion {
		exclusionDaysToSet := 0

		// Get current state to determine how to handle MFA data
		userState := strings.ToUpper(user.State)

		// Calculate exclusion_days based on what the API returned and user state
		if user.MFA.ExclusionDays > 0 {
			// API returned exclusionDays directly (STAGED users)
			exclusionDaysToSet = user.MFA.ExclusionDays
		} else if user.MFA.ExclusionUntil != "" && userState == "ACTIVATED" {
			// API returned exclusionUntil (ACTIVATED users) - calculate days from now
			if exclusionTime, err := time.Parse(time.RFC3339, user.MFA.ExclusionUntil); err == nil {
				now := time.Now()
				// Calculate days difference using date comparison (not hours)
				// This ensures we get the correct number of days regardless of time of day
				nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				exclusionDate := time.Date(exclusionTime.Year(), exclusionTime.Month(), exclusionTime.Day(), 0, 0, 0, 0, exclusionTime.Location())

				daysFromNow := int(exclusionDate.Sub(nowDate).Hours() / 24)
				// Add 1 day to match JumpCloud's calculation method
				if daysFromNow >= 0 {
					exclusionDaysToSet = daysFromNow + 1
				}
			}
		} else {
			// No exclusion data from API, preserve current state value if exists
			if v, ok := d.GetOk("mfa"); ok {
				mfaList := v.([]any)
				if len(mfaList) > 0 {
					mfaMap := mfaList[0].(map[string]any)
					if days, exists := mfaMap["exclusion_days"]; exists {
						exclusionDaysToSet = days.(int)
					}
				}
			}
		}

		mfaConfig := []map[string]any{
			{
				"exclusion":      user.MFA.Exclusion,
				"exclusion_days": exclusionDaysToSet,
				"configured":     user.MFA.Configured,
			},
		}
		if err := d.Set("mfa", mfaConfig); err != nil {
			return diag.FromErr(fmt.Errorf("error setting mfa: %v", err))
		}
	}

	// Set security keys if present
	if len(user.SecurityKeys) > 0 {
		keys := make([]map[string]any, 0, len(user.SecurityKeys))
		for _, key := range user.SecurityKeys {
			keyMap := map[string]any{
				"name": key.Name,
			}
			keys = append(keys, keyMap)
		}
		if err := d.Set("security_keys", keys); err != nil {
			return diag.FromErr(fmt.Errorf("error setting security_keys: %v", err))
		}
	} else {
		// Set an empty list to prevent "(known after apply)" in plans
		if err := d.Set("security_keys", []map[string]any{}); err != nil {
			return diag.FromErr(fmt.Errorf("error setting empty security_keys: %v", err))
		}
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	userID := d.Id()

	// We always make a special update for problematic fields now
	// But track changes for logging purposes
	hasProblematicFieldChanges := d.HasChange("password_recovery_email") ||
		d.HasChange("local_user_account") ||
		d.HasChange("password_never_expires") ||
		d.HasChange("bypass_managed_device_lockout") ||
		d.HasChange("enable_user_portal_multifactor") ||
		d.HasChange("mfa_enabled") ||
		d.HasChange("require_mfa") ||
		d.HasChange("sudo") ||
		d.HasChange("enable_global_admin_sudo") ||
		d.HasChange("passwordless_sudo") ||
		d.HasChange("global_passwordless_sudo") ||
		d.HasChange("ldap_binding_user") ||
		d.HasChange("enforce_uid_gid_consistency") ||
		d.HasChange("password_authority") ||
		d.HasChange("delegated_authority")

	if hasProblematicFieldChanges {
		tflog.Debug(ctx, "Detected changes to problematic fields that require special handling")
	}

	// Store the values of problematic fields
	passwordRecoveryEmail := d.Get("password_recovery_email").(string)
	localUserAccount := d.Get("local_user_account").(string)
	passwordNeverExpires := d.Get("password_never_expires").(bool)
	bypassManagedDeviceLockout := d.Get("bypass_managed_device_lockout").(bool)

	// Build user object from resource data
	user := &User{
		Username:    d.Get("username").(string), // Include username in updates
		Email:       d.Get("email").(string),
		FirstName:   d.Get("firstname").(string),
		LastName:    d.Get("lastname").(string),
		MiddleName:  d.Get("middlename").(string),
		Description: d.Get("description").(string),
		DisplayName: d.Get("displayname").(string),
		// Usar require_mfa se estiver definido, caso contrário usar mfa_enabled para compatibilidade
		MFAEnabled:           getFirstDefinedBool(d, []string{"require_mfa", "mfa_enabled"}),
		PasswordNeverExpires: passwordNeverExpires,
		AlternateEmail:       d.Get("alternate_email").(string),
		Company:              d.Get("company").(string),
		CostCenter:           d.Get("cost_center").(string),
		Department:           d.Get("department").(string),
		EmployeeIdentifier:   d.Get("employee_identifier").(string),
		EmployeeType:         d.Get("employee_type").(string),
		JobTitle:             d.Get("job_title").(string),
		Location:             d.Get("location").(string),
		// Mapeamento correto: enforce_uid_gid_consistency -> enable_managed_uid
		EnableManagedUID: d.Get("enforce_uid_gid_consistency").(bool),
		// Mapeamento correto: require_mfa -> enable_user_portal_multifactor
		EnableUserPortalMultifactor: getFirstDefinedBool(d, []string{"require_mfa", "mfa_enabled"}),
		ExternallyManaged:           d.Get("externally_managed").(bool),
		LDAPBindingUser:             d.Get("ldap_binding_user").(bool),
		// Mapeamento correto: global_passwordless_sudo -> passwordless_sudo
		PasswordlessSudo: d.Get("global_passwordless_sudo").(bool),
		SambaServiceUser: d.Get("samba_service_user").(bool),
		// Usar enable_global_admin_sudo se estiver definido, caso contrário usar sudo para compatibilidade
		Sudo:      getFirstDefinedBool(d, []string{"enable_global_admin_sudo", "sudo"}),
		Suspended: d.Get("suspended").(bool),
		// Campos para controle de estado e agendamento de ativação
		State:                   formatStateForAPI(d.Get("state").(string)),
		ActivationScheduled:     d.Get("activation_scheduled").(bool),
		ScheduledActivationDate: d.Get("scheduled_activation_date").(string),
		// Mapeamento correto: bypass_managed_device_lockout -> disableDeviceMaxLoginAttempts
		DisableDeviceMaxLoginAttempts: d.Get("bypass_managed_device_lockout").(bool),
		AllowPublicKey:                d.Get("allow_public_key").(bool),
		PasswordRecoveryEmail:         passwordRecoveryEmail,
		// Estes campos são mapeados para outros campos na API
		EnforceUIDGIDConsistency: d.Get("enforce_uid_gid_consistency").(bool),
		GlobalPasswordlessSudo:   d.Get("global_passwordless_sudo").(bool),
		// Usar as funções específicas para formatar corretamente os campos de autoridade
		DelegatedAuthority:         formatAuthorityField(d.Get("delegated_authority").(string)),
		PasswordAuthority:          formatPasswordAuthority(d.Get("password_authority").(string)),
		ManagedAppleID:             d.Get("managed_apple_id").(string),
		BypassManagedDeviceLockout: bypassManagedDeviceLockout,
		// Mapeamento correto: local_user_account -> systemUsername
		SystemUsername: localUserAccount,
		// Mapeamento correto: password_recovery_email -> recoveryEmail
		RecoveryEmail: &RecoveryEmail{
			Address: passwordRecoveryEmail,
		},
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
			// The JumpCloud API expects the manager ID directly, not wrapped in an object
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
		addresses := v.([]any)
		userAddresses := make([]Address, 0, len(addresses))

		for _, addr := range addresses {
			addrMap := addr.(map[string]any)
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
		phones := v.([]any)
		userPhones := make([]PhoneNumber, 0, len(phones))

		for _, phone := range phones {
			phoneMap := phone.(map[string]any)
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
		keys := v.([]any)
		userKeys := make([]SSHKey, 0, len(keys))

		for _, key := range keys {
			keyMap := key.(map[string]any)
			userKeys = append(userKeys, SSHKey{
				Name:      keyMap["name"].(string),
				PublicKey: keyMap["public_key"].(string),
			})
		}

		user.SSHKeys = userKeys
	}

	// Set MFA configuration based on user state
	userState := d.Get("state").(string)
	if mfaConfig := buildMFAConfig(d, userState); mfaConfig != nil {
		user.MFA = *mfaConfig

		// Log da conversão para debug
		if v, ok := d.GetOk("mfa"); ok {
			mfaList := v.([]any)
			if len(mfaList) > 0 {
				mfaMap := mfaList[0].(map[string]any)
				if days, ok := mfaMap["exclusion_days"]; ok {
					logMFAConversion(ctx, userState, days.(int), mfaConfig)
				}
			}
		}
	}

	// Ensure these fields are explicitly set, even if they're empty
	// This is necessary because the JumpCloud API requires these fields to be present
	if user.LocalUserAccount == "" {
		user.LocalUserAccount = "" // Explicitly set to empty string
	}

	if user.PasswordRecoveryEmail == "" {
		user.PasswordRecoveryEmail = "" // Explicitly set to empty string
	}

	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing user: %v", err))
	}

	// Update user via API
	// Use the same direct API path as in create and read
	tflog.Debug(ctx, fmt.Sprintf("Updating user with ID: %s", userID))
	tflog.Debug(ctx, fmt.Sprintf("Request body: %s", string(userJSON)))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/systemusers/%s", userID), userJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating user %s: %v", userID, err))
	}

	// Always make a separate API call for problematic fields
	// These fields need special handling to ensure they're properly set
	{
		// Create a special update object with all the problematic fields
		// Make sure to use the exact field names expected by the JumpCloud API
		specialUpdate := map[string]any{
			// Mapeamento correto: password_recovery_email -> recoveryEmail
			"recoveryEmail": map[string]string{
				"address": passwordRecoveryEmail,
			},
			// Mapeamento correto: local_user_account -> systemUsername
			"systemUsername": localUserAccount,
			// Tentar vários formatos de nome para os mesmos campos
			// Nomes corretos da API
			"password_never_expires":        passwordNeverExpires,
			"disableDeviceMaxLoginAttempts": bypassManagedDeviceLockout,

			// Usar os novos campos com fallback para os antigos
			"enable_user_portal_multifactor": getFirstDefinedBool(d, []string{"require_mfa", "mfa_enabled"}),

			// Nomes corretos da API
			"sudo":               getFirstDefinedBool(d, []string{"enable_global_admin_sudo", "sudo"}),
			"passwordless_sudo":  d.Get("global_passwordless_sudo").(bool),
			"ldap_binding_user":  d.Get("ldap_binding_user").(bool),
			"enable_managed_uid": d.Get("enforce_uid_gid_consistency").(bool),

			// Adicionando campos de autoridade que estavam faltando
			// Tratar como objetos ou nulos, não como strings simples
			"delegatedAuthority": formatAuthorityField(d.Get("delegated_authority").(string)),

			// Tratamento especial para password_authority
			"passwordAuthority": formatPasswordAuthority(d.Get("password_authority").(string)),

			// Adicionar restrictedFields quando password_authority for "Scim"
			"restrictedFields": func() any {
				if d.Get("password_authority").(string) == "Scim" {
					return []map[string]any{
						{
							"field": "password",
							"type":  "scim",
							"id":    nil,
						},
					}
				}
				return nil
			}(),
		}

		// Convert to JSON
		specialUpdateJSON, err := json.Marshal(specialUpdate)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error serializing special update: %v", err))
		}

		// Make a PUT request to update just these fields
		// JumpCloud API doesn't support PATCH, so we need to use PUT
		tflog.Debug(ctx, fmt.Sprintf("Making special update for problematic fields for user ID: %s", userID))
		tflog.Debug(ctx, fmt.Sprintf("Special update request body: %s", string(specialUpdateJSON)))
		_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/systemusers/%s", userID), specialUpdateJSON)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error making special update for user %s: %v", userID, err))
		}
	}

	// Preserve os valores dos campos de recuperação de email e conta local
	// que não são retornados corretamente pela API
	if err := d.Set("password_recovery_email", passwordRecoveryEmail); err != nil {
		return diag.FromErr(fmt.Errorf("error setting password_recovery_email: %v", err))
	}

	if err := d.Set("local_user_account", localUserAccount); err != nil {
		return diag.FromErr(fmt.Errorf("error setting local_user_account: %v", err))
	}

	// Preserve os valores dos campos de autoridade
	if err := d.Set("password_authority", d.Get("password_authority").(string)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting password_authority: %v", err))
	}

	if err := d.Set("delegated_authority", d.Get("delegated_authority").(string)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting delegated_authority: %v", err))
	}

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	userID := d.Id()

	// Check if the user has Samba service enabled
	isSambaServiceUser := d.Get("samba_service_user").(bool)

	// If the user has Samba service enabled, we need to disable it first
	if isSambaServiceUser {
		tflog.Debug(ctx, fmt.Sprintf("User %s has Samba service enabled, disabling it before deletion", userID))

		// Create an update to disable Samba service
		sambaUpdate := map[string]any{
			"samba_service_user": false,
		}

		// Convert to JSON
		sambaUpdateJSON, err := json.Marshal(sambaUpdate)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error serializing Samba update: %v", err))
		}

		// Make a PUT request to disable Samba service
		tflog.Debug(ctx, fmt.Sprintf("Disabling Samba service for user ID: %s", userID))
		tflog.Debug(ctx, fmt.Sprintf("Samba update request body: %s", string(sambaUpdateJSON)))
		_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/systemusers/%s", userID), sambaUpdateJSON)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error disabling Samba service for user %s: %v", userID, err))
		}

		// Wait a moment for the change to take effect
		time.Sleep(2 * time.Second)
	}

	// Delete user via API
	// Use the same direct API path as in create, read, and update
	tflog.Debug(ctx, fmt.Sprintf("Deleting user with ID: %s", userID))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/systemusers/%s", userID), nil)
	if err != nil {
		// If the resource is already gone, don't return an error
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("User %s not found during delete, removing from state", userID))
			d.SetId("")
			return nil
		}

		// Check if the error is related to Samba service
		if strings.Contains(err.Error(), "Samba service is enabled") {
			return diag.FromErr(fmt.Errorf("error deleting user %s: %v. Please disable Samba service before deleting the user", userID, err))
		}

		return diag.FromErr(fmt.Errorf("error deleting user %s: %v", userID, err))
	}

	return nil
}

// buildMFAConfig builds the MFA configuration based on user state
// Automatically converts exclusion_days to the correct format based on state
func buildMFAConfig(d *schema.ResourceData, userState string) *MFAConfig {
	if v, ok := d.GetOk("mfa"); ok {
		mfaList := v.([]any)
		if len(mfaList) > 0 {
			mfaMap := mfaList[0].(map[string]any)

			// Normalize state for comparison
			normalizedState := strings.ToUpper(userState)

			// If user is SUSPENDED, don't send MFA block
			if normalizedState == "SUSPENDED" {
				return nil
			}

			config := &MFAConfig{
				Exclusion:  mfaMap["exclusion"].(bool),
				Configured: mfaMap["configured"].(bool),
			}

			// Process exclusion_days based on state
			if days, ok := mfaMap["exclusion_days"]; ok && days.(int) > 0 {
				exclusionDays := days.(int)

				switch normalizedState {
				case "STAGED":
					// For STAGED, use the value directly
					config.ExclusionDays = exclusionDays

				case "ACTIVATED":
					// For ACTIVATED, calculate future date
					futureDate := time.Now().AddDate(0, 0, exclusionDays)
					config.ExclusionUntil = futureDate.Format(time.RFC3339)

				default:
					// For other states, use STAGED as default
					config.ExclusionDays = exclusionDays
				}
			}

			return config
		}
	}
	return nil
}

// logMFAConversion logs the automatic MFA conversion for debugging
func logMFAConversion(ctx context.Context, userState string, days int, config *MFAConfig) {
	switch userState {
	case "STAGED":
		tflog.Debug(ctx, fmt.Sprintf("MFA: Using exclusionDays=%d for STAGED user", config.ExclusionDays))
	case "ACTIVATED":
		tflog.Debug(ctx, fmt.Sprintf("MFA: Converted %d days to exclusionUntil=%s for ACTIVATED user", days, config.ExclusionUntil))
	case "SUSPENDED":
		tflog.Debug(ctx, "MFA: Omitting MFA block for SUSPENDED user")
	}
}

// resourceUserImport imports an existing user by ID
func resourceUserImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// The ID provided during import should be the JumpCloud user ID
	userID := d.Id()

	// Validate that the ID is not empty
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	// Set the ID in the resource data
	d.SetId(userID)

	// Call the read function to populate all the fields
	diags := resourceUserRead(ctx, d, meta)
	if diags.HasError() {
		// Convert diagnostics to error for import
		var errMsgs []string
		for _, diag := range diags {
			errMsgs = append(errMsgs, diag.Summary)
		}
		return nil, fmt.Errorf("failed to read user during import: %s", strings.Join(errMsgs, "; "))
	}

	// Check if the resource was found (ID would be empty if not found)
	if d.Id() == "" {
		return nil, fmt.Errorf("user with ID %s not found", userID)
	}

	return []*schema.ResourceData{d}, nil
}
