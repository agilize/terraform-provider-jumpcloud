# Configuração do Provider JumpCloud
provider "jumpcloud" {
  # É recomendado usar variáveis de ambiente para as credenciais:
  # - JUMPCLOUD_API_KEY
  # - JUMPCLOUD_ORG_ID
}

#####################################
# EXEMPLOS DE BUSCA POR NOME
#####################################

# Recuperar informações de um sistema específico pelo nome
data "jumpcloud_system" "web_server" {
  display_name = "web-server-01"
}

# Uso das informações recuperadas
output "web_server_details" {
  description = "Detalhes do servidor web"
  value = {
    id           = data.jumpcloud_system.web_server.id
    display_name = data.jumpcloud_system.web_server.display_name
    os           = data.jumpcloud_system.web_server.os
    created      = data.jumpcloud_system.web_server.created
    tags         = data.jumpcloud_system.web_server.tags
  }
}

#####################################
# EXEMPLOS DE BUSCA POR ID
#####################################

# Recuperar informações de um sistema pelo ID do JumpCloud
# Útil para referências estáveis quando o nome pode mudar
data "jumpcloud_system" "db_server" {
  id = "5f8a7b6c5d4e3f2a1b0c9d8e" # Substitua pelo ID real do sistema
}

# Verificação dos atributos avançados do sistema
output "db_server_security_config" {
  description = "Configurações de segurança do servidor de banco de dados"
  value = {
    allow_ssh_root_login        = data.jumpcloud_system.db_server.allow_ssh_root_login
    allow_password_auth         = data.jumpcloud_system.db_server.allow_ssh_password_authentication
    allow_mfa                   = data.jumpcloud_system.db_server.allow_multi_factor_authentication
    ssh_root_enabled            = data.jumpcloud_system.db_server.ssh_root_enabled
    has_active_directory        = data.jumpcloud_system.db_server.has_active_directory
  }
}

#####################################
# INTEGRAÇÃO COM OUTROS RECURSOS
#####################################

# Exemplo de uso do data source para integração com outros recursos
# Útil para referenciar sistemas existentes em novas configurações

# 1. Buscar um sistema existente
data "jumpcloud_system" "existing_app_server" {
  display_name = "app-server-prod"
}

# 2. Buscar um usuário existente
data "jumpcloud_user" "admin_user" {
  email = "admin@example.com"
}

# 3. Criar uma associação entre o usuário e o sistema
# (Assumindo que existe um recurso para essa associação)
resource "jumpcloud_user_system_association" "admin_access" {
  user_id   = data.jumpcloud_user.admin_user.id
  system_id = data.jumpcloud_system.existing_app_server.id
}

#####################################
# FILTRAGEM E CONDICIONAIS
#####################################

# Uso condicional baseado nos atributos do sistema
locals {
  # Determinar se o sistema precisa de atualização de segurança
  needs_security_update = (
    data.jumpcloud_system.web_server.allow_ssh_root_login == true || 
    data.jumpcloud_system.web_server.allow_ssh_password_authentication == true
  )
  
  # Construir lista de tags baseada em atributos
  server_category = contains(data.jumpcloud_system.web_server.tags, "production") ? "prod" : "non-prod"
}

# Saída condicional baseada na análise
output "security_recommendations" {
  description = "Recomendações de segurança baseadas na configuração atual"
  value = local.needs_security_update ? [
    "Desativar login SSH como root",
    "Implementar autenticação baseada apenas em chaves SSH",
    "Ativar MFA para todo acesso ao sistema"
  ] : ["Configurações de segurança adequadas"]
}

# Agrupar sistemas por ambiente usando outputs
output "environment_classification" {
  description = "Classificação do servidor por ambiente"
  value = local.server_category
}

#####################################
# METADATA E MONITORING
#####################################

# Usar dados do sistema para integração com monitoramento
output "monitoring_config" {
  description = "Configuração para ferramentas de monitoramento"
  value = {
    system_id     = data.jumpcloud_system.web_server.id
    hostname      = data.jumpcloud_system.web_server.hostname
    display_name  = data.jumpcloud_system.web_server.display_name
    os_family     = data.jumpcloud_system.web_server.os
    created_date  = data.jumpcloud_system.web_server.created
    environment   = local.server_category
    
    # Atributos personalizados para monitoramento
    monitoring_group = contains(data.jumpcloud_system.web_server.tags, "critical") ? "high-priority" : "standard"
    check_interval   = contains(data.jumpcloud_system.web_server.tags, "critical") ? "30s" : "5m"
  }
} 