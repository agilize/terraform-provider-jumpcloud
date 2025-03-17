# jumpcloud_api_key

Gerencia chaves de API no JumpCloud. Este recurso permite criar e gerenciar chaves de API que podem ser usadas para autenticar requisições às APIs do JumpCloud.

## Exemplo de Uso

### Chave de API para Automação
```hcl
# Criar uma chave de API para automação
resource "jumpcloud_api_key" "automation" {
  name        = "Automation API Key"
  description = "Chave de API para automação de processos"
  expires     = timeadd(timestamp(), "8760h") # Expira em 1 ano
}

# Configurar permissões para a chave
resource "jumpcloud_api_key_binding" "automation_user_management" {
  api_key_id    = jumpcloud_api_key.automation.id
  resource_type = "user"
  permissions   = ["read", "list", "create", "update"]
}

# Exportar a chave de forma segura
output "automation_api_key" {
  value     = jumpcloud_api_key.automation.key
  sensitive = true
}
```

### Chave de API Temporária
```hcl
# Criar uma chave de API temporária para um projeto
resource "jumpcloud_api_key" "temp_project" {
  name        = "Temporary Project Key"
  description = "Chave temporária para projeto de migração"
  expires     = timeadd(timestamp(), "168h") # Expira em 1 semana
}

# Configurar permissões mínimas necessárias
resource "jumpcloud_api_key_binding" "temp_project_access" {
  api_key_id    = jumpcloud_api_key.temp_project.id
  resource_type = "user"
  permissions   = ["read", "list"]
}

# Exportar a chave e data de expiração
output "temp_project_key" {
  value = {
    key     = jumpcloud_api_key.temp_project.key
    expires = jumpcloud_api_key.temp_project.expires
  }
  sensitive = true
}
```

### Chave de API para Monitoramento
```hcl
# Criar uma chave de API para monitoramento
resource "jumpcloud_api_key" "monitoring" {
  name        = "System Monitoring Key"
  description = "Chave para monitoramento de sistemas e eventos"
}

# Configurar permissões para monitorar sistemas
resource "jumpcloud_api_key_binding" "monitoring_system_access" {
  api_key_id    = jumpcloud_api_key.monitoring.id
  resource_type = "system"
  permissions   = ["read", "list"]
}

# Configurar permissões para monitorar eventos
resource "jumpcloud_api_key_binding" "monitoring_event_access" {
  api_key_id    = jumpcloud_api_key.monitoring.id
  resource_type = "auth_event"
  permissions   = ["read", "list"]
}

# Exportar configuração para uso em ferramentas de monitoramento
output "monitoring_config" {
  value = {
    api_key_id = jumpcloud_api_key.monitoring.id
    key        = jumpcloud_api_key.monitoring.key
    created    = jumpcloud_api_key.monitoring.created
  }
  sensitive = true
}
```

### Chave de API para CI/CD
```hcl
# Criar uma chave de API para pipeline de CI/CD
resource "jumpcloud_api_key" "cicd" {
  name        = "CI/CD Pipeline Key"
  description = "Chave para automação de deploy"
  expires     = timeadd(timestamp(), "4380h") # Expira em 6 meses
}

# Configurar permissões para gerenciar aplicações
resource "jumpcloud_api_key_binding" "cicd_app_management" {
  api_key_id    = jumpcloud_api_key.cicd.id
  resource_type = "application"
  permissions   = ["read", "list", "create", "update", "delete"]
}

# Configurar permissões para gerenciar políticas
resource "jumpcloud_api_key_binding" "cicd_policy_management" {
  api_key_id    = jumpcloud_api_key.cicd.id
  resource_type = "policy"
  permissions   = ["read", "list", "create", "update"]
}

# Exportar configuração para o pipeline
output "cicd_config" {
  value = {
    key        = jumpcloud_api_key.cicd.key
    expires    = jumpcloud_api_key.cicd.expires
    created_at = jumpcloud_api_key.cicd.created
  }
  sensitive = true
}
```

## Argumentos

Os seguintes argumentos são suportados:

* `name` - (Obrigatório) Nome da chave de API. Deve ser único dentro da organização.
* `description` - (Opcional) Descrição do propósito da chave de API.
* `expires` - (Opcional) Data de expiração da chave no formato RFC3339. Se não especificado, a chave não expira.

## Atributos Exportados

Além dos argumentos acima, os seguintes atributos são exportados:

* `id` - ID único da chave de API.
* `key` - A chave de API gerada. Este valor é sensível e só é mostrado uma vez após a criação.
* `created` - Data de criação da chave no formato ISO 8601.
* `updated` - Data da última atualização da chave no formato ISO 8601.

## Importação

Chaves de API podem ser importadas usando seu ID:

```shell
terraform import jumpcloud_api_key.automation j1_api_key_1234567890
```

## Notas de Uso

### Segurança

1. Trate chaves de API como credenciais sensíveis.
2. Use o atributo `expires` para limitar a vida útil das chaves.
3. Implemente rotação regular de chaves.
4. Siga o princípio do menor privilégio ao configurar permissões.

### Boas Práticas

1. Use nomes descritivos que identifiquem claramente o propósito da chave.
2. Documente o uso e escopo de cada chave.
3. Mantenha um inventário de chaves ativas.
4. Configure alertas para expiração de chaves.

### Exemplo de Rotação de Chaves

```python
from datetime import datetime, timedelta
import requests

def rotate_api_key(
    current_key: str,
    key_name: str,
    description: str
) -> str:
    """
    Rotaciona uma chave de API criando uma nova e desativando a antiga.
    
    Args:
        current_key: Chave atual a ser rotacionada
        key_name: Nome para a nova chave
        description: Descrição da nova chave
        
    Returns:
        str: Nova chave de API
    """
    # Criar nova chave
    new_key = create_api_key(
        name=key_name,
        description=description,
        expires=datetime.now() + timedelta(days=90)
    )
    
    # Validar nova chave
    if not validate_api_key(new_key):
        raise ValueError("Failed to validate new API key")
        
    # Desativar chave antiga
    disable_api_key(current_key)
    
    return new_key

def validate_api_key(api_key: str) -> bool:
    """
    Valida se uma chave de API está funcionando corretamente.
    
    Args:
        api_key: Chave de API a ser validada
        
    Returns:
        bool: True se a chave é válida
    """
    try:
        response = requests.get(
            'https://api.jumpcloud.com/v2/organizations',
            headers={'x-api-key': api_key}
        )
        return response.status_code == 200
    except:
        return False
```

### Exemplo de Monitoramento de Expiração

```python
from datetime import datetime, timedelta
import smtplib
from email.message import EmailMessage
from typing import List, Dict

def check_expiring_keys(
    keys: List[Dict],
    warning_days: int = 30
) -> List[Dict]:
    """
    Verifica chaves que estão próximas de expirar.
    
    Args:
        keys: Lista de chaves de API com seus metadados
        warning_days: Dias de antecedência para alertar
        
    Returns:
        List[Dict]: Lista de chaves próximas de expirar
    """
    now = datetime.now()
    warning_date = now + timedelta(days=warning_days)
    
    expiring = []
    for key in keys:
        if key.get('expires'):
            expires = datetime.fromisoformat(key['expires'])
            if now < expires <= warning_date:
                expiring.append({
                    'id': key['id'],
                    'name': key['name'],
                    'expires': key['expires'],
                    'days_left': (expires - now).days
                })
    
    return expiring

def send_expiration_alert(
    expiring_keys: List[Dict],
    email_to: str
) -> None:
    """
    Envia alerta por email sobre chaves próximas de expirar.
    
    Args:
        expiring_keys: Lista de chaves próximas de expirar
        email_to: Endereço de email para enviar o alerta
    """
    if not expiring_keys:
        return
        
    msg = EmailMessage()
    msg.set_content(
        'As seguintes chaves de API estão próximas de expirar:\n\n' +
        '\n'.join(
            f"- {k['name']}: expira em {k['days_left']} dias"
            for k in expiring_keys
        )
    )
    
    msg['Subject'] = 'Alerta: Chaves de API próximas de expirar'
    msg['From'] = 'alerts@example.com'
    msg['To'] = email_to
    
    # Implementar envio do email
``` 