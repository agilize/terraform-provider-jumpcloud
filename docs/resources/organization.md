# jumpcloud_organization Resource

Gerencia organizações no JumpCloud. Este recurso permite criar e gerenciar organizações em um ambiente multi-tenant, configurando detalhes como nome, contato, branding e domínios permitidos.

## Exemplo de Uso

### Organização Subsidiária Básica
```hcl
# Criar uma organização subsidiária
resource "jumpcloud_organization" "subsidiary" {
  name           = "Subsidiary Corp"
  display_name   = "Subsidiary Corporation"
  parent_org_id  = var.parent_organization_id
  
  # Informações de contato
  contact_name   = "John Doe"
  contact_email  = "john.doe@subsidiary.com"
  contact_phone  = "+1 555-0123"
  
  # Detalhes da organização
  website        = "https://www.subsidiary.com"
  logo_url      = "https://assets.subsidiary.com/logo.png"
  
  # Domínios permitidos
  allowed_domains = [
    "subsidiary.com",
    "sub.subsidiary.com"
  ]
}

# Configurar as configurações da organização
resource "jumpcloud_organization_settings" "subsidiary_settings" {
  org_id = jumpcloud_organization.subsidiary.id
  
  # Configurações de senha
  password_policy = {
    min_length            = 12
    min_numeric          = 1
    min_uppercase        = 1
    min_lowercase        = 1
    min_special          = 1
    max_attempts         = 5
    lockout_time_seconds = 300
  }
  
  # Configurações de MFA
  require_mfa             = true
  allow_multi_factor_auth = true
  
  # Outras configurações
  system_insights_enabled = true
  retention_days         = 90
  timezone              = "America/New_York"
}

# Exportar o ID da organização subsidiária
output "subsidiary_org_id" {
  value = jumpcloud_organization.subsidiary.id
}
```

### Organização com Configurações Avançadas
```hcl
# Criar uma organização com configurações avançadas
resource "jumpcloud_organization" "enterprise" {
  name           = "Enterprise Division"
  display_name   = "Enterprise Solutions Division"
  parent_org_id  = var.parent_organization_id
  
  # Informações de contato
  contact_name   = "Jane Smith"
  contact_email  = "jane.smith@enterprise.com"
  contact_phone  = "+1 555-4567"
  
  # Detalhes da organização
  website        = "https://enterprise.example.com"
  logo_url      = "https://assets.enterprise.com/logo.png"
  
  # Domínios permitidos com subdomínios
  allowed_domains = [
    "enterprise.com",
    "*.enterprise.com",
    "enterprise.example.com"
  ]
}

# Configurar configurações avançadas
resource "jumpcloud_organization_settings" "enterprise_settings" {
  org_id = jumpcloud_organization.enterprise.id
  
  # Política de senha rigorosa
  password_policy = {
    min_length            = 16
    min_numeric          = 2
    min_uppercase        = 2
    min_lowercase        = 2
    min_special          = 2
    max_attempts         = 3
    lockout_time_seconds = 600
    prevent_reuse        = true
    expire_days         = 90
  }
  
  # Segurança avançada
  require_mfa                = true
  allow_multi_factor_auth    = true
  allow_public_key_auth      = true
  allow_ssh_root_login      = false
  
  # Configurações de sistema
  system_insights_enabled    = true
  retention_days            = 180
  timezone                 = "UTC"
  
  # Templates de email personalizados
  email_templates = {
    welcome = {
      subject = "Bem-vindo à Enterprise Division"
      body    = file("${path.module}/templates/welcome.html")
    }
    password_reset = {
      subject = "Redefinição de Senha Solicitada"
      body    = file("${path.module}/templates/password_reset.html")
    }
  }
}

# Criar uma chave de API para a organização
resource "jumpcloud_api_key" "enterprise_api" {
  name        = "Enterprise API Key"
  description = "Chave de API para automação da Enterprise Division"
  expires     = timeadd(timestamp(), "8760h") # Expira em 1 ano
}

# Configurar permissões da chave de API
resource "jumpcloud_api_key_binding" "enterprise_api_access" {
  api_key_id    = jumpcloud_api_key.enterprise_api.id
  resource_type = "organization"
  permissions   = ["read", "list", "update"]
}

# Exportar informações da organização
output "enterprise_info" {
  value = {
    org_id   = jumpcloud_organization.enterprise.id
    api_key  = jumpcloud_api_key.enterprise_api.key
    domains  = jumpcloud_organization.enterprise.allowed_domains
  }
  sensitive = true
}
```

### Organização para Ambiente de Desenvolvimento
```hcl
# Criar uma organização para desenvolvimento
resource "jumpcloud_organization" "dev" {
  name           = "Development"
  display_name   = "Development Environment"
  parent_org_id  = var.parent_organization_id
  
  # Informações de contato
  contact_name   = "Dev Team Lead"
  contact_email  = "devteam@example.com"
  contact_phone  = "+1 555-7890"
  
  # Detalhes da organização
  website        = "https://dev.example.com"
  logo_url      = "https://assets.example.com/dev-logo.png"
  
  # Domínios permitidos para desenvolvimento
  allowed_domains = [
    "dev.example.com",
    "test.example.com",
    "staging.example.com"
  ]
}

# Configurar configurações menos restritivas para desenvolvimento
resource "jumpcloud_organization_settings" "dev_settings" {
  org_id = jumpcloud_organization.dev.id
  
  # Política de senha para desenvolvimento
  password_policy = {
    min_length            = 8
    min_numeric          = 1
    min_uppercase        = 1
    min_lowercase        = 1
    min_special          = 0
    max_attempts         = 10
    lockout_time_seconds = 300
  }
  
  # Configurações de desenvolvimento
  require_mfa             = false
  allow_multi_factor_auth = true
  system_insights_enabled = true
  retention_days         = 30
  timezone              = "UTC"
}

# Criar webhook para notificações de eventos
resource "jumpcloud_webhook" "dev_events" {
  name        = "Dev Environment Events"
  url         = "https://dev-monitor.example.com/events"
  enabled     = true
  description = "Webhook para monitoramento do ambiente de desenvolvimento"
}

# Configurar assinaturas do webhook
resource "jumpcloud_webhook_subscription" "dev_user_events" {
  webhook_id   = jumpcloud_webhook.dev_events.id
  event_type   = "user.created"
  description  = "Monitorar criação de usuários no ambiente de desenvolvimento"
}

# Exportar configuração do ambiente de desenvolvimento
output "dev_environment" {
  value = {
    org_id    = jumpcloud_organization.dev.id
    webhook_id = jumpcloud_webhook.dev_events.id
    domains   = jumpcloud_organization.dev.allowed_domains
  }
}
```

## Argumentos

Os seguintes argumentos são suportados:

* `name` - (Obrigatório) Nome da organização. Deve ser único dentro do tenant pai.
* `display_name` - (Opcional) Nome de exibição da organização.
* `parent_org_id` - (Obrigatório) ID da organização pai.
* `contact_name` - (Opcional) Nome do contato principal da organização.
* `contact_email` - (Opcional) Email do contato principal.
* `contact_phone` - (Opcional) Telefone do contato principal.
* `website` - (Opcional) Website da organização.
* `logo_url` - (Opcional) URL do logo da organização.
* `allowed_domains` - (Opcional) Lista de domínios permitidos para usuários da organização.

## Atributos Exportados

Além dos argumentos acima, os seguintes atributos são exportados:

* `id` - ID único da organização.
* `created` - Data de criação da organização no formato ISO 8601.
* `updated` - Data da última atualização da organização no formato ISO 8601.

## Importação

Organizações podem ser importadas usando seu ID:

```shell
terraform import jumpcloud_organization.subsidiary j1_org_1234567890
```

## Notas de Uso

### Hierarquia de Organizações

1. Uma organização deve ter exatamente uma organização pai.
2. A hierarquia de organizações não pode ser alterada após a criação.
3. A exclusão de uma organização pai não é permitida se houver organizações filhas.

### Domínios Permitidos

1. Use `*.domain.com` para permitir todos os subdomínios.
2. Domínios devem ser únicos entre organizações irmãs.
3. Subdomínios específicos têm precedência sobre wildcards.

### Boas Práticas

1. Use nomes descritivos e consistentes.
2. Mantenha a documentação de contato atualizada.
3. Revise regularmente os domínios permitidos.
4. Configure webhooks para monitorar eventos importantes.

### Exemplo de Validação de Domínio

```python
from typing import List
import re

def validate_domain_pattern(domain: str) -> bool:
    """
    Valida se um padrão de domínio é válido.
    
    Args:
        domain: Padrão de domínio a ser validado
        
    Returns:
        bool: True se o padrão é válido
    """
    if domain.startswith('*.'):
        domain = domain[2:]
    
    pattern = r'^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$'
    return bool(re.match(pattern, domain, re.IGNORECASE))

def validate_allowed_domains(
    domains: List[str],
    existing_domains: List[str]
) -> List[str]:
    """
    Valida uma lista de domínios permitidos.
    
    Args:
        domains: Lista de domínios a serem validados
        existing_domains: Lista de domínios já em uso
        
    Returns:
        List[str]: Lista de erros encontrados
    """
    errors = []
    
    for domain in domains:
        if not validate_domain_pattern(domain):
            errors.append(f"Invalid domain pattern: {domain}")
        
        if domain in existing_domains:
            errors.append(f"Domain already in use: {domain}")
            
        if domain.startswith('*.'):
            base_domain = domain[2:]
            for existing in existing_domains:
                if existing.endswith(base_domain) and existing != domain:
                    errors.append(
                        f"Conflict with existing domain {existing}"
                    )
    
    return errors
```

### Exemplo de Auditoria de Organizações

```python
from datetime import datetime
from typing import Dict, List

def audit_organization_hierarchy(
    organizations: List[Dict]
) -> Dict[str, List[str]]:
    """
    Audita a hierarquia de organizações.
    
    Args:
        organizations: Lista de organizações com seus metadados
        
    Returns:
        Dict[str, List[str]]: Relatório de problemas encontrados
    """
    issues = {}
    org_map = {org['id']: org for org in organizations}
    
    for org in organizations:
        org_issues = []
        
        # Verificar organização pai
        if org.get('parent_org_id'):
            parent = org_map.get(org['parent_org_id'])
            if not parent:
                org_issues.append("Parent organization not found")
        
        # Verificar contatos
        if not org.get('contact_email'):
            org_issues.append("Missing contact email")
        
        # Verificar domínios
        domains = org.get('allowed_domains', [])
        domain_errors = validate_allowed_domains(
            domains,
            [d for o in organizations if o['id'] != org['id']
             for d in o.get('allowed_domains', [])]
        )
        org_issues.extend(domain_errors)
        
        if org_issues:
            issues[org['id']] = org_issues
    
    return issues
``` 