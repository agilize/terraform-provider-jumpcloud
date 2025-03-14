# Configuração de Segurança Avançada para JumpCloud Systems
# Este exemplo demonstra práticas recomendadas de segurança para sistemas críticos

provider "jumpcloud" {
  # Recomendado usar variáveis de ambiente para credenciais
  # JUMPCLOUD_API_KEY e JUMPCLOUD_ORG_ID
}

#################################################
# EXEMPLO DE SERVIDOR COM SEGURANÇA APRIMORADA
#################################################

resource "jumpcloud_system" "secure_server" {
  display_name                      = "secure-finance-server"
  description                       = "Servidor financeiro com configurações de segurança avançadas"
  
  # Configurações de segurança recomendadas
  allow_ssh_root_login              = false    # Desativar login SSH como root
  allow_ssh_password_authentication = false    # Usar apenas autenticação por chave SSH
  allow_multi_factor_authentication = true     # Ativar MFA para todos os acessos
  
  # Tags para controle de acesso e políticas
  tags = [
    "high-security", 
    "finance", 
    "pci-dss", 
    "soc2-compliance",
    "critical-data"
  ]
  
  # Metadata com informações de segurança
  attributes = {
    # Classificação do sistema
    security_tier            = "tier-1"
    data_classification      = "confidential"
    compliance_requirements  = "pci-dss,soc2,gdpr,hipaa"
    
    # Controles de segurança implementados
    encryption_at_rest       = "enabled"
    encryption_in_transit    = "enabled"
    vulnerability_scanning   = "daily"
    backup_frequency         = "hourly"
    
    # Informações operacionais
    patch_schedule           = "weekly-sunday-02:00"
    logging_level            = "enhanced"
    incident_response_team   = "security-team-alpha"
    auth_methods_allowed     = "pubkey,mfa"
    
    # Responsáveis
    security_contact         = "security@example.com"
    owner                    = "financial-department"
  }
  
  # Garantir que estamos usando as configurações mais seguras
  ssh_root_enabled           = false
  agent_bound                = true
}

#################################################
# EXEMPLO DE IMPLEMENTAÇÃO DE BASTION HOST
#################################################

resource "jumpcloud_system" "bastion_host" {
  display_name                      = "secure-bastion-host"
  description                       = "Host bastion para acesso seguro à infraestrutura interna"
  
  # Configurações de segurança extremas para host de salto
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = false
  allow_multi_factor_authentication = true
  
  tags = [
    "bastion",
    "gateway",
    "perimeter-security",
    "access-control"
  ]
  
  attributes = {
    purpose                  = "secure-gateway"
    network_zone             = "dmz"
    allowed_source_ips       = "office-ips,vpn-endpoints"
    session_recording        = "enabled"
    session_timeout          = "15m"
    idle_timeout             = "5m"
    audit_logging            = "verbose"
    ssh_key_rotation         = "30d"
    allowed_ports            = "22,443"
    firewall_profile         = "strict"
  }
}

#################################################
# OUTPUTS PARA INTEGRAÇÃO COM OUTRAS FERRAMENTAS
#################################################

output "secure_system_id" {
  description = "ID do servidor seguro"
  value       = jumpcloud_system.secure_server.id
  
  # Marcar como sensitive para não exibir em logs
  sensitive   = true
}

output "bastion_id" {
  description = "ID do host bastion"
  value       = jumpcloud_system.bastion_host.id
  
  # Marcar como sensitive para não exibir em logs
  sensitive   = true
} 