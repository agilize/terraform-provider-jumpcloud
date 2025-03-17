package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// SSOApplicationSAML representa os metadados SAML para uma aplicação SSO
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

// SSOApplicationOIDC representa os metadados OIDC para uma aplicação SSO
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

// SSOApplication representa uma aplicação SSO no JumpCloud
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

// resourceSSOApplication retorna o recurso para gerenciar aplicações SSO
func resourceSSOApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSSOApplicationCreate,
		ReadContext:   resourceSSOApplicationRead,
		UpdateContext: resourceSSOApplicationUpdate,
		DeleteContext: resourceSSOApplicationDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome único da aplicação SSO",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Nome de exibição da aplicação",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da aplicação",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"saml", "oidc"}, false),
				Description:  "Tipo de aplicação SSO: SAML ou OIDC",
				ForceNew:     true, // Não permite alterar o tipo após a criação
			},
			"sso_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL para acesso à aplicação pelo usuário",
			},
			"logo_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL do logotipo da aplicação",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Se a aplicação está ativa",
			},
			"beta_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se a aplicação está em acesso beta",
			},
			"require_mfa": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Se a aplicação requer MFA para acesso",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"group_associations": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs dos grupos associados à aplicação",
			},
			"user_associations": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "IDs dos usuários associados à aplicação",
			},
			// Configuração específica SAML
			"saml": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entity_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Entity ID do Service Provider (SP)",
						},
						"assertion_consumer_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "URL para onde as asserções SAML serão enviadas",
						},
						"sp_certificate": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Certificado do Service Provider (SP)",
						},
						"idp_certificate": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Certificado do Identity Provider (IdP)",
						},
						"idp_entity_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Entity ID do Identity Provider (IdP)",
						},
						"idp_sso_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL de SSO do Identity Provider (IdP)",
						},
						"name_id_format": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "email",
							ValidateFunc: validation.StringInSlice([]string{"email", "persistent", "transient", "unspecified"}, false),
							Description:  "Formato do NameID (email, persistent, transient, unspecified)",
						},
						"saml_signing_algorithm": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "sha256",
							ValidateFunc: validation.StringInSlice([]string{"sha1", "sha256", "sha512"}, false),
							Description:  "Algoritmo de assinatura SAML (sha1, sha256, sha512)",
						},
						"sign_assertion": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Se a asserção SAML deve ser assinada",
						},
						"sign_response": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Se a resposta SAML deve ser assinada",
						},
						"encrypt_assertion": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Se a asserção SAML deve ser criptografada",
						},
						"default_relay_state": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Estado de relay padrão para redirecionamento após autenticação",
						},
						"attribute_statements": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Declarações de atributos SAML",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Nome do atributo SAML",
									},
									"name_format": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "unspecified",
										ValidateFunc: validation.StringInSlice([]string{"unspecified", "uri", "basic"}, false),
										Description:  "Formato do nome do atributo (unspecified, uri, basic)",
									},
									"value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Valor ou expressão para o atributo",
									},
								},
							},
						},
					},
				},
			},
			// Configuração específica OIDC
			"oidc": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do cliente OIDC",
						},
						"client_secret": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "Segredo do cliente OIDC",
						},
						"redirect_uris": {
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "URIs de redirecionamento permitidas",
						},
						"response_types": {
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Tipos de resposta permitidos (code, token, id_token)",
						},
						"grant_types": {
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Tipos de concessão permitidos (authorization_code, implicit, refresh_token)",
						},
						"authorization_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL de autorização do provedor",
						},
						"token_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL do token do provedor",
						},
						"user_info_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL de informações do usuário",
						},
						"jwks_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "URL do JWKS (JSON Web Key Set)",
						},
						"scopes": {
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Escopos permitidos (openid, profile, email, etc)",
						},
					},
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da aplicação",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da aplicação",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// resourceSSOApplicationCreate cria uma nova aplicação SSO
func resourceSSOApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir aplicação SSO
	application := &SSOApplication{
		Name:       d.Get("name").(string),
		Type:       d.Get("type").(string),
		Active:     d.Get("active").(bool),
		BetaAccess: d.Get("beta_access").(bool),
		RequireMFA: d.Get("require_mfa").(bool),
	}

	// Campos opcionais
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

	// Processar configurações específicas do tipo
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

		// Processar declarações de atributos
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

		// Processar URIs de redirecionamento
		if v, ok := oidcConfig["redirect_uris"].([]interface{}); ok && len(v) > 0 {
			redirectURIs := make([]string, len(v))
			for i, uri := range v {
				redirectURIs[i] = uri.(string)
			}
			application.OIDC.RedirectURIs = redirectURIs
		}

		// Processar tipos de resposta
		if v, ok := oidcConfig["response_types"].([]interface{}); ok && len(v) > 0 {
			responseTypes := make([]string, len(v))
			for i, rt := range v {
				responseTypes[i] = rt.(string)
			}
			application.OIDC.ResponseTypes = responseTypes
		}

		// Processar tipos de concessão
		if v, ok := oidcConfig["grant_types"].([]interface{}); ok && len(v) > 0 {
			grantTypes := make([]string, len(v))
			for i, gt := range v {
				grantTypes[i] = gt.(string)
			}
			application.OIDC.GrantTypes = grantTypes
		}

		// Processar escopos
		if v, ok := oidcConfig["scopes"].([]interface{}); ok && len(v) > 0 {
			scopes := make([]string, len(v))
			for i, s := range v {
				scopes[i] = s.(string)
			}
			application.OIDC.Scopes = scopes
		}
	}

	// Processar associações de grupos
	if v, ok := d.GetOk("group_associations"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		application.GroupAssociations = groupIDs
	}

	// Processar associações de usuários
	if v, ok := d.GetOk("user_associations"); ok {
		users := v.(*schema.Set).List()
		userIDs := make([]string, len(users))
		for i, u := range users {
			userIDs[i] = u.(string)
		}
		application.UserAssociations = userIDs
	}

	// Serializar para JSON
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar aplicação SSO: %v", err))
	}

	// Criar aplicação via API
	tflog.Debug(ctx, fmt.Sprintf("Criando aplicação SSO: %s", application.Name))
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/applications", applicationJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar aplicação SSO: %v", err))
	}

	// Deserializar resposta
	var createdApplication SSOApplication
	if err := json.Unmarshal(resp, &createdApplication); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdApplication.ID == "" {
		return diag.FromErr(fmt.Errorf("aplicação SSO criada sem ID"))
	}

	d.SetId(createdApplication.ID)
	return resourceSSOApplicationRead(ctx, d, meta)
}

// resourceSSOApplicationRead lê os detalhes de uma aplicação SSO
func resourceSSOApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da aplicação SSO não fornecido"))
	}

	// Buscar aplicação via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo aplicação SSO com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/applications/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Aplicação SSO %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler aplicação SSO: %v", err))
	}

	// Deserializar resposta
	var application SSOApplication
	if err := json.Unmarshal(resp, &application); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
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

	// Configuração específica SAML
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

		// Processar declarações de atributos, se existirem
		if application.SAML.AttributeStatements != nil && len(application.SAML.AttributeStatements) > 0 {
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

	// Configuração específica OIDC
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

	// Obter associações de grupos
	if application.GroupAssociations != nil {
		d.Set("group_associations", application.GroupAssociations)
	}

	// Obter associações de usuários
	if application.UserAssociations != nil {
		d.Set("user_associations", application.UserAssociations)
	}

	return diags
}

// resourceSSOApplicationUpdate atualiza uma aplicação SSO existente
func resourceSSOApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da aplicação SSO não fornecido"))
	}

	// Construir aplicação SSO atualizada
	application := &SSOApplication{
		ID:         id,
		Name:       d.Get("name").(string),
		Type:       d.Get("type").(string),
		Active:     d.Get("active").(bool),
		BetaAccess: d.Get("beta_access").(bool),
		RequireMFA: d.Get("require_mfa").(bool),
	}

	// Campos opcionais
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

	// Processar configurações específicas do tipo
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

		// Processar declarações de atributos
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

		// Processar URIs de redirecionamento
		if v, ok := oidcConfig["redirect_uris"].([]interface{}); ok && len(v) > 0 {
			redirectURIs := make([]string, len(v))
			for i, uri := range v {
				redirectURIs[i] = uri.(string)
			}
			application.OIDC.RedirectURIs = redirectURIs
		}

		// Processar tipos de resposta
		if v, ok := oidcConfig["response_types"].([]interface{}); ok && len(v) > 0 {
			responseTypes := make([]string, len(v))
			for i, rt := range v {
				responseTypes[i] = rt.(string)
			}
			application.OIDC.ResponseTypes = responseTypes
		}

		// Processar tipos de concessão
		if v, ok := oidcConfig["grant_types"].([]interface{}); ok && len(v) > 0 {
			grantTypes := make([]string, len(v))
			for i, gt := range v {
				grantTypes[i] = gt.(string)
			}
			application.OIDC.GrantTypes = grantTypes
		}

		// Processar escopos
		if v, ok := oidcConfig["scopes"].([]interface{}); ok && len(v) > 0 {
			scopes := make([]string, len(v))
			for i, s := range v {
				scopes[i] = s.(string)
			}
			application.OIDC.Scopes = scopes
		}
	}

	// Processar associações de grupos
	if v, ok := d.GetOk("group_associations"); ok {
		groups := v.(*schema.Set).List()
		groupIDs := make([]string, len(groups))
		for i, g := range groups {
			groupIDs[i] = g.(string)
		}
		application.GroupAssociations = groupIDs
	}

	// Processar associações de usuários
	if v, ok := d.GetOk("user_associations"); ok {
		users := v.(*schema.Set).List()
		userIDs := make([]string, len(users))
		for i, u := range users {
			userIDs[i] = u.(string)
		}
		application.UserAssociations = userIDs
	}

	// Serializar para JSON
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar aplicação SSO: %v", err))
	}

	// Atualizar aplicação via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando aplicação SSO: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/applications/%s", id), applicationJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar aplicação SSO: %v", err))
	}

	// Deserializar resposta
	var updatedApplication SSOApplication
	if err := json.Unmarshal(resp, &updatedApplication); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceSSOApplicationRead(ctx, d, meta)
}

// resourceSSOApplicationDelete exclui uma aplicação SSO
func resourceSSOApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Obter cliente
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da aplicação SSO não fornecido"))
	}

	// Excluir aplicação via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo aplicação SSO: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/applications/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Aplicação SSO %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir aplicação SSO: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
