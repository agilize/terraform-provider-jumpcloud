# jumpcloud_mfa_settings Data Source

Este data source permite obter informações sobre as configurações de MFA (Multi-Factor Authentication) para uma organização JumpCloud. Pode ser útil para avaliar as políticas de MFA existentes antes de fazer alterações ou para monitorar configurações entre várias organizações.

## Exemplo de Uso

### Buscar configurações da organização atual

```hcl
data "jumpcloud_mfa_settings" "current" {}

output "mfa_methods_enabled" {
  value = data.jumpcloud_mfa_settings.current.enabled_methods
}

output "system_insights_status" {
  value = data.jumpcloud_mfa_settings.current.system_insights_enrolled ? "Ativado" : "Desativado"
}
```

### Buscar configurações de uma organização específica (multi-tenant)

```hcl
data "jumpcloud_mfa_settings" "child_org" {
  organization_id = var.child_organization_id
}

# Validação de configurações para conformidade
output "mfa_exclusion_window_compliant" {
  value = data.jumpcloud_mfa_settings.child_org.exclusion_window_days <= 7
  description = "Conforme se a janela de exclusão for menor ou igual a 7 dias"
}
```

## Argument Reference

Os seguintes argumentos são suportados:

* `organization_id` - (Opcional) ID da organização para obter as configurações de MFA. Se não especificado, será usado o ID da organização atual configurada no provider.

## Attribute Reference

Os seguintes atributos são exportados:

* `id` - ID das configurações de MFA ou "current" para a organização atual.
* `system_insights_enrolled` - Se o System Insights está habilitado para MFA.
* `exclusion_window_days` - Número de dias de janela de exclusão para MFA (período de graça).
* `enabled_methods` - Lista de métodos MFA habilitados, como `totp`, `duo`, `push`, `sms`, `email`, `webauthn` e `security_questions`.
* `updated` - Data da última atualização das configurações de MFA. 