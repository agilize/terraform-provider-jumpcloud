provider "jumpcloud" {
  # É recomendado usar variáveis de ambiente para as credenciais:
  # - JUMPCLOUD_API_KEY
  # - JUMPCLOUD_ORG_ID
}

# Criar um usuário para associação
resource "jumpcloud_user" "example_user" {
  username    = "example.user"
  email       = "example.user@example.com"
  firstname   = "Example"
  lastname    = "User"
  password    = "SecurePassword123!"
  description = "Usuário de exemplo criado pelo Terraform"
}

# Sistema com o qual o usuário será associado
resource "jumpcloud_system" "example_system" {
  display_name = "example-server"
  description  = "Servidor de exemplo para demonstração"
  
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = false
  allow_multi_factor_authentication = true
  
  tags = ["example", "terraform-managed"]
}

# Associar o usuário ao sistema
resource "jumpcloud_user_system_association" "example_association" {
  user_id   = jumpcloud_user.example_user.id
  system_id = jumpcloud_system.example_system.id
}

# Exemplo de associação de múltiplos usuários a um sistema crítico
# Primeiro, criamos vários usuários
resource "jumpcloud_user" "admin_team" {
  count = 3
  
  username    = "admin.user${count.index + 1}"
  email       = "admin.user${count.index + 1}@example.com"
  firstname   = "Admin"
  lastname    = "User ${count.index + 1}"
  password    = "SecurePassword${count.index + 1}!"
  description = "Administrador ${count.index + 1}"
  
  attributes = {
    role = "administrator"
  }
}

# Em seguida, criamos o sistema crítico
resource "jumpcloud_system" "critical_system" {
  display_name = "critical-database-server"
  description  = "Servidor de banco de dados crítico com acesso restrito"
  
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = false
  allow_multi_factor_authentication = true
  
  tags = ["critical", "database", "restricted-access"]
  
  attributes = {
    security_level = "high"
    backup         = "hourly"
    monitoring     = "enhanced"
  }
}

# Associar cada administrador ao sistema crítico
resource "jumpcloud_user_system_association" "admin_access" {
  count = length(jumpcloud_user.admin_team)
  
  user_id   = jumpcloud_user.admin_team[count.index].id
  system_id = jumpcloud_system.critical_system.id
}

# Saída para referência
output "association_id" {
  description = "ID da associação entre usuário e sistema"
  value       = jumpcloud_user_system_association.example_association.id
}

output "admin_associations" {
  description = "IDs das associações de administradores"
  value       = jumpcloud_user_system_association.admin_access[*].id
} 