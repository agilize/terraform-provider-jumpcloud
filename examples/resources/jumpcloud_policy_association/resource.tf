# Configure the JumpCloud Provider
provider "jumpcloud" {
  api_key = var.jumpcloud_api_key # ou use variáveis de ambiente: JUMPCLOUD_API_KEY
  org_id  = var.jumpcloud_org_id  # ou use: JUMPCLOUD_ORG_ID
}

# Criar um grupo de usuários para associar à política
resource "jumpcloud_user_group" "finance" {
  name        = "Finance Department"
  description = "Grupo de usuários do departamento financeiro"
}

# Criar uma política de complexidade de senha
resource "jumpcloud_policy" "password_complexity" {
  name        = "Secure Password Policy"
  description = "Política de complexidade de senha para o departamento financeiro"
  type        = "password_complexity"
  active      = true
  
  configurations = {
    min_length             = "12"
    requires_uppercase     = "true"
    requires_lowercase     = "true"
    requires_number        = "true"
    requires_special_char  = "true"
    password_expires_days  = "90"
    enable_password_expiry = "true"
  }
}

# Associar a política ao grupo de usuários
resource "jumpcloud_policy_association" "finance_password_policy" {
  policy_id = jumpcloud_policy.password_complexity.id
  group_id  = jumpcloud_user_group.finance.id
  type      = "user_group"
}

# Criar um grupo de sistemas para associar à política
resource "jumpcloud_system_group" "servers" {
  name        = "Production Servers"
  description = "Grupo de servidores de produção"
}

# Criar uma política de atualizações de sistema
resource "jumpcloud_policy" "system_updates" {
  name        = "System Updates Policy"
  description = "Política para controle de atualizações de sistema"
  type        = "system_updates"
  active      = true
  
  configurations = {
    auto_update_enabled = "true"
    auto_update_time    = "02:00"
  }
}

# Associar a política ao grupo de sistemas
resource "jumpcloud_policy_association" "servers_update_policy" {
  policy_id = jumpcloud_policy.system_updates.id
  group_id  = jumpcloud_system_group.servers.id
  type      = "system_group"
}

# Política de MFA para todos os colaboradores
resource "jumpcloud_policy" "mfa_policy" {
  name        = "Required MFA Policy"
  description = "Política global de MFA para todos os usuários"
  type        = "mfa"
  active      = true
  
  configurations = {
    allow_totp_enrollment      = "true"
    require_mfa_for_all_users  = "true"
  }
}

# Associar política de MFA a vários grupos
resource "jumpcloud_user_group" "it" {
  name = "IT Department"
}

resource "jumpcloud_user_group" "executives" {
  name = "Executive Team"
}

resource "jumpcloud_policy_association" "it_mfa" {
  policy_id = jumpcloud_policy.mfa_policy.id
  group_id  = jumpcloud_user_group.it.id
  type      = "user_group"
}

resource "jumpcloud_policy_association" "executives_mfa" {
  policy_id = jumpcloud_policy.mfa_policy.id
  group_id  = jumpcloud_user_group.executives.id
  type      = "user_group"
}

resource "jumpcloud_policy_association" "finance_mfa" {
  policy_id = jumpcloud_policy.mfa_policy.id
  group_id  = jumpcloud_user_group.finance.id
  type      = "user_group"
}

# Output das associações
output "finance_password_policy_association" {
  value = jumpcloud_policy_association.finance_password_policy.id
}

output "servers_update_policy_association" {
  value = jumpcloud_policy_association.servers_update_policy.id
}

output "mfa_policy_associations" {
  value = [
    jumpcloud_policy_association.it_mfa.id,
    jumpcloud_policy_association.executives_mfa.id,
    jumpcloud_policy_association.finance_mfa.id
  ]
} 