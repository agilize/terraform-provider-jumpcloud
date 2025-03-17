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

// SoftwareDeployment representa uma implantação de software no JumpCloud
type SoftwareDeployment struct {
	ID          string                 `json:"_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	PackageID   string                 `json:"packageId"`
	TargetType  string                 `json:"targetType"` // system, system_group, user, user_group
	TargetIDs   []string               `json:"targetIds"`
	Schedule    map[string]interface{} `json:"schedule,omitempty"`
	Status      string                 `json:"status,omitempty"`   // scheduled, in_progress, completed, failed, canceled
	Progress    int                    `json:"progress,omitempty"` // percentual de conclusão (0-100)
	StartTime   string                 `json:"startTime,omitempty"`
	EndTime     string                 `json:"endTime,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"` // parâmetros específicos para esta implantação
	OrgID       string                 `json:"orgId,omitempty"`
	Created     string                 `json:"created,omitempty"`
	Updated     string                 `json:"updated,omitempty"`
}

func resourceSoftwareDeployment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSoftwareDeploymentCreate,
		ReadContext:   resourceSoftwareDeploymentRead,
		UpdateContext: resourceSoftwareDeploymentUpdate,
		DeleteContext: resourceSoftwareDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Nome da implantação de software",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Descrição da implantação de software",
			},
			"package_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID do pacote de software a ser implantado",
			},
			"target_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"system", "system_group", "user", "user_group",
				}, false),
				Description: "Tipo de destino para a implantação (system, system_group, user, user_group)",
			},
			"target_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "IDs dos destinos para implantação (sistemas, grupos, usuários)",
			},
			"schedule": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
				Description:      "Configuração de agendamento em formato JSON (imediato se não especificado)",
			},
			"parameters": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
				Description:      "Parâmetros específicos para esta implantação em formato JSON",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status atual da implantação (scheduled, in_progress, completed, failed, canceled)",
			},
			"progress": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Percentual de conclusão da implantação (0-100)",
			},
			"start_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data e hora de início da implantação",
			},
			"end_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data e hora de término da implantação",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data de criação da implantação",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data da última atualização da implantação",
			},
		},
	}
}

func resourceSoftwareDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construir objeto SoftwareDeployment a partir dos dados do terraform
	deployment := &SoftwareDeployment{
		Name:       d.Get("name").(string),
		PackageID:  d.Get("package_id").(string),
		TargetType: d.Get("target_type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		deployment.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		deployment.OrgID = v.(string)
	}

	// Processar IDs de destino
	if v, ok := d.GetOk("target_ids"); ok {
		targetSet := v.(*schema.Set)
		targetIDs := make([]string, targetSet.Len())
		for i, id := range targetSet.List() {
			targetIDs[i] = id.(string)
		}
		deployment.TargetIDs = targetIDs
	}

	// Processar agendamento (JSON)
	if v, ok := d.GetOk("schedule"); ok {
		scheduleJSON := v.(string)
		var schedule map[string]interface{}
		if err := json.Unmarshal([]byte(scheduleJSON), &schedule); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar agendamento: %v", err))
		}
		deployment.Schedule = schedule
	}

	// Processar parâmetros (JSON)
	if v, ok := d.GetOk("parameters"); ok {
		parametersJSON := v.(string)
		var parameters map[string]interface{}
		if err := json.Unmarshal([]byte(parametersJSON), &parameters); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar parâmetros: %v", err))
		}
		deployment.Parameters = parameters
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(deployment)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar implantação de software: %v", err))
	}

	// Construir URL para requisição
	url := "/api/v2/software/deployments"
	if deployment.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, deployment.OrgID)
	}

	// Fazer requisição para criar implantação
	tflog.Debug(ctx, "Criando implantação de software")
	resp, err := c.DoRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao criar implantação de software: %v", err))
	}

	// Deserializar resposta
	var createdDeployment SoftwareDeployment
	if err := json.Unmarshal(resp, &createdDeployment); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID no state
	d.SetId(createdDeployment.ID)

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceSoftwareDeploymentRead(ctx, d, meta)
}

func resourceSoftwareDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID da implantação
	deploymentID := d.Id()

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/software/deployments/%s%s", deploymentID, orgIDParam)

	// Fazer requisição para ler implantação
	tflog.Debug(ctx, fmt.Sprintf("Lendo implantação de software: %s", deploymentID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		// Se o recurso não for encontrado, remover do state
		if err.Error() == "Status Code: 404" {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao ler implantação de software: %v", err))
	}

	// Deserializar resposta
	var deployment SoftwareDeployment
	if err := json.Unmarshal(resp, &deployment); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Mapear valores para o schema
	d.Set("name", deployment.Name)
	d.Set("description", deployment.Description)
	d.Set("package_id", deployment.PackageID)
	d.Set("target_type", deployment.TargetType)
	d.Set("status", deployment.Status)
	d.Set("progress", deployment.Progress)
	d.Set("start_time", deployment.StartTime)
	d.Set("end_time", deployment.EndTime)
	d.Set("created", deployment.Created)
	d.Set("updated", deployment.Updated)

	// Definir IDs de destino
	if err := d.Set("target_ids", deployment.TargetIDs); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao definir target_ids: %v", err))
	}

	// Serializar agendamento para JSON
	if deployment.Schedule != nil {
		scheduleJSON, err := json.Marshal(deployment.Schedule)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar agendamento: %v", err))
		}
		d.Set("schedule", string(scheduleJSON))
	}

	// Serializar parâmetros para JSON
	if deployment.Parameters != nil {
		parametersJSON, err := json.Marshal(deployment.Parameters)
		if err != nil {
			return diag.FromErr(fmt.Errorf("erro ao serializar parâmetros: %v", err))
		}
		d.Set("parameters", string(parametersJSON))
	}

	// Definir OrgID se presente
	if deployment.OrgID != "" {
		d.Set("org_id", deployment.OrgID)
	}

	return diags
}

func resourceSoftwareDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID da implantação
	deploymentID := d.Id()

	// Verificar se a implantação pode ser atualizada com base no status atual
	// Ler o status atual da implantação
	currentStatus, err := getDeploymentStatus(ctx, c, deploymentID, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao verificar status da implantação: %v", err))
	}

	// Se a implantação já estiver em andamento ou concluída, não permitir atualização
	if currentStatus == "in_progress" || currentStatus == "completed" {
		return diag.FromErr(fmt.Errorf("não é possível atualizar uma implantação com status '%s'", currentStatus))
	}

	// Construir objeto SoftwareDeployment a partir dos dados do terraform
	deployment := &SoftwareDeployment{
		ID:         deploymentID,
		Name:       d.Get("name").(string),
		PackageID:  d.Get("package_id").(string),
		TargetType: d.Get("target_type").(string),
	}

	// Campos opcionais
	if v, ok := d.GetOk("description"); ok {
		deployment.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		deployment.OrgID = v.(string)
	}

	// Processar IDs de destino
	if v, ok := d.GetOk("target_ids"); ok {
		targetSet := v.(*schema.Set)
		targetIDs := make([]string, targetSet.Len())
		for i, id := range targetSet.List() {
			targetIDs[i] = id.(string)
		}
		deployment.TargetIDs = targetIDs
	}

	// Processar agendamento (JSON)
	if v, ok := d.GetOk("schedule"); ok {
		scheduleJSON := v.(string)
		var schedule map[string]interface{}
		if err := json.Unmarshal([]byte(scheduleJSON), &schedule); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar agendamento: %v", err))
		}
		deployment.Schedule = schedule
	}

	// Processar parâmetros (JSON)
	if v, ok := d.GetOk("parameters"); ok {
		parametersJSON := v.(string)
		var parameters map[string]interface{}
		if err := json.Unmarshal([]byte(parametersJSON), &parameters); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao deserializar parâmetros: %v", err))
		}
		deployment.Parameters = parameters
	}

	// Serializar para JSON
	reqBody, err := json.Marshal(deployment)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao serializar implantação de software: %v", err))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/software/deployments/%s", deploymentID)
	if deployment.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, deployment.OrgID)
	}

	// Fazer requisição para atualizar implantação
	tflog.Debug(ctx, fmt.Sprintf("Atualizando implantação de software: %s", deploymentID))
	_, err = c.DoRequest(http.MethodPut, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao atualizar implantação de software: %v", err))
	}

	// Ler o recurso para atualizar o state com todos os campos computados
	return resourceSoftwareDeploymentRead(ctx, d, meta)
}

func resourceSoftwareDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID da implantação
	deploymentID := d.Id()

	// Verificar se a implantação pode ser excluída com base no status atual
	// Ler o status atual da implantação
	currentStatus, err := getDeploymentStatus(ctx, c, deploymentID, d.Get("org_id").(string))
	if err != nil {
		// Se o recurso não for encontrado, consideramos que já foi excluído
		if err.Error() == "Status Code: 404" {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("erro ao verificar status da implantação: %v", err))
	}

	// Se a implantação estiver em andamento, cancelá-la antes de excluir
	if currentStatus == "in_progress" {
		if err := cancelDeployment(ctx, c, deploymentID, d.Get("org_id").(string)); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao cancelar implantação em andamento: %v", err))
		}
	}

	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/software/deployments/%s%s", deploymentID, orgIDParam)

	// Fazer requisição para excluir implantação
	tflog.Debug(ctx, fmt.Sprintf("Excluindo implantação de software: %s", deploymentID))
	_, err = c.DoRequest(http.MethodDelete, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao excluir implantação de software: %v", err))
	}

	// Remover ID do state
	d.SetId("")

	return diags
}

// getDeploymentStatus obtém o status atual da implantação
func getDeploymentStatus(ctx context.Context, c ClientInterface, deploymentID, orgID string) (string, error) {
	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if orgID != "" {
		orgIDParam = fmt.Sprintf("?orgId=%s", orgID)
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/software/deployments/%s%s", deploymentID, orgIDParam)

	// Fazer requisição para ler implantação
	tflog.Debug(ctx, fmt.Sprintf("Verificando status da implantação: %s", deploymentID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	// Deserializar resposta
	var deployment struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(resp, &deployment); err != nil {
		return "", fmt.Errorf("erro ao deserializar resposta: %v", err)
	}

	return deployment.Status, nil
}

// cancelDeployment cancela uma implantação em andamento
func cancelDeployment(ctx context.Context, c ClientInterface, deploymentID, orgID string) error {
	// Obter parâmetro orgId se disponível
	var orgIDParam string
	if orgID != "" {
		orgIDParam = fmt.Sprintf("?orgId=%s", orgID)
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/software/deployments/%s/cancel%s", deploymentID, orgIDParam)

	// Fazer requisição para cancelar implantação
	tflog.Debug(ctx, fmt.Sprintf("Cancelando implantação: %s", deploymentID))
	_, err := c.DoRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	return nil
}
