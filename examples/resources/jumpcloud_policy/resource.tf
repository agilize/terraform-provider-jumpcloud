# Configure the JumpCloud Provider
provider "jumpcloud" {
  api_key = var.jumpcloud_api_key # ou use variáveis de ambiente: JUMPCLOUD_API_KEY
  org_id  = var.jumpcloud_org_id  # ou use: JUMPCLOUD_ORG_ID
}

# Exemplo de política de complexidade de senha
resource "jumpcloud_policy" "password_complexity" {
  name        = "Secure Password Policy"
  description = "Política de complexidade de senha segura para todos os usuários"
  type        = "password_complexity"
  active      = true
  
  # Configurações específicas para política de complexidade de senha
  configurations = {
    min_length             = "12"         # Comprimento mínimo da senha
    requires_uppercase     = "true"       # Requer letras maiúsculas
    requires_lowercase     = "true"       # Requer letras minúsculas
    requires_number        = "true"       # Requer números
    requires_special_char  = "true"       # Requer caracteres especiais
    password_expires_days  = "90"         # Senha expira em 90 dias
    enable_password_expiry = "true"       # Ativar expiração de senha
  }
}

# Exemplo de política de MFA (Autenticação Multi-Fator)
resource "jumpcloud_policy" "mfa_policy" {
  name        = "Required MFA Policy"
  description = "Política que requer MFA para todos os usuários"
  type        = "mfa"
  active      = true
  
  configurations = {
    allow_sms_enrollment            = "true"   # Permitir MFA via SMS
    allow_voice_call_enrollment     = "true"   # Permitir MFA via chamada de voz
    allow_totp_enrollment           = "true"   # Permitir MFA via aplicativo autenticador (TOTP)
    allow_push_notification         = "true"   # Permitir MFA via notificação push
    require_mfa_for_all_users       = "true"   # Exigir MFA para todos os usuários
    days_before_password_expiration = "7"      # Lembrete 7 dias antes da expiração
  }
}

# Exemplo de política de bloqueio de conta após tentativas malsucedidas
resource "jumpcloud_policy" "account_lockout" {
  name        = "Security Lockout Policy"
  description = "Política para bloquear contas após várias tentativas de login fracassadas"
  type        = "account_lockout_timeout"
  active      = true
  
  configurations = {
    max_failed_login_attempts    = "5"      # Máximo de tentativas
    lockout_time_in_minutes      = "30"     # Tempo de bloqueio em minutos
    reset_counter_after_minutes  = "10"     # Resetar contador após período de inatividade
  }
}

# Associando políticas a grupos de usuários (requer recursos adicionais)
/*
resource "jumpcloud_user_group" "finance" {
  name = "Finance Department"
}

resource "jumpcloud_policy_association" "finance_password_policy" {
  policy_id  = jumpcloud_policy.password_complexity.id
  group_id   = jumpcloud_user_group.finance.id
  type       = "user_group"
}
*/

# Output das políticas criadas
output "password_policy_id" {
  value       = jumpcloud_policy.password_complexity.id
  description = "ID da política de senha"
}

output "mfa_policy_id" {
  value       = jumpcloud_policy.mfa_policy.id
  description = "ID da política de MFA"
}

output "account_lockout_policy_id" {
  value       = jumpcloud_policy.account_lockout.id
  description = "ID da política de bloqueio de conta"
} 