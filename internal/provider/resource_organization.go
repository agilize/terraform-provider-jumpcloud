package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Organization representa uma organização no JumpCloud
type Organization struct {
	ID             string            `json:"_id,omitempty"`
	Name           string            `json:"name"`
	DisplayName    string            `json:"displayName,omitempty"`
	LogoURL        string            `json:"logoUrl,omitempty"`
	Website        string            `json:"website,omitempty"`
	ContactName    string            `json:"contactName,omitempty"`
	ContactEmail   string            `json:"contactEmail,omitempty"`
	ContactPhone   string            `json:"contactPhone,omitempty"`
	Settings       map[string]string `json:"settings,omitempty"`
	ParentOrgID    string            `json:"parentOrgId,omitempty"`
	AllowedDomains []string          `json:"allowedDomains,omitempty"`
	Created        string            `json:"created,omitempty"`
	Updated        string            `json:"updated,omitempty"`
}

// ValidateEmail valida se um endereço de email é válido
func ValidateEmail(email string) error {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(pattern, email)
	if err != nil {
		return fmt.Errorf("erro ao validar email: %v", err)
	}
	if !matched {
		return fmt.Errorf("endereço de email inválido: %s", email)
	}
	return nil
}

// ValidateDomainPattern valida se um padrão de domínio é válido
func ValidateDomainPattern(domain string) error {
	if strings.HasPrefix(domain, "*.") {
		domain = domain[2:]
	}

	pattern := `^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`
	matched, err := regexp.MatchString(pattern, domain)
	if err != nil {
		return fmt.Errorf("erro ao validar domínio: %v", err)
	}
	if !matched {
		return fmt.Errorf("padrão de domínio inválido: %s", domain)
	}
	return nil
}

// ValidateAllowedDomains valida uma lista de domínios permitidos
func ValidateAllowedDomains(domains []string) error {
	for _, domain := range domains {
		if err := ValidateDomainPattern(domain); err != nil {
			return err
		}
	}
	return nil
}

// expandStringSet converte um schema.Set em uma slice de strings
func expandStringSet(set *schema.Set) []string {
	list := set.List()
	result := make([]string, 0, len(list))
	for _, v := range list {
		result = append(result, v.(string))
	}
	return result
}

// resourceOrganization retorna o recurso para gerenciar organizações no JumpCloud
func resourceOrganization() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationCreate,
		ReadContext:   resourceOrganizationRead,
		UpdateContext: resourceOrganizationUpdate,
		DeleteContext: resourceOrganizationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Nome único da organização",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(
						regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`),
						"deve conter apenas letras, números, espaços, hífens e underscores",
					),
				),
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Nome de exibição da organização",
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"logo_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "URL do logotipo da organização",
				ValidateFunc: validation.IsURLWithHTTPS,
			},
			"website": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Website da organização",
				ValidateFunc: validation.IsURLWithHTTPS,
			},
			"contact_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Nome do contato principal da organização",
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"contact_email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "E-mail do contato principal da organização",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if err := ValidateEmail(value); err != nil {
						errors = append(errors, err)
					}
					return
				},
			},
			"contact_phone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Telefone do contato principal da organização",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^\+?[1-9]\d{1,14}$`),
					"deve ser um número de telefone válido no formato E.164",
				),
			},
			"settings": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Configurações específicas da organização",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"parent_org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID da organização pai (para organizações filhas)",
			},
			"allowed_domains": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:         schema.HashString,
				Description: "Lista de domínios permitidos para a organização",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da organização",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da organização",
			},
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
			// Validar domínios permitidos
			if d.HasChange("allowed_domains") {
				domains := d.Get("allowed_domains").(*schema.Set).List()
				domainsList := make([]string, 0, len(domains))
				for _, domain := range domains {
					domainsList = append(domainsList, domain.(string))
				}
				if err := ValidateAllowedDomains(domainsList); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// resourceOrganizationCreate cria uma nova organização no JumpCloud
func resourceOrganizationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	org := &Organization{
		Name:           d.Get("name").(string),
		DisplayName:    d.Get("display_name").(string),
		ParentOrgID:    d.Get("parent_org_id").(string),
		LogoURL:        d.Get("logo_url").(string),
		Website:        d.Get("website").(string),
		ContactName:    d.Get("contact_name").(string),
		ContactEmail:   d.Get("contact_email").(string),
		ContactPhone:   d.Get("contact_phone").(string),
		AllowedDomains: expandStringSet(d.Get("allowed_domains").(*schema.Set)),
	}

	// Validar domínios antes de criar
	if err := ValidateAllowedDomains(org.AllowedDomains); err != nil {
		return diag.FromErr(err)
	}

	// Serializar o objeto para JSON
	orgJSON, err := json.Marshal(org)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar organização: %v", err))
	}

	tflog.Debug(ctx, "Criando organização no JumpCloud", map[string]interface{}{
		"name": org.Name,
	})

	// Implementar chamada à API do JumpCloud para criar a organização
	responseBody, err := c.DoRequest(http.MethodPost, "/api/v2/organizations", orgJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar organização: %v", err))
	}

	// Decodificar resposta
	var createdOrg Organization
	if err := json.Unmarshal(responseBody, &createdOrg); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao decodificar resposta: %v", err))
	}

	d.SetId(createdOrg.ID)
	return resourceOrganizationRead(ctx, d, m)
}

// resourceOrganizationRead lê uma organização existente no JumpCloud
func resourceOrganizationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	id := d.Id()
	responseBody, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/organizations/%s", id), nil)
	if err != nil {
		// Se a organização não for encontrada, remover do estado
		if IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("erro ao ler organização: %v", err))
	}

	// Decodificar resposta
	var org Organization
	if err := json.Unmarshal(responseBody, &org); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao decodificar resposta: %v", err))
	}

	if err := d.Set("name", org.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("display_name", org.DisplayName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("logo_url", org.LogoURL); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("website", org.Website); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("contact_name", org.ContactName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("contact_email", org.ContactEmail); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("contact_phone", org.ContactPhone); err != nil {
		return diag.FromErr(err)
	}

	// Converter settings de map[string]string para map[string]interface{}
	if org.Settings != nil {
		settingsMap := make(map[string]interface{})
		for k, v := range org.Settings {
			settingsMap[k] = v
		}
		if err := d.Set("settings", settingsMap); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("parent_org_id", org.ParentOrgID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("allowed_domains", org.AllowedDomains); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created", org.Created); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated", org.Updated); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// resourceOrganizationUpdate atualiza uma organização existente no JumpCloud
func resourceOrganizationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	id := d.Id()
	org := &Organization{
		ID:             id,
		Name:           d.Get("name").(string),
		DisplayName:    d.Get("display_name").(string),
		ParentOrgID:    d.Get("parent_org_id").(string),
		LogoURL:        d.Get("logo_url").(string),
		Website:        d.Get("website").(string),
		ContactName:    d.Get("contact_name").(string),
		ContactEmail:   d.Get("contact_email").(string),
		ContactPhone:   d.Get("contact_phone").(string),
		AllowedDomains: expandStringSet(d.Get("allowed_domains").(*schema.Set)),
	}

	// Validar domínios antes de atualizar
	if err := ValidateAllowedDomains(org.AllowedDomains); err != nil {
		return diag.FromErr(err)
	}

	// Serializar o objeto para JSON
	orgJSON, err := json.Marshal(org)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar organização: %v", err))
	}

	tflog.Debug(ctx, "Atualizando organização no JumpCloud", map[string]interface{}{
		"id":   id,
		"name": org.Name,
	})

	// Implementar chamada à API do JumpCloud para atualizar a organização
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/organizations/%s", id), orgJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar organização: %v", err))
	}

	return resourceOrganizationRead(ctx, d, m)
}

// resourceOrganizationDelete exclui uma organização existente no JumpCloud
func resourceOrganizationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	id := d.Id()

	tflog.Debug(ctx, "Excluindo organização do JumpCloud", map[string]interface{}{
		"id": id,
	})

	// Implementar chamada à API do JumpCloud para excluir a organização
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/organizations/%s", id), nil)
	if err != nil {
		// Se a organização não for encontrada, não é necessário retornar um erro
		if IsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir organização: %v", err))
	}

	d.SetId("")
	return nil
}
