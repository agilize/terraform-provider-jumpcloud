package provider

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
)

// DeploymentTargetStatus representa o status de um destino específico na implantação
type DeploymentTargetStatus struct {
	TargetID     string `json:"targetId"`
	TargetName   string `json:"targetName"`
	TargetType   string `json:"targetType"` // system, user
	Status       string `json:"status"`     // pending, in_progress, completed, failed, canceled
	ErrorMessage string `json:"errorMessage,omitempty"`
	StartTime    string `json:"startTime,omitempty"`
	EndTime      string `json:"endTime,omitempty"`
}

// DeploymentStatus representa o status detalhado de uma implantação
type DeploymentStatus struct {
	ID                string                   `json:"_id"`
	Name              string                   `json:"name"`
	Status            string                   `json:"status"` // scheduled, in_progress, completed, failed, canceled
	Progress          int                      `json:"progress"`
	StartTime         string                   `json:"startTime,omitempty"`
	EndTime           string                   `json:"endTime,omitempty"`
	TargetStatuses    []DeploymentTargetStatus `json:"targetStatuses,omitempty"`
	PackageID         string                   `json:"packageId"`
	PackageName       string                   `json:"packageName"`
	PackageVersion    string                   `json:"packageVersion"`
	TotalTargets      int                      `json:"totalTargets"`
	SuccessTargets    int                      `json:"successTargets"`
	FailedTargets     int                      `json:"failedTargets"`
	PendingTargets    int                      `json:"pendingTargets"`
	InProgressTargets int                      `json:"inProgressTargets"`
	CanceledTargets   int                      `json:"canceledTargets"`
}

func dataSourceSoftwareDeploymentStatus() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSoftwareDeploymentStatusRead,
		Schema: map[string]*schema.Schema{
			"deployment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "ID da implantação de software",
			},
			"include_target_details": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Inclui detalhes de status para cada destino da implantação",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID da organização para ambientes multi-tenant",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Nome da implantação",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status geral da implantação (scheduled, in_progress, completed, failed, canceled)",
			},
			"progress": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Percentual de conclusão da implantação (0-100)",
			},
			"package_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID do pacote de software sendo implantado",
			},
			"package_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Nome do pacote de software sendo implantado",
			},
			"package_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Versão do pacote de software sendo implantado",
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
			"total_targets": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número total de destinos na implantação",
			},
			"success_targets": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número de destinos com implantação bem-sucedida",
			},
			"failed_targets": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número de destinos com falha na implantação",
			},
			"pending_targets": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número de destinos pendentes de implantação",
			},
			"in_progress_targets": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número de destinos com implantação em andamento",
			},
			"canceled_targets": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Número de destinos com implantação cancelada",
			},
			"target_statuses": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Detalhes de status para cada destino (se include_target_details=true)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID do destino",
						},
						"target_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Nome do destino",
						},
						"target_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tipo do destino (system, user)",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status da implantação para este destino",
						},
						"error_message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Mensagem de erro (se houver)",
						},
						"start_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data e hora de início para este destino",
						},
						"end_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Data e hora de término para este destino",
						},
					},
				},
			},
		},
	}
}

func dataSourceSoftwareDeploymentStatusRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(m)
	if diagErr != nil {
		return diagErr
	}

	// Obter ID da implantação
	deploymentID := d.Get("deployment_id").(string)
	includeTargetDetails := d.Get("include_target_details").(bool)

	// Obter parâmetros para a consulta
	params := "?include=package"
	if includeTargetDetails {
		params += "&include=targets"
	}

	// Adicionar orgId se disponível
	if v, ok := d.GetOk("org_id"); ok {
		params += fmt.Sprintf("&orgId=%s", v.(string))
	}

	// Construir URL para requisição
	url := fmt.Sprintf("/api/v2/software/deployments/%s/status%s", deploymentID, params)

	// Fazer requisição para obter status
	tflog.Debug(ctx, fmt.Sprintf("Consultando status da implantação: %s", deploymentID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("erro ao consultar status da implantação: %v", err))
	}

	// Deserializar resposta
	var status DeploymentStatus
	if err := json.Unmarshal(resp, &status); err != nil {
		return diag.FromErr(fmt.Errorf("erro ao deserializar resposta: %v", err))
	}

	// Definir ID do data source
	d.SetId(fmt.Sprintf("%s-%d", deploymentID, time.Now().Unix()))

	// Mapear valores para o schema
	d.Set("name", status.Name)
	d.Set("status", status.Status)
	d.Set("progress", status.Progress)
	d.Set("package_id", status.PackageID)
	d.Set("package_name", status.PackageName)
	d.Set("package_version", status.PackageVersion)
	d.Set("start_time", status.StartTime)
	d.Set("end_time", status.EndTime)
	d.Set("total_targets", status.TotalTargets)
	d.Set("success_targets", status.SuccessTargets)
	d.Set("failed_targets", status.FailedTargets)
	d.Set("pending_targets", status.PendingTargets)
	d.Set("in_progress_targets", status.InProgressTargets)
	d.Set("canceled_targets", status.CanceledTargets)

	// Processar detalhes de status de destinos, se incluídos
	if includeTargetDetails && status.TargetStatuses != nil {
		targetStatuses := make([]map[string]interface{}, len(status.TargetStatuses))
		for i, ts := range status.TargetStatuses {
			targetStatus := map[string]interface{}{
				"target_id":     ts.TargetID,
				"target_name":   ts.TargetName,
				"target_type":   ts.TargetType,
				"status":        ts.Status,
				"error_message": ts.ErrorMessage,
				"start_time":    ts.StartTime,
				"end_time":      ts.EndTime,
			}
			targetStatuses[i] = targetStatus
		}
		if err := d.Set("target_statuses", targetStatuses); err != nil {
			return diag.FromErr(fmt.Errorf("erro ao definir target_statuses: %v", err))
		}
	}

	return diags
}
