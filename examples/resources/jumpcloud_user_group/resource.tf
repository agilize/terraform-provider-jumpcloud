provider "jumpcloud" {
  # É recomendado usar variáveis de ambiente para as credenciais:
  # - JUMPCLOUD_API_KEY
  # - JUMPCLOUD_ORG_ID
}

# Exemplo básico de grupo de usuários
resource "jumpcloud_user_group" "developers" {
  name        = "dev-team"
  description = "Grupo para a equipe de desenvolvimento"
  
  attributes = {
    department = "Engineering"
    location   = "Remote"
    role       = "Developer"
  }
}

# Grupo com configurações mais detalhadas
resource "jumpcloud_user_group" "it_admins" {
  name        = "it-administrators"
  description = "Administradores de TI com acesso privilegiado"
  
  attributes = {
    department      = "IT"
    access_level    = "Admin"
    requires_mfa    = "true"
    on_call_rotation = "yes"
    security_level  = "high"
  }
}

# Exemplo de grupo para contratados temporários
resource "jumpcloud_user_group" "contractors" {
  name        = "temporary-contractors"
  description = "Grupo para contratados temporários com acesso limitado"
  
  attributes = {
    department     = "Various"
    access_level   = "Limited"
    contract_end   = "2024-12-31"
    requires_mfa   = "true"
    data_access    = "restricted"
    security_level = "medium"
  }
}

# Outputs para referência
output "developers_group_id" {
  description = "ID do grupo de desenvolvedores"
  value       = jumpcloud_user_group.developers.id
}

output "admin_group_id" {
  description = "ID do grupo de administradores"
  value       = jumpcloud_user_group.it_admins.id
} 