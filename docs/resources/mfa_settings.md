# jumpcloud_mfa_settings Resource

Este recurso permite gerenciar as configurações de MFA (Multi-Factor Authentication) no JumpCloud. Como estas configurações são definidas por organização, este é um recurso singleton - apenas uma instância deve existir por organização JumpCloud.

## Exemplo de Uso

### Configuração básica de MFA

```hcl
resource "jumpcloud_mfa_settings" "corporate_mfa" {
  # Habilitar MFA baseado em System Insights
  system_insights_enrolled = true
  
  # Configurar dias de janela de exclusão (período de graça)
  exclusion_window_days = 7
  
  # Métodos MFA permitidos
  enabled_methods = [
    "totp",      # Time-based One-Time Password (Google Authenticator, etc.)
    "push",      # Notificações push
    "webauthn"   # FIDO2/WebAuthn (Yubikey, etc.)
  ]
}
```

### Configuração de MFA em ambiente multi-tenant

```hcl
resource "jumpcloud_mfa_settings" "child_org_mfa" {
  # ID da organização específica (para implementações multi-tenant)
  organization_id = var.child_organization_id
  
  system_insights_enrolled = true
  
  # Configuração mais restritiva - apenas TOTP
  enabled_methods = ["totp"]
  
  # Sem janela de exclusão - aplicação imediata
  exclusion_window_days = 0
}
```

## Argument Reference

Os seguintes argumentos são suportados:

* `system_insights_enrolled` - (Opcional) Se o System Insights está habilitado para MFA. Padrão: `false`.
* `exclusion_window_days` - (Opcional) Número de dias de janela de exclusão para MFA. Esta é uma "janela de graça" durante a qual os usuários podem acessar sem MFA após a configuração ser ativada. Valores de 0 a 30, onde 0 significa aplicação imediata. Padrão: `0`.
* `enabled_methods` - (Opcional) Lista de métodos MFA habilitados. Valores válidos: `totp`, `duo`, `push`, `sms`, `email`, `webauthn`, `security_questions`.
* `organization_id` - (Opcional) ID da organização para implementações multi-tenant. Se não especificado, será usado o ID da organização atual configurada no provider.

## Attribute Reference

Além dos argumentos listados acima, os seguintes atributos são exportados:

* `id` - ID das configurações de MFA ou "current" para a organização atual.
* `updated` - Data da última atualização das configurações de MFA.

## Import

Configurações de MFA JumpCloud podem ser importadas usando o ID da organização ou "current" se houver apenas uma:

```
terraform import jumpcloud_mfa_settings.example {organization_id}
```

ou

```
terraform import jumpcloud_mfa_settings.example current
```

## Notas de Implementação

Este recurso gerencia um singleton por organização JumpCloud. O comportamento de exclusão (`terraform destroy`) redefine as configurações de MFA para os valores padrão do JumpCloud em vez de excluí-los completamente, pois as configurações de MFA sempre existem para cada organização. 