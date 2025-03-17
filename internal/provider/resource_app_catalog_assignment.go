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

// AppCatalogAssignment representa uma atribuição de aplicação a usuários/grupos no JumpCloud
type AppCatalogAssignment struct {
	ID             string                 `json:"_id,omitempty"`
	ApplicationID  string                 `json:"applicationId"`
	TargetType     string                 `json:"targetType"` // user, group
	TargetID       string                 `json:"targetId"`
	AssignmentType string                 `json:"assignmentType"`          // optional, required
	InstallPolicy  string                 `json:"installPolicy,omitempty"` // auto, manual
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
	OrgID          string                 `json:"orgId,omitempty"`
	Created        string                 `json:"created,omitempty"`
	Updated        string                 `json:"updated,omitempty"`
}

func resourceAppCatalogAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppCatalogAssignmentCreate,
		ReadContext:   resourceAppCatalogAssignmentRead,
		UpdateContext: resourceAppCatalogAssignmentUpdate,
		DeleteContext: resourceAppCatalogAssignmentDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID da aplicação no catálogo",
			},
			"target_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"user", "group"}, false),
				Description:  "Tipo do alvo da atribuição (user, group)",
			},
			"target_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID do usuário ou grupo alvo da atribuição",
			},
			"assignment_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "optional",
				ValidateFunc: validation.StringInSlice([]string{"optional", "required"}, false),
				Description:  "Tipo da atribuição (optional, required)",
			},
			"install_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "manual",
				ValidateFunc: validation.StringInSlice([]string{"auto", "manual"}, false),
				Description:  "Política de instalação (auto, manual)",
			},
			"configuration": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Configuração específica para a aplicação em formato JSON",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					jsonStr := val.(string)
					if jsonStr == "" {
						return
					}
					var js map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
						errs = append(errs, fmt.Errorf("%q: JSON inválido: %s", key, err))
					}
					return
				},
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da atribuição",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da atribuição",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceAppCatalogAssignmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Processar configuração (string JSON para map)
	var config map[string]interface{}
	if configStr, ok := d.GetOk("configuration"); ok && configStr.(string) != "" {
		if err := json.Unmarshal([]byte(configStr.(string)), &config); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar configuração: %v", err))
		}
	}

	// Construir atribuição
	assignment := &AppCatalogAssignment{
		ApplicationID:  d.Get("application_id").(string),
		TargetType:     d.Get("target_type").(string),
		TargetID:       d.Get("target_id").(string),
		AssignmentType: d.Get("assignment_type").(string),
		InstallPolicy:  d.Get("install_policy").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		assignment.OrgID = v.(string)
	}

	// Adicionar configuração se definida
	if config != nil {
		assignment.Configuration = config
	}

	// Serializar para JSON
	assignmentJSON, err := json.Marshal(assignment)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar atribuição: %v", err))
	}

	// Criar atribuição via API
	tflog.Debug(ctx, "Criando atribuição de aplicação")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/appcatalog/assignments", assignmentJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar atribuição de aplicação: %v", err))
	}

	// Deserializar resposta
	var createdAssignment AppCatalogAssignment
	if err := json.Unmarshal(resp, &createdAssignment); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	if createdAssignment.ID == "" {
		return diag.FromErr(fmt.Errorf("atribuição criada sem ID"))
	}

	d.SetId(createdAssignment.ID)
	return resourceAppCatalogAssignmentRead(ctx, d, m)
}

func resourceAppCatalogAssignmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da atribuição não fornecido"))
	}

	// Buscar atribuição via API
	tflog.Debug(ctx, fmt.Sprintf("Lendo atribuição de aplicação com ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/appcatalog/assignments/%s", id), nil)
	if err != nil {
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Atribuição %s não encontrada, removendo do state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler atribuição de aplicação: %v", err))
	}

	// Deserializar resposta
	var assignment AppCatalogAssignment
	if err := json.Unmarshal(resp, &assignment); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir valores no state
	d.Set("application_id", assignment.ApplicationID)
	d.Set("target_type", assignment.TargetType)
	d.Set("target_id", assignment.TargetID)
	d.Set("assignment_type", assignment.AssignmentType)
	d.Set("install_policy", assignment.InstallPolicy)
	d.Set("created", assignment.Created)
	d.Set("updated", assignment.Updated)

	// Converter mapa de configuração para JSON se existir
	if assignment.Configuration != nil {
		configJSON, err := json.Marshal(assignment.Configuration)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar configuração: %v", err))
		}
		d.Set("configuration", string(configJSON))
	} else {
		d.Set("configuration", "")
	}

	if assignment.OrgID != "" {
		d.Set("org_id", assignment.OrgID)
	}

	return diags
}

func resourceAppCatalogAssignmentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da atribuição não fornecido"))
	}

	// Processar configuração (string JSON para map)
	var config map[string]interface{}
	if configStr, ok := d.GetOk("configuration"); ok && configStr.(string) != "" {
		if err := json.Unmarshal([]byte(configStr.(string)), &config); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar configuração: %v", err))
		}
	}

	// Construir atribuição atualizada
	assignment := &AppCatalogAssignment{
		ID:             id,
		ApplicationID:  d.Get("application_id").(string),
		TargetType:     d.Get("target_type").(string),
		TargetID:       d.Get("target_id").(string),
		AssignmentType: d.Get("assignment_type").(string),
		InstallPolicy:  d.Get("install_policy").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("org_id"); ok {
		assignment.OrgID = v.(string)
	}

	// Adicionar configuração se definida
	if config != nil {
		assignment.Configuration = config
	}

	// Serializar para JSON
	assignmentJSON, err := json.Marshal(assignment)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar atribuição: %v", err))
	}

	// Atualizar atribuição via API
	tflog.Debug(ctx, fmt.Sprintf("Atualizando atribuição de aplicação: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/appcatalog/assignments/%s", id), assignmentJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar atribuição de aplicação: %v", err))
	}

	// Deserializar resposta
	var updatedAssignment AppCatalogAssignment
	if err := json.Unmarshal(resp, &updatedAssignment); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	return resourceAppCatalogAssignmentRead(ctx, d, m)
}

func resourceAppCatalogAssignmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("ID da atribuição não fornecido"))
	}

	// Excluir atribuição via API
	tflog.Debug(ctx, fmt.Sprintf("Excluindo atribuição de aplicação: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/appcatalog/assignments/%s", id), nil)
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if isNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Atribuição %s não encontrada, considerando excluída", id))
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("erro ao excluir atribuição de aplicação: %v", err))
	}

	// Remover do state
	d.SetId("")
	return diag.Diagnostics{}
}
