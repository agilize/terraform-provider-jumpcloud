package provider

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Expandir uma lista de strings para testes de MDM
func expandStringListForMDMTests(list []interface{}) []string {
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}

// Mock para a função de criação do resource MDM Devices
func resourceMDMDevicesCreateTest(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)

	// Preparar os dados para a requisição
	payload := map[string]interface{}{
		"name":           d.Get("name"),
		"deviceIds":      expandStringListForMDMTests(d.Get("device_ids").([]interface{})),
		"deviceGroupIds": expandStringListForMDMTests(d.Get("device_group_ids").([]interface{})),
		"status":         d.Get("status"),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Fazer a requisição para criar os dispositivos MDM
	resp, err := client.DoRequest(http.MethodPost, "/api/v2/mdm/devices", payloadBytes)
	if err != nil {
		return diag.FromErr(err)
	}

	// Decodificar a resposta
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	// Definir o ID do recurso
	d.SetId(result["_id"].(string))

	// Ler o recurso para atualizar o state
	return resourceMDMDevicesReadTest(ctx, d, m)
}

// Mock para a função de leitura do resource MDM Devices
func resourceMDMDevicesReadTest(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)
	var diags diag.Diagnostics

	// Fazer a requisição para obter os dispositivos MDM
	resp, err := client.DoRequest(http.MethodGet, "/api/v2/mdm/devices/"+d.Id(), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	// Decodificar a resposta
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	// Atualizar o state
	if err := d.Set("name", result["name"]); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	// Converter os device_ids e device_group_ids para o formato correto
	if deviceIds, ok := result["deviceIds"].([]interface{}); ok {
		if err := d.Set("device_ids", deviceIds); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if deviceGroupIds, ok := result["deviceGroupIds"].([]interface{}); ok {
		if err := d.Set("device_group_ids", deviceGroupIds); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if err := d.Set("status", result["status"]); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

// Mock para a função de atualização do resource MDM Devices
func resourceMDMDevicesUpdateTest(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(JumpCloudClient)

	// Preparar os dados para a requisição
	payload := map[string]interface{}{
		"name":           d.Get("name"),
		"deviceIds":      expandStringListForMDMTests(d.Get("device_ids").([]interface{})),
		"deviceGroupIds": expandStringListForMDMTests(d.Get("device_group_ids").([]interface{})),
		"status":         d.Get("status"),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// Fazer a requisição para atualizar os dispositivos MDM
	_, err = client.DoRequest(http.MethodPut, "/api/v2/mdm/devices/"+d.Id(), payloadBytes)
	if err != nil {
		return diag.FromErr(err)
	}

	// Ler o recurso para atualizar o state
	return resourceMDMDevicesReadTest(ctx, d, m)
}
