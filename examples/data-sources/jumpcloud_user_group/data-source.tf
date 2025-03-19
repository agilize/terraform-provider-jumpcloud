provider "jumpcloud" {
  # É recomendado usar variáveis de ambiente para as credenciais:
  # - JUMPCLOUD_API_KEY
  # - JUMPCLOUD_ORG_ID
}

# Buscar um grupo de usuários por nome
data "jumpcloud_user_group" "by_name" {
  name = "dev-team"
}

# Buscar um grupo de usuários por ID
data "jumpcloud_user_group" "by_id" {
  id = "5f0c1b2c3d4e5f6g7h8i9j0k" # Substitua pelo ID real do grupo
}

# Uso da informação do grupo para output
output "dev_group_details" {
  description = "Detalhes do grupo de desenvolvedores"
  value = {
    id          = data.jumpcloud_user_group.by_name.id
    name        = data.jumpcloud_user_group.by_name.name
    description = data.jumpcloud_user_group.by_name.description
    attributes  = data.jumpcloud_user_group.by_name.attributes
  }
}

# Uso com condicionais
locals {
  is_admin_group = contains(keys(data.jumpcloud_user_group.by_name.attributes), "access_level") && data.jumpcloud_user_group.by_name.attributes["access_level"] == "Admin"
}

output "group_access_level" {
  description = "Nível de acesso do grupo"
  value = local.is_admin_group ? "Grupo de administradores" : "Grupo regular"
}

# Uso em conjunto com recursos
resource "jumpcloud_user" "new_team_member" {
  username    = "new.developer"
  email       = "new.developer@example.com"
  firstname   = "New"
  lastname    = "Developer"
  password    = "SecurePassword123!"
  description = "Novo membro da equipe de ${data.jumpcloud_user_group.by_name.name}"
  
  attributes = {
    group_id    = data.jumpcloud_user_group.by_name.id
    department  = lookup(data.jumpcloud_user_group.by_name.attributes, "department", "Undefined")
  }
} 