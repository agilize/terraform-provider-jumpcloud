# jumpcloud_api_key_binding Resource

Gerencia permissões de chaves de API no JumpCloud. Este recurso permite definir quais operações uma chave de API pode realizar e em quais recursos, permitindo um controle granular de acesso.

## Exemplo de Uso

### Automação de Usuários
```hcl
# Criar uma chave de API para automação
resource "jumpcloud_api_key" "user_automation" {
  name        = "User Automation API Key"
  description = "Chave de API para automação de usuários"
}

# Configurar permissões para gerenciar todos os usuários
resource "jumpcloud_api_key_binding" "user_management" {
  api_key_id    = jumpcloud_api_key.user_automation.id
  resource_type = "user"
  permissions   = ["read", "list", "create", "update", "delete"]
}

# Configurar permissões para gerenciar grupos de usuários
resource "jumpcloud_api_key_binding" "user_group_management" {
  api_key_id    = jumpcloud_api_key.user_automation.id
  resource_type = "user_group"
  permissions   = ["read", "list", "create", "update", "delete"]
}
```

### Monitoramento de Sistemas
```hcl
# Criar uma chave de API para monitoramento
resource "jumpcloud_api_key" "system_monitor" {
  name        = "System Monitor API Key"
  description = "Chave de API para monitoramento de sistemas"
}

# Configurar permissões de leitura para sistemas específicos
resource "jumpcloud_api_key_binding" "system_monitoring" {
  api_key_id    = jumpcloud_api_key.system_monitor.id
  resource_type = "system"
  permissions   = ["read", "list"]
  resource_ids  = ["sys_123", "sys_456", "sys_789"]
}

# Configurar permissões para monitorar grupos de sistemas
resource "jumpcloud_api_key_binding" "system_group_monitoring" {
  api_key_id    = jumpcloud_api_key.system_monitor.id
  resource_type = "system_group"
  permissions   = ["read", "list"]
}
```

### Gerenciamento de Aplicações
```hcl
# Criar uma chave de API para gerenciamento de aplicações
resource "jumpcloud_api_key" "app_management" {
  name        = "Application Management API Key"
  description = "Chave de API para gerenciamento de aplicações"
}

# Configurar permissões para gerenciar aplicações
resource "jumpcloud_api_key_binding" "application_management" {
  api_key_id    = jumpcloud_api_key.app_management.id
  resource_type = "application"
  permissions   = ["read", "list", "create", "update", "delete"]
}

# Configurar permissões para gerenciar associações de usuários
resource "jumpcloud_api_key_binding" "application_user_binding" {
  api_key_id    = jumpcloud_api_key.app_management.id
  resource_type = "application_user"
  permissions   = ["read", "list", "create", "delete"]
}
```

### Monitoramento de Eventos
```hcl
# Criar uma chave de API para monitoramento de eventos
resource "jumpcloud_api_key" "event_monitor" {
  name        = "Event Monitor API Key"
  description = "Chave de API para monitoramento de eventos"
}

# Configurar permissões para monitorar eventos de autenticação
resource "jumpcloud_api_key_binding" "auth_event_monitoring" {
  api_key_id    = jumpcloud_api_key.event_monitor.id
  resource_type = "auth_event"
  permissions   = ["read", "list"]
}

# Configurar permissões para monitorar eventos de diretório
resource "jumpcloud_api_key_binding" "directory_event_monitoring" {
  api_key_id    = jumpcloud_api_key.event_monitor.id
  resource_type = "directory_event"
  permissions   = ["read", "list"]
}
```

## Argumentos

Os seguintes argumentos são suportados:

* `api_key_id` - (Obrigatório) ID da chave de API à qual este binding se aplica.
* `resource_type` - (Obrigatório) Tipo de recurso ao qual o binding se aplica. Valores válidos incluem:
  * `user` - Usuários
  * `user_group` - Grupos de usuários
  * `system` - Sistemas
  * `system_group` - Grupos de sistemas
  * `application` - Aplicações
  * `application_user` - Associações de usuários a aplicações
  * `policy` - Políticas
  * `command` - Comandos
  * `auth_event` - Eventos de autenticação
  * `directory_event` - Eventos de diretório
  * `webhook` - Webhooks
  * `organization` - Organizações
* `permissions` - (Obrigatório) Lista de permissões concedidas à chave de API para o tipo de recurso especificado. Valores válidos incluem:
  * `read` - Permissão para ler recursos
  * `list` - Permissão para listar recursos
  * `create` - Permissão para criar recursos
  * `update` - Permissão para atualizar recursos
  * `delete` - Permissão para excluir recursos
* `resource_ids` - (Opcional) Lista de IDs específicos de recursos aos quais as permissões se aplicam. Se omitido, as permissões se aplicam a todos os recursos do tipo especificado.

## Atributos Exportados

Além dos argumentos acima, os seguintes atributos são exportados:

* `id` - ID único do binding da chave de API.
* `created` - Data de criação do binding no formato ISO 8601.
* `updated` - Data da última atualização do binding no formato ISO 8601.

## Importação

Bindings de chave de API podem ser importados usando seu ID:

```shell
terraform import jumpcloud_api_key_binding.user_management j1_api_key_binding_1234567890
```

## Notas de Uso

### Segurança

1. Siga o princípio do menor privilégio ao conceder permissões.
2. Use `resource_ids` para limitar o escopo das permissões quando possível.
3. Revise regularmente as permissões concedidas às chaves de API.
4. Documente o propósito e uso de cada binding.

### Boas Práticas

1. Agrupe bindings relacionados com a mesma chave de API.
2. Use descrições claras nas chaves de API para identificar seu propósito.
3. Mantenha um inventário de bindings e suas permissões.
4. Implemente rotação regular de chaves de API.

### Exemplo de Validação de Permissões

```python
from typing import List, Dict

def validate_api_key_permissions(
    api_key: str,
    required_permissions: Dict[str, List[str]]
) -> bool:
    """
    Valida se uma chave de API tem as permissões necessárias.
    
    Args:
        api_key: A chave de API a ser validada
        required_permissions: Dicionário de tipo de recurso para lista de permissões
        
    Returns:
        bool: True se a chave tem todas as permissões necessárias
    """
    # Implementar lógica de validação de permissões
    return True

# Exemplo de uso
required_permissions = {
    'user': ['read', 'list', 'create'],
    'user_group': ['read', 'list'],
    'system': ['read']
}

is_valid = validate_api_key_permissions(
    'your_api_key',
    required_permissions
)
```

### Exemplo de Auditoria de Permissões

```python
from datetime import datetime
from typing import Dict, List

def audit_api_key_bindings(
    bindings: List[Dict]
) -> Dict[str, List[str]]:
    """
    Audita bindings de chaves de API para identificar permissões sensíveis.
    
    Args:
        bindings: Lista de bindings a serem auditados
        
    Returns:
        Dict[str, List[str]]: Relatório de permissões sensíveis por chave
    """
    sensitive_permissions = {
        'user': ['delete'],
        'system': ['delete'],
        'organization': ['update', 'delete']
    }
    
    audit_report = {}
    
    for binding in bindings:
        api_key_id = binding['api_key_id']
        resource_type = binding['resource_type']
        permissions = binding['permissions']
        
        if resource_type in sensitive_permissions:
            sensitive = sensitive_permissions[resource_type]
            found = [p for p in permissions if p in sensitive]
            
            if found:
                if api_key_id not in audit_report:
                    audit_report[api_key_id] = []
                audit_report[api_key_id].extend([
                    f"{resource_type}:{p}" for p in found
                ])
    
    return audit_report

# Exemplo de uso
bindings = [
    {
        'api_key_id': 'key1',
        'resource_type': 'user',
        'permissions': ['read', 'delete']
    },
    {
        'api_key_id': 'key1',
        'resource_type': 'organization',
        'permissions': ['read', 'update']
    }
]

report = audit_api_key_bindings(bindings)
``` 