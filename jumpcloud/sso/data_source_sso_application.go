package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
)

// DataSourceSSOApplication returns the schema for the SSO application data source
func DataSourceSSOApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSSOApplicationRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the SSO application",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the SSO application",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Display name of the application",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the application",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of SSO application (saml, oidc)",
			},
			"sso_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL for user access to the application",
			},
			"logo_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the application logo",
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the application is active",
			},
			"beta_access": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the application is in beta access",
			},
			"require_mfa": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the application requires MFA for access",
			},
			"org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Organization ID for multi-tenant environments",
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
			// SAML specific configuration
			"saml": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entity_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Entity ID of the Service Provider (SP)",
						},
						"assertion_consumer_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL where SAML assertions will be sent",
						},
						"sp_certificate": {
							Type:        schema.TypeString,
							Computed:    true,
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
							Type:        schema.TypeString,
							Computed:    true,
							Description: "NameID format",
						},
						"saml_signing_algorithm": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "SAML signing algorithm",
						},
						"sign_assertion": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the SAML assertion is signed",
						},
						"sign_response": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the SAML response is signed",
						},
						"encrypt_assertion": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the SAML assertion is encrypted",
						},
						"default_relay_state": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Default relay state for redirect after authentication",
						},
						"attribute_statements": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "SAML attribute statements",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Name of the SAML attribute",
									},
									"name_format": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Format of the attribute name",
									},
									"value": {
										Type:        schema.TypeString,
										Computed:    true,
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
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "OIDC client ID",
						},
						"redirect_uris": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Allowed redirect URIs",
						},
						"response_types": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Allowed response types",
						},
						"grant_types": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Allowed grant types",
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
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Allowed scopes",
						},
					},
				},
			},
			"group_associations": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs of groups associated with the application",
			},
			"user_associations": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs of users associated with the application",
			},
		},
	}
}

// dataSourceSSOApplicationRead reads the details of an SSO application
func dataSourceSSOApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, ok := meta.(apiclient.Client)
	if !ok {
		return diag.FromErr(fmt.Errorf("invalid client configuration"))
	}

	id := d.Get("id").(string)
	if id == "" {
		return diag.FromErr(fmt.Errorf("SSO application ID not provided"))
	}

	// Set the ID for the data source
	d.SetId(id)

	// Fetch application via API
	tflog.Debug(ctx, fmt.Sprintf("Reading SSO application with ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/applications/%s", id), nil)
	if err != nil {
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
