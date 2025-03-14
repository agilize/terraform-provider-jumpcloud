# Configuração do Provider JumpCloud
# É recomendado usar variáveis de ambiente ou um módulo de secrets para as credenciais
provider "jumpcloud" {
  api_key = var.jumpcloud_api_key  # JUMPCLOUD_API_KEY
  org_id  = var.jumpcloud_org_id   # JUMPCLOUD_ORG_ID
}

# Variáveis de uso comum para padronização
locals {
  environment = "production"
  region      = "us-east"
  common_tags = ["managed-by-terraform", local.environment]
}

#######################
# INFRAESTRUTURA WEB
#######################

# Servidor Web com configuração padrão de produção
resource "jumpcloud_system" "web_server" {
  display_name                      = "web-server-01"
  allow_ssh_root_login              = false                     # Desativa login como root por SSH para maior segurança
  allow_ssh_password_authentication = true                      # Permite autenticação por senha
  allow_multi_factor_authentication = true                      # Ativa MFA para acesso ao sistema
  description                       = "Servidor web NGINX para o ambiente de produção"
  
  # Tags para organização e aplicação de políticas
  tags = concat(local.common_tags, [
    "web",
    "nginx",
    "public-facing"
  ])
  
  # Atributos personalizados para metadados e automação
  attributes = {
    environment = local.environment
    region      = local.region
    role        = "web"
    tier        = "frontend"
    backup      = "daily"
    patching    = "weekly"
  }
}

# Exemplo de criação de múltiplos servidores web com contador
resource "jumpcloud_system" "web_cluster" {
  count = 3
  
  display_name                      = "web-server-${count.index + 2}"  # Iniciando de 2 para não conflitar com web-server-01
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = true
  allow_multi_factor_authentication = true
  description                       = "Servidor web ${count.index + 2} do cluster de produção"
  
  tags = concat(local.common_tags, [
    "web",
    "nginx",
    "cluster"
  ])
  
  attributes = {
    environment = local.environment
    region      = local.region
    role        = "web"
    cluster     = "primary"
    node        = "${count.index + 2}"
  }
}

#######################
# INFRAESTRUTURA DB
#######################

# Servidor de banco de dados com alta segurança
resource "jumpcloud_system" "db_server" {
  display_name                      = "db-server-01"
  allow_ssh_root_login              = false                      # Desativa login como root por SSH
  allow_ssh_password_authentication = false                      # Desativa autenticação por senha, apenas chaves SSH
  allow_multi_factor_authentication = true                       # Ativa MFA para maior segurança
  description                       = "Servidor de banco de dados principal MySQL"
  
  # Tags para organização e políticas de segurança específicas
  tags = concat(local.common_tags, [
    "database",
    "mysql",
    "restricted-access",
    "pci-dss"
  ])
  
  # Atributos personalizados
  attributes = {
    environment    = local.environment
    region         = local.region
    role           = "database"
    tier           = "data"
    backup         = "hourly"
    patching       = "monthly"
    encryption     = "enabled"
    compliance     = "pci-dss,hipaa"
    data_class     = "confidential"
  }
  
  # Configurações específicas adicionais
  ssh_root_enabled = false
  agent_bound      = true
}

#######################
# OUTPUTS ÚTEIS
#######################

# Exportar IDs dos sistemas criados para referência
output "web_server_id" {
  description = "ID do servidor web principal"
  value       = jumpcloud_system.web_server.id
}

output "db_server_id" {
  description = "ID do servidor de banco de dados"
  value       = jumpcloud_system.db_server.id
}

output "web_cluster_ids" {
  description = "IDs dos servidores web no cluster"
  value       = jumpcloud_system.web_cluster[*].id
} 