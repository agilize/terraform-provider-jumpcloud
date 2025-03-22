package sso

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
)

// SSOApplicationSAML represents the SAML metadata for an SSO application
type SSOApplicationSAML struct {
	EntityID             string                   `json:"entityId,omitempty"`
	AssertionConsumerURL string                   `json:"assertionConsumerUrl,omitempty"`
	SPCertificate        string                   `json:"spCertificate,omitempty"`
	IDPCertificate       string                   `json:"idpCertificate,omitempty"`
	IDPEntityID          string                   `json:"idpEntityId,omitempty"`
	IDPSSOURL            string                   `json:"idpSsoUrl,omitempty"`
	NameIDFormat         string                   `json:"nameIdFormat,omitempty"`
	SAMLSigningAlgorithm string                   `json:"samlSigningAlgorithm,omitempty"`
	SignAssertion        bool                     `json:"signAssertion"`
	SignResponse         bool                     `json:"signResponse"`
	EncryptAssertion     bool                     `json:"encryptAssertion"`
	DefaultRelayState    string                   `json:"defaultRelayState,omitempty"`
	AttributeStatements  []map[string]interface{} `json:"attributeStatements,omitempty"`
}

// SSOApplicationOIDC represents the OIDC metadata for an SSO application
type SSOApplicationOIDC struct {
	ClientID         string   `json:"clientId,omitempty"`
	ClientSecret     string   `json:"clientSecret,omitempty"`
	RedirectURIs     []string `json:"redirectUris,omitempty"`
	ResponseTypes    []string `json:"responseTypes,omitempty"`
	GrantTypes       []string `json:"grantTypes,omitempty"`
	AuthorizationURL string   `json:"authorizationUrl,omitempty"`
	TokenURL         string   `json:"tokenUrl,omitempty"`
	UserInfoURL      string   `json:"userInfoUrl,omitempty"`
	JwksURL          string   `json:"jwksUrl,omitempty"`
	Scopes           []string `json:"scopes,omitempty"`
}

// SSOApplication represents an SSO application in JumpCloud
type SSOApplication struct {
	ID                string                 `json:"_id,omitempty"`
	Name              string                 `json:"name"`
	DisplayName       string                 `json:"displayName,omitempty"`
	Description       string                 `json:"description,omitempty"`
	Type              string                 `json:"type"` // saml, oidc
	SSOURL            string                 `json:"ssoUrl,omitempty"`
	LogoURL           string                 `json:"logoUrl,omitempty"`
	Active            bool                   `json:"active"`
	SAML              *SSOApplicationSAML    `json:"saml,omitempty"`
	OIDC              *SSOApplicationOIDC    `json:"oidc,omitempty"`
	Config            map[string]interface{} `json:"config,omitempty"`
	Created           string                 `json:"created,omitempty"`
	Updated           string                 `json:"updated,omitempty"`
	BetaAccess        bool                   `json:"betaAccess"`
	RequireMFA        bool                   `json:"requireMFA"`
	GroupAssociations []string               `json:"groupAssociations,omitempty"`
	UserAssociations  []string               `json:"userAssociations,omitempty"`
	OrgID             string                 `json:"orgId,omitempty"`
}

// ResourceSSOApplication returns the resource schema for JumpCloud SSO applications
func ResourceSSOApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSSOApplicationCreate,
		ReadContext:   resourceSSOApplicationRead,
		UpdateContext: resourceSSOApplicationUpdate,
		DeleteContext: resourceSSOApplicationDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name of the SSO application",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Display name of the application",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the application",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"saml", "oidc"}, false),
				Description:  "Type of SSO application: SAML or OIDC",
				ForceNew:     true, // Type cannot be changed after creation
			},
			"sso_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL for user access to the application",
			},
			"logo_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL of the application logo",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the application is active",
			},
			"beta_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the application is in beta access",
			},
			"require_mfa": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the application requires MFA for access",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"group_associations": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs of groups associated with the application",
			},
			"user_associations": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs of users associated with the application",
			},
			// SAML specific configuration
			"saml": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entity_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Entity ID of the Service Provider (SP)",
						},
						"assertion_consumer_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "URL where SAML assertions will be sent",
						},
						"sp_certificate": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Service Provider (SP) certificate",
						},
						"idp_certificate": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identity Provider (IdP) certificate",
						},
						"idp_entity_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Entity ID of the Identity Provider (IdP)",
						},
						"idp_sso_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "SSO URL of the Identity Provider (IdP)",
						},
						"name_id_format": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "email",
							ValidateFunc: validation.StringInSlice([]string{"email", "persistent", "transient", "unspecified"}, false),
							Description:  "NameID format (email, persistent, transient, unspecified)",
						},
						"saml_signing_algorithm": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "sha256",
							ValidateFunc: validation.StringInSlice([]string{"sha1", "sha256", "sha512"}, false),
							Description:  "SAML signing algorithm (sha1, sha256, sha512)",
						},
						"sign_assertion": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether the SAML assertion should be signed",
						},
						"sign_response": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether the SAML response should be signed",
						},
						"encrypt_assertion": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether the SAML assertion should be encrypted",
						},
						"default_relay_state": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Default relay state for redirect after authentication",
						},
						"attribute_statements": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "SAML attribute statements",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Name of the SAML attribute",
									},
									"name_format": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "unspecified",
										ValidateFunc: validation.StringInSlice([]string{"unspecified", "uri", "basic"}, false),
										Description:  "Format of the attribute name (unspecified, uri, basic)",
									},
									"value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Value or expression for the attribute",
									},
								},
							},
						},
					},
				},
			},
			// OIDC specific configuration
			"oidc": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "OIDC client ID",
						},
						"client_secret": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "OIDC client secret",
						},
						"redirect_uris": {
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Allowed redirect URIs",
						},
						"response_types": {
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Allowed response types (code, token, id_token)",
						},
						"grant_types": {
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Allowed grant types (authorization_code, implicit, refresh_token)",
						},
						"authorization_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Authorization URL of the provider",
						},
						"token_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Token URL of the provider",
						},
						"user_info_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "User info URL",
						},
						"jwks_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "JWKS (JSON Web Key Set) URL",
						},
						"scopes": {
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Allowed scopes (openid, profile, email, etc)",
						},
					},
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the application",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of the last update to the application",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// resourceSSOApplicationCreate creates a new SSO application
func resourceSSOApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(apiclient.Client)
	if !ok {
		return diag.FromErr(fmt.Errorf("invalid client configuration"))
	}

	// Build SSO application
	application := &SSOApplication{
		Name:       d.Get("name").(string),
		Type:       d.Get("type").(string),
		Active:     d.Get("active").(bool),
		BetaAccess: d.Get("beta_access").(bool),
		RequireMFA: d.Get("require_mfa").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("display_name"); ok {
		application.DisplayName = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		application.Description = v.(string)
	}

	if v, ok := d.GetOk("sso_url"); ok {
		application.SSOURL = v.(string)
	}

	if v, ok := d.GetOk("logo_url"); ok {
		application.LogoURL = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		application.OrgID = v.(string)
	}

	// Process type-specific configurations
	if application.Type == "saml" && d.Get("saml") != nil && len(d.Get("saml").([]interface{})) > 0 {
		samlConfig := d.Get("saml").([]interface{})[0].(map[string]interface{})

		application.SAML = &SSOApplicationSAML{
			EntityID:             samlConfig["entity_id"].(string),
			AssertionConsumerURL: samlConfig["assertion_consumer_url"].(string),
			NameIDFormat:         samlConfig["name_id_format"].(string),
			SAMLSigningAlgorithm: samlConfig["saml_signing_algorithm"].(string),
			SignAssertion:        samlConfig["sign_assertion"].(bool),
			SignResponse:         samlConfig["sign_response"].(bool),
			EncryptAssertion:     samlConfig["encrypt_assertion"].(bool),
		}

		if v, ok := samlConfig["sp_certificate"]; ok && v.(string) != "" {
			application.SAML.SPCertificate = v.(string)
		}

		if v, ok := samlConfig["default_relay_state"]; ok && v.(string) != "" {
			application.SAML.DefaultRelayState = v.(string)
		}

		// Process attribute statements
		if attrStatements, ok := samlConfig["attribute_statements"].([]interface{}); ok && len(attrStatements) > 0 {
			statements := make([]map[string]interface{}, len(attrStatements))

			for i, stmt := range attrStatements {
				stmtMap := stmt.(map[string]interface{})
				statements[i] = map[string]interface{}{
					"name":       stmtMap["name"].(string),
					"nameFormat": stmtMap["name_format"].(string),
					"value":      stmtMap["value"].(string),
				}
			}

			application.SAML.AttributeStatements = statements
		}
	} else if application.Type == "oidc" && d.Get("oidc") != nil && len(d.Get("oidc").([]interface{})) > 0 {
		oidcConfig := d.Get("oidc").([]interface{})[0].(map[string]interface{})

		application.OIDC = &SSOApplicationOIDC{}

		// Process redirect URIs
		if v, ok := oidcConfig["redirect_uris"].([]interface{}); ok && len(v) > 0 {
			redirectURIs := make([]string, len(v))
			for i, uri := range v {
				redirectURIs[i] = uri.(string)
			}
			application.OIDC.RedirectURIs = redirectURIs
		}

		// Process response types
		if v, ok := oidcConfig["response_types"].([]interface{}); ok && len(v) > 0 {
			responseTypes := make([]string, len(v))
			for i, rt := range v {
				responseTypes[i] = rt.(string)
			}
			application.OIDC.ResponseTypes = responseTypes
		}

		// Process grant types
		if v, ok := oidcConfig["grant_types"].([]interface{}); ok && len(v) > 0 {
			grantTypes := make([]string, len(v))
			for i, gt := range v {
				grantTypes[i] = gt.(string)
			}
			application.OIDC.GrantTypes = grantTypes
		}

		// Process scopes
		if v, ok := oidcConfig["scopes"].([]interface{}); ok && len(v) > 0 {
			scopes := make([]string, len(v))
			for i, s := range v {
				scopes[i] = s.(string)
			}
			application.OIDC.Scopes = scopes
		}
	}

	// Process group associations
	if v, ok := d.GetOk("group_associations"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		application.GroupAssociations = groupIDs
	}

	// Process user associations
	if v, ok := d.GetOk("user_associations"); ok {
		users := v.(*schema.Set).List()
		userIDs := make([]string, len(users))
		for i, u := range users {
			userIDs[i] = u.(string)
		}
		application.UserAssociations = userIDs
	}

	// Serialize to JSON
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing SSO application: %v", err))
	}

	// Create application via API
	tflog.Debug(ctx, fmt.Sprintf("Creating SSO application: %s", application.Name))
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/applications", applicationJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating SSO application: %v", err))
	}

	// Deserialize response
	var createdApplication SSOApplication
	if err := json.Unmarshal(resp, &createdApplication); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdApplication.ID == "" {
		return diag.FromErr(fmt.Errorf("SSO application created without ID"))
	}

	d.SetId(createdApplication.ID)
	return resourceSSOApplicationRead(ctx, d, meta)
}

// resourceSSOApplicationRead reads the details of an SSO application
func resourceSSOApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, ok := meta.(apiclient.Client)
	if !ok {
		return diag.FromErr(fmt.Errorf("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("SSO application ID not provided"))
	}

	// Fetch application via API
	tflog.Debug(ctx, fmt.Sprintf("Reading SSO application with ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/applications/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("SSO application %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading SSO application: %v", err))
	}

	// Deserialize response
	var application SSOApplication
	if err := json.Unmarshal(resp, &application); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
	d.Set("name", application.Name)
	d.Set("display_name", application.DisplayName)
	d.Set("description", application.Description)
	d.Set("type", application.Type)
	d.Set("sso_url", application.SSOURL)
	d.Set("logo_url", application.LogoURL)
	d.Set("active", application.Active)
	d.Set("beta_access", application.BetaAccess)
	d.Set("require_mfa", application.RequireMFA)
	d.Set("created", application.Created)
	d.Set("updated", application.Updated)

	if application.OrgID != "" {
		d.Set("org_id", application.OrgID)
	}

	// SAML specific configuration
	if application.Type == "saml" && application.SAML != nil {
		samlConfig := []map[string]interface{}{
			{
				"entity_id":              application.SAML.EntityID,
				"assertion_consumer_url": application.SAML.AssertionConsumerURL,
				"sp_certificate":         application.SAML.SPCertificate,
				"idp_certificate":        application.SAML.IDPCertificate,
				"idp_entity_id":          application.SAML.IDPEntityID,
				"idp_sso_url":            application.SAML.IDPSSOURL,
				"name_id_format":         application.SAML.NameIDFormat,
				"saml_signing_algorithm": application.SAML.SAMLSigningAlgorithm,
				"sign_assertion":         application.SAML.SignAssertion,
				"sign_response":          application.SAML.SignResponse,
				"encrypt_assertion":      application.SAML.EncryptAssertion,
				"default_relay_state":    application.SAML.DefaultRelayState,
			},
		}

		// Process attribute statements if they exist
		if len(application.SAML.AttributeStatements) > 0 {
			attrStatements := make([]map[string]interface{}, len(application.SAML.AttributeStatements))

			for i, stmt := range application.SAML.AttributeStatements {
				attrStatements[i] = map[string]interface{}{
					"name":        stmt["name"],
					"name_format": stmt["nameFormat"],
					"value":       stmt["value"],
				}
			}

			samlConfig[0]["attribute_statements"] = attrStatements
		}

		d.Set("saml", samlConfig)
	}

	// OIDC specific configuration
	if application.Type == "oidc" && application.OIDC != nil {
		oidcConfig := []map[string]interface{}{
			{
				"client_id":         application.OIDC.ClientID,
				"client_secret":     application.OIDC.ClientSecret,
				"redirect_uris":     application.OIDC.RedirectURIs,
				"response_types":    application.OIDC.ResponseTypes,
				"grant_types":       application.OIDC.GrantTypes,
				"authorization_url": application.OIDC.AuthorizationURL,
				"token_url":         application.OIDC.TokenURL,
				"user_info_url":     application.OIDC.UserInfoURL,
				"jwks_url":          application.OIDC.JwksURL,
				"scopes":            application.OIDC.Scopes,
			},
		}

		d.Set("oidc", oidcConfig)
	}

	// Get group associations
	if application.GroupAssociations != nil {
		d.Set("group_associations", application.GroupAssociations)
	}

	// Get user associations
	if application.UserAssociations != nil {
		d.Set("user_associations", application.UserAssociations)
	}

	return diags
}

// resourceSSOApplicationUpdate updates an existing SSO application
func resourceSSOApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(apiclient.Client)
	if !ok {
		return diag.FromErr(fmt.Errorf("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("SSO application ID not provided"))
	}

	// Build updated SSO application
	application := &SSOApplication{
		ID:         id,
		Name:       d.Get("name").(string),
		Type:       d.Get("type").(string),
		Active:     d.Get("active").(bool),
		BetaAccess: d.Get("beta_access").(bool),
		RequireMFA: d.Get("require_mfa").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("display_name"); ok {
		application.DisplayName = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		application.Description = v.(string)
	}

	if v, ok := d.GetOk("sso_url"); ok {
		application.SSOURL = v.(string)
	}

	if v, ok := d.GetOk("logo_url"); ok {
		application.LogoURL = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		application.OrgID = v.(string)
	}

	// Process type-specific configurations
	if application.Type == "saml" && d.Get("saml") != nil && len(d.Get("saml").([]interface{})) > 0 {
		samlConfig := d.Get("saml").([]interface{})[0].(map[string]interface{})

		application.SAML = &SSOApplicationSAML{
			EntityID:             samlConfig["entity_id"].(string),
			AssertionConsumerURL: samlConfig["assertion_consumer_url"].(string),
			NameIDFormat:         samlConfig["name_id_format"].(string),
			SAMLSigningAlgorithm: samlConfig["saml_signing_algorithm"].(string),
			SignAssertion:        samlConfig["sign_assertion"].(bool),
			SignResponse:         samlConfig["sign_response"].(bool),
			EncryptAssertion:     samlConfig["encrypt_assertion"].(bool),
		}

		if v, ok := samlConfig["sp_certificate"]; ok && v.(string) != "" {
			application.SAML.SPCertificate = v.(string)
		}

		if v, ok := samlConfig["default_relay_state"]; ok && v.(string) != "" {
			application.SAML.DefaultRelayState = v.(string)
		}

		// Process attribute statements
		if attrStatements, ok := samlConfig["attribute_statements"].([]interface{}); ok && len(attrStatements) > 0 {
			statements := make([]map[string]interface{}, len(attrStatements))

			for i, stmt := range attrStatements {
				stmtMap := stmt.(map[string]interface{})
				statements[i] = map[string]interface{}{
					"name":       stmtMap["name"].(string),
					"nameFormat": stmtMap["name_format"].(string),
					"value":      stmtMap["value"].(string),
				}
			}

			application.SAML.AttributeStatements = statements
		}
	} else if application.Type == "oidc" && d.Get("oidc") != nil && len(d.Get("oidc").([]interface{})) > 0 {
		oidcConfig := d.Get("oidc").([]interface{})[0].(map[string]interface{})

		application.OIDC = &SSOApplicationOIDC{}

		// Process redirect URIs
		if v, ok := oidcConfig["redirect_uris"].([]interface{}); ok && len(v) > 0 {
			redirectURIs := make([]string, len(v))
			for i, uri := range v {
				redirectURIs[i] = uri.(string)
			}
			application.OIDC.RedirectURIs = redirectURIs
		}

		// Process response types
		if v, ok := oidcConfig["response_types"].([]interface{}); ok && len(v) > 0 {
			responseTypes := make([]string, len(v))
			for i, rt := range v {
				responseTypes[i] = rt.(string)
			}
			application.OIDC.ResponseTypes = responseTypes
		}

		// Process grant types
		if v, ok := oidcConfig["grant_types"].([]interface{}); ok && len(v) > 0 {
			grantTypes := make([]string, len(v))
			for i, gt := range v {
				grantTypes[i] = gt.(string)
			}
			application.OIDC.GrantTypes = grantTypes
		}

		// Process scopes
		if v, ok := oidcConfig["scopes"].([]interface{}); ok && len(v) > 0 {
			scopes := make([]string, len(v))
			for i, s := range v {
				scopes[i] = s.(string)
			}
			application.OIDC.Scopes = scopes
		}
	}

	// Process group associations
	if v, ok := d.GetOk("group_associations"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		application.GroupAssociations = groupIDs
	}

	// Process user associations
	if v, ok := d.GetOk("user_associations"); ok {
		users := v.(*schema.Set).List()
		userIDs := make([]string, len(users))
		for i, u := range users {
			userIDs[i] = u.(string)
		}
		application.UserAssociations = userIDs
	}

	// Serialize to JSON
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing SSO application: %v", err))
	}

	// Update application via API
	tflog.Debug(ctx, fmt.Sprintf("Updating SSO application: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/applications/%s", id), applicationJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating SSO application: %v", err))
	}

	return resourceSSOApplicationRead(ctx, d, meta)
}

// resourceSSOApplicationDelete deletes an SSO application
func resourceSSOApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, ok := meta.(apiclient.Client)
	if !ok {
		return diag.FromErr(fmt.Errorf("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("SSO application ID not provided"))
	}

	// Delete application via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting SSO application: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/applications/%s", id), nil)
	if err != nil {
		// If the resource is not found, consider it already deleted
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("SSO application %s not found, considering it deleted", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error deleting SSO application: %v", err))
	}

	// Remove from state
	d.SetId("")
	return diags
}

// isNotFoundError checks if the error is a "not found" error
func isNotFoundError(err error) bool {
	return err != nil && err.Error() == "status code 404"
}
