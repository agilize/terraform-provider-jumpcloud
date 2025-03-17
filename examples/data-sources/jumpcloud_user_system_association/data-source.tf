provider "jumpcloud" {
  # É recomendado usar variáveis de ambiente para as credenciais:
  # - JUMPCLOUD_API_KEY
  # - JUMPCLOUD_ORG_ID
}

# Buscar dados de usuário e sistema existentes
data "jumpcloud_user" "existing_user" {
  email = "existing.user@example.com"
}

data "jumpcloud_system" "existing_system" {
  display_name = "existing-server"
}

# Verificar se o usuário tem acesso ao sistema
data "jumpcloud_user_system_association" "check_access" {
  user_id   = data.jumpcloud_user.existing_user.id
  system_id = data.jumpcloud_system.existing_system.id
}

# Uso simples para output informativo
output "access_status" {
  description = "Status de acesso do usuário ao sistema"
  value = "${data.jumpcloud_user_system_association.check_access.associated ? 
    "O usuário ${data.jumpcloud_user.existing_user.username} tem acesso ao sistema ${data.jumpcloud_system.existing_system.display_name}" : 
    "O usuário ${data.jumpcloud_user.existing_user.username} NÃO tem acesso ao sistema ${data.jumpcloud_system.existing_system.display_name}"}"
}

# Uso avançado para criar associação condicional
locals {
  needs_access = !data.jumpcloud_user_system_association.check_access.associated
}

# Criar a associação apenas se ela não existir
resource "jumpcloud_user_system_association" "conditional_access" {
  count = local.needs_access ? 1 : 0
  
  user_id   = data.jumpcloud_user.existing_user.id
  system_id = data.jumpcloud_system.existing_system.id
}

# Verificar acesso de vários usuários a um sistema
data "jumpcloud_user" "admin_users" {
  count = 3
  email = "admin.user${count.index + 1}@example.com"
}

data "jumpcloud_system" "critical_system" {
  display_name = "critical-database-server"
}

data "jumpcloud_user_system_association" "admin_access_checks" {
  count = length(data.jumpcloud_user.admin_users)
  
  user_id   = data.jumpcloud_user.admin_users[count.index].id
  system_id = data.jumpcloud_system.critical_system.id
}

# Gerar relatório de acesso
output "admin_access_report" {
  description = "Relatório de acesso de administradores"
  value = [
    for i, check in data.jumpcloud_user_system_association.admin_access_checks:
    {
      username    = data.jumpcloud_user.admin_users[i].username
      has_access  = check.associated
      user_id     = check.user_id
      system_name = data.jumpcloud_system.critical_system.display_name
    }
  ]
} 