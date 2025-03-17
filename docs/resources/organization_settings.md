# jumpcloud_organization_settings Resource

Este recurso permite gerenciar as configurações de uma organização no JumpCloud, incluindo políticas de senha, configurações de MFA, insights, templates de e-mail e outras configurações de segurança.

## Exemplo de Uso

```hcl
# Configurações básicas de organização com política de senha personalizada
resource "jumpcloud_organization_settings" "main_org" {
  org_id = var.jumpcloud_org_id
  
  password_policy {
    min_length            = 12
    requires_lowercase    = true
    requires_uppercase    = true
    requires_number       = true
    requires_special_char = true
    expiration_days       = 90
    max_history           = 10
  }
  
  # Configurar MFA
  allow_multi_factor_auth = true
  require_mfa             = true
  allowed_mfa_methods     = ["totp", "push", "webauthn"]
  
  # Configurações adicionais
  system_insights_enabled     = true
  directory_insights_enabled  = true
  ldap_integration_enabled    = false
  allow_public_key_authentication = true
}

# Configurações de organização com templates de e-mail personalizados
resource "jumpcloud_organization_settings" "custom_emails" {
  org_id = var.child_org_id
  
  # Templates de e-mail personalizados
  new_user_email_template  = file("${path.module}/templates/new_user_email.html")
  password_reset_template  = file("${path.module}/templates/password_reset.html")
  
  # Configurações de segurança básicas
  password_policy {
    min_length      = 10
    expiration_days = 60
  }
  
  # Desabilitar insights para economizar custos
  system_insights_enabled    = false
  directory_insights_enabled = false
}
```

## Argument Reference

Os seguintes argumentos são suportados:

* `org_id` - (Obrigatório) O ID da organização para a qual as configurações serão aplicadas.

### Política de Senha

* `password_policy` - (Opcional) Bloco de configuração da política de senha. Pode conter:
  * `min_length` - (Opcional) Comprimento mínimo da senha. Deve estar entre 8 e 64 caracteres. Padrão: `8`.
  * `requires_lowercase` - (Opcional) Exigir letras minúsculas. Padrão: `true`.
  * `requires_uppercase` - (Opcional) Exigir letras maiúsculas. Padrão: `true`.
  * `requires_number` - (Opcional) Exigir números. Padrão: `true`.
  * `requires_special_char` - (Opcional) Exigir caracteres especiais. Padrão: `true`.
  * `expiration_days` - (Opcional) Dias até a expiração da senha. `0` significa que a senha nunca expira. Deve estar entre 0 e 365. Padrão: `90`.
  * `max_history` - (Opcional) Número de senhas antigas a serem lembradas. Deve estar entre 0 e 24. Padrão: `5`.

### Configurações de Autenticação

* `allow_multi_factor_auth` - (Opcional) Permitir autenticação multifator. Padrão: `true`.
* `require_mfa` - (Opcional) Exigir MFA para todos os usuários. Padrão: `false`.
* `allowed_mfa_methods` - (Opcional) Lista de métodos MFA permitidos na organização. Valores válidos: `totp`, `duo`, `push`, `sms`, `email`, `webauthn`, `security_questions`.
* `allow_public_key_authentication` - (Opcional) Permitir autenticação por chave pública SSH. Padrão: `true`.

### Configurações de Insights e Integração

* `system_insights_enabled` - (Opcional) Habilitar System Insights. Padrão: `true`.
* `directory_insights_enabled` - (Opcional) Habilitar Directory Insights. Padrão: `true`.
* `ldap_integration_enabled` - (Opcional) Habilitar integração LDAP. Padrão: `false`.

### Configurações de Sistema e Usuário

* `new_system_user_state_managed` - (Opcional) Se o estado de usuários em novos sistemas é gerenciado pelo JumpCloud. Padrão: `true`.
* `new_user_email_template` - (Opcional) Template HTML para e-mails de novos usuários.
* `password_reset_template` - (Opcional) Template HTML para e-mails de redefinição de senha.

## Attribute Reference

Além de todos os argumentos acima, os seguintes atributos são exportados:

* `id` - O ID das configurações da organização.
* `created` - A data de criação das configurações.
* `updated` - A data da última atualização das configurações.

## Import

As configurações de organização podem ser importadas usando o ID da organização, por exemplo:

```
$ terraform import jumpcloud_organization_settings.main_org 5f1b1bb2c9e9a9b7e8d6c5a4
```

## Exemplos Avançados

### Configuração Completa com Integração Multi-Recurso

```hcl
# Configuração da organização subsidiária
resource "jumpcloud_organization" "subsidiaria" {
  name           = "Subsidiária Brasil"
  display_name   = "Acme Brasil Ltda."
  parent_org_id  = var.parent_organization_id
  contact_name   = "Gerente de TI"
  contact_email  = "ti@acmebrasil.exemplo.com"
  website        = "https://brasil.acme.exemplo.com"
  
  # Domínios permitidos para esta organização
  allowed_domains = [
    "acmebrasil.exemplo.com",
    "acme-br.exemplo.com"
  ]
}

# Configuração detalhada de segurança para a organização
resource "jumpcloud_organization_settings" "subsidiaria_settings" {
  org_id = jumpcloud_organization.subsidiaria.id
  
  # Configuração de política de senha robusta
  password_policy {
    min_length            = 14
    requires_lowercase    = true
    requires_uppercase    = true
    requires_number       = true
    requires_special_char = true
    expiration_days       = 90
    max_history           = 24  # Não permite reusar as últimas 24 senhas
  }
  
  # Configuração de MFA obrigatório com métodos permitidos
  allow_multi_factor_auth = true
  require_mfa             = true
  allowed_mfa_methods     = ["totp", "push", "webauthn"]
  
  # Ativar monitoramento e insights
  system_insights_enabled    = true
  directory_insights_enabled = true
  
  # Gerenciamento automático de status de usuário para sistemas
  new_system_user_state_managed = true
  
  # Permitir autenticação por chave pública (SSH)
  allow_public_key_authentication = true
}

# Configuração de webhook para eventos de segurança
resource "jumpcloud_webhook" "security_alerts" {
  name        = "Alertas de Segurança"
  url         = "https://siem.acme.exemplo.com/api/jumpcloud"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook para alertas de segurança da subsidiária Brasil"
  
  event_types = [
    "user.login.failed",
    "user.mfa.disabled",
    "system.disconnected",
    "organization.settings.updated"
  ]
}

# Assinatura de evento específica para atualizações de configuração
resource "jumpcloud_webhook_subscription" "settings_change" {
  webhook_id  = jumpcloud_webhook.security_alerts.id
  event_type  = "organization.settings.updated"
  description = "Monitorar alterações nas configurações de segurança"
}

# Criar API key para automação
resource "jumpcloud_api_key" "automation" {
  name        = "Automação Brasil"
  description = "API Key para automação de tarefas na subsidiária Brasil"
  expires     = timeadd(timestamp(), "8760h") # Expira em 1 ano
}

# Configurar permissões para a API key (somente leitura)
resource "jumpcloud_api_key_binding" "read_only" {
  api_key_id    = jumpcloud_api_key.automation.id
  resource_type = "organization"
  resource_ids  = [jumpcloud_organization.subsidiaria.id]
  permissions   = ["read"]
}

# Outputs importantes
output "org_id" {
  value = jumpcloud_organization.subsidiaria.id
  description = "ID da organização subsidiária no JumpCloud"
}

output "api_key" {
  value = jumpcloud_api_key.automation.key
  description = "API key para automação (mostrado apenas durante a criação)"
  sensitive = true
}
```

### Configuração para Conformidade com Requisitos de Segurança

```hcl
resource "jumpcloud_organization_settings" "compliance_settings" {
  org_id = var.org_id
  
  # Política de senha conforme requisitos de conformidade
  password_policy {
    min_length            = 16
    requires_lowercase    = true
    requires_uppercase    = true
    requires_number       = true
    requires_special_char = true
    expiration_days       = 60
    max_history           = 24
  }
  
  # MFA obrigatório para todos os usuários
  allow_multi_factor_auth = true
  require_mfa             = true
  
  # Restringir apenas a métodos de MFA mais seguros
  # Não permitindo SMS que é mais vulnerável a ataques
  allowed_mfa_methods     = ["totp", "webauthn"]
  
  # Ativar todos os insights e monitoramento
  system_insights_enabled    = true
  directory_insights_enabled = true
  
  # Gerenciamento rigoroso de usuários e autenticação
  new_system_user_state_managed = true
  allow_public_key_authentication = true
  
  # Template personalizado para redefinição de senha
  password_reset_template = <<-EOT
Prezado(a) {{user.firstname}},

Uma redefinição de senha foi solicitada para a sua conta na Acme Brasil.
Por motivos de segurança e conformidade, sua nova senha deve ter:
- No mínimo 16 caracteres
- Letras maiúsculas e minúsculas
- Números
- Caracteres especiais

A senha expirará em 60 dias, conforme nossa política de segurança.

Atenciosamente,
Equipe de Segurança da Informação
Acme Brasil
  EOT
}
``` 