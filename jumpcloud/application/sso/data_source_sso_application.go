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

// DataSourceSSOApplication returns the data source schema for JumpCloud SSO application
func DataSourceSSOApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSSOApplicationRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
				Description:   "ID of the SSO application",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id"},
				Description:   "Name of the SSO application",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Display name of the SSO application",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the SSO application",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the SSO application (saml or oidc)",
			},
			"sso_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SSO URL of the application",
			},
			"logo_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of the application logo",
			},
			"active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the SSO application is active",
			},
			"beta_access": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the SSO application has beta access",
			},
			"require_mfa": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the SSO application requires MFA",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp of the SSO application",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update timestamp of the SSO application",
			},
			"saml": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entity_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Entity ID for the SAML application",
						},
						"assertion_consumer_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Assertion Consumer URL for the SAML application",
						},
						"sp_certificate": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service Provider certificate for the SAML application",
						},
						"idp_certificate": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identity Provider certificate for the SAML application",
						},
						"idp_entity_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identity Provider Entity ID for the SAML application",
						},
						"idp_sso_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identity Provider SSO URL for the SAML application",
						},
						"name_id_format": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name ID format for the SAML application",
						},
						"saml_signing_algorithm": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Signing algorithm for the SAML application",
						},
						"sign_assertion": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to sign the SAML assertion",
						},
						"sign_response": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to sign the SAML response",
						},
						"encrypt_assertion": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to encrypt the SAML assertion",
						},
						"default_relay_state": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Default relay state for the SAML application",
						},
						"attribute_statements": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Attribute statements for the SAML application",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Name of the attribute",
									},
									"values": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "Values for the attribute",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"name_format": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Name format for the attribute",
									},
								},
							},
						},
					},
				},
			},
			"oidc": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Client ID for the OIDC application",
						},
						"client_secret": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "Client secret for the OIDC application",
						},
						"redirect_uris": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Redirect URIs for the OIDC application",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"response_types": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Response types for the OIDC application",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"grant_types": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Grant types for the OIDC application",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"authorization_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Authorization URL for the OIDC application",
						},
						"token_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Token URL for the OIDC application",
						},
						"user_info_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "User info URL for the OIDC application",
						},
						"jwks_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "JWKS URL for the OIDC application",
						},
						"scopes": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Scopes for the OIDC application",
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

func dataSourceSSOApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(apiclient.Client)
	tflog.Debug(ctx, "Reading SSO application data source")

	// Get application by ID or name
	var app SSOApplication

	if id, ok := d.GetOk("id"); ok {
		tflog.Debug(ctx, fmt.Sprintf("Looking up SSO application by ID: %s", id.(string)))

		resBody, err := client.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/applications/%s", id.(string)), nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error retrieving SSO application: %v", err))
		}

		if err := json.Unmarshal(resBody, &app); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing SSO application response: %v", err))
		}

		d.SetId(app.ID)
	} else if name, ok := d.GetOk("name"); ok {
		tflog.Debug(ctx, fmt.Sprintf("Looking up SSO application by name: %s", name.(string)))

		// Get all applications and filter by name
		resBody, err := client.DoRequest(http.MethodGet, "/api/v2/applications", nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error retrieving SSO applications: %v", err))
		}

		var apps []SSOApplication
		if err := json.Unmarshal(resBody, &apps); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing SSO applications response: %v", err))
		}

		found := false
		for _, a := range apps {
			if a.Name == name.(string) {
				app = a
				found = true
				break
			}
		}

		if !found {
			return diag.FromErr(fmt.Errorf("no SSO application found with name: %s", name.(string)))
		}

		d.SetId(app.ID)
	} else {
		return diag.FromErr(fmt.Errorf("either id or name must be specified"))
	}

	// Set the application properties in the data source
	if err := d.Set("name", app.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %v", err))
	}

	if err := d.Set("display_name", app.DisplayName); err != nil {
		return diag.FromErr(fmt.Errorf("error setting display_name: %v", err))
	}

	if err := d.Set("description", app.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %v", err))
	}

	if err := d.Set("type", app.Type); err != nil {
		return diag.FromErr(fmt.Errorf("error setting type: %v", err))
	}

	if err := d.Set("sso_url", app.SSOURL); err != nil {
		return diag.FromErr(fmt.Errorf("error setting sso_url: %v", err))
	}

	if err := d.Set("logo_url", app.LogoURL); err != nil {
		return diag.FromErr(fmt.Errorf("error setting logo_url: %v", err))
	}

	if err := d.Set("active", app.Active); err != nil {
		return diag.FromErr(fmt.Errorf("error setting active: %v", err))
	}

	if err := d.Set("beta_access", app.BetaAccess); err != nil {
		return diag.FromErr(fmt.Errorf("error setting beta_access: %v", err))
	}

	if err := d.Set("require_mfa", app.RequireMFA); err != nil {
		return diag.FromErr(fmt.Errorf("error setting require_mfa: %v", err))
	}

	if err := d.Set("created", app.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %v", err))
	}

	if err := d.Set("updated", app.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
	}

	// Handle SAML specific fields
	if app.Type == "saml" && app.SAML != nil {
		saml := []map[string]interface{}{
			{
				"entity_id":              app.SAML.EntityID,
				"assertion_consumer_url": app.SAML.AssertionConsumerURL,
				"sp_certificate":         app.SAML.SPCertificate,
				"idp_certificate":        app.SAML.IDPCertificate,
				"idp_entity_id":          app.SAML.IDPEntityID,
				"idp_sso_url":            app.SAML.IDPSSOURL,
				"name_id_format":         app.SAML.NameIDFormat,
				"saml_signing_algorithm": app.SAML.SAMLSigningAlgorithm,
				"sign_assertion":         app.SAML.SignAssertion,
				"sign_response":          app.SAML.SignResponse,
				"encrypt_assertion":      app.SAML.EncryptAssertion,
				"default_relay_state":    app.SAML.DefaultRelayState,
			},
		}

		// Handle attribute statements
		if len(app.SAML.AttributeStatements) > 0 {
			attrStatements := make([]map[string]interface{}, len(app.SAML.AttributeStatements))
			for i, stmt := range app.SAML.AttributeStatements {
				attrStatement := map[string]interface{}{
					"name":        stmt["name"],
					"name_format": stmt["nameFormat"],
				}

				// Convert values to list
				if values, ok := stmt["values"]; ok {
					valuesList := values.([]interface{})
					stringValues := make([]string, len(valuesList))
					for j, v := range valuesList {
						stringValues[j] = v.(string)
					}
					attrStatement["values"] = stringValues
				}

				attrStatements[i] = attrStatement
			}
			saml[0]["attribute_statements"] = attrStatements
		}

		if err := d.Set("saml", saml); err != nil {
			return diag.FromErr(fmt.Errorf("error setting saml: %v", err))
		}
	}

	// Handle OIDC specific fields
	if app.Type == "oidc" && app.OIDC != nil {
		oidc := []map[string]interface{}{
			{
				"client_id":         app.OIDC.ClientID,
				"client_secret":     app.OIDC.ClientSecret,
				"redirect_uris":     app.OIDC.RedirectURIs,
				"response_types":    app.OIDC.ResponseTypes,
				"grant_types":       app.OIDC.GrantTypes,
				"authorization_url": app.OIDC.AuthorizationURL,
				"token_url":         app.OIDC.TokenURL,
				"user_info_url":     app.OIDC.UserInfoURL,
				"jwks_url":          app.OIDC.JwksURL,
				"scopes":            app.OIDC.Scopes,
			},
		}
		if err := d.Set("oidc", oidc); err != nil {
			return diag.FromErr(fmt.Errorf("error setting oidc: %v", err))
		}
	}

	return nil
}
