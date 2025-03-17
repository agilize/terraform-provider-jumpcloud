# jumpcloud_webhook

Gerencia webhooks no JumpCloud, permitindo que você configure notificações em tempo real para eventos específicos da sua organização.

## Exemplo de Uso

### Webhook Básico para Monitoramento de Segurança
```hcl
resource "jumpcloud_webhook" "security_monitoring" {
  name        = "Security Events Monitor"
  url         = "https://security.example.com/jumpcloud-events"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook para monitoramento de eventos de segurança"
  
  event_types = [
    "user.login.failed",
    "user.admin.updated",
    "security.alert",
    "mfa.disabled"
  ]
}
```

### Webhook para Automação de Usuários
```hcl
resource "jumpcloud_webhook" "user_automation" {
  name        = "User Management Automation"
  url         = "https://automation.example.com/users"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook para automação de gerenciamento de usuários"
  
  event_types = [
    "user.created",
    "user.updated",
    "user.deleted",
    "user.login.success"
  ]
}
```

### Webhook para Monitoramento de Sistemas
```hcl
resource "jumpcloud_webhook" "system_monitoring" {
  name        = "System Events Monitor"
  url         = "https://monitoring.example.com/systems"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook para monitoramento de eventos de sistemas"
  
  event_types = [
    "system.created",
    "system.updated",
    "system.deleted"
  ]
}
```

### Webhook para Auditoria de Aplicações
```hcl
resource "jumpcloud_webhook" "application_audit" {
  name        = "Application Access Audit"
  url         = "https://audit.example.com/applications"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook para auditoria de acesso a aplicações"
  
  event_types = [
    "application.access.granted",
    "application.access.revoked"
  ]
}
```

## Argumentos

Os seguintes argumentos são suportados:

* `name` - (Obrigatório) Nome do webhook. Deve ser único dentro da organização.
* `url` - (Obrigatório) URL de destino para onde os eventos serão enviados. Deve usar HTTPS.
* `secret` - (Opcional) Chave secreta usada para assinar as requisições webhook. Recomendado para segurança.
* `enabled` - (Opcional) Define se o webhook está ativo. Padrão é `true`.
* `event_types` - (Obrigatório) Lista de tipos de eventos que dispararão o webhook. Deve conter pelo menos um evento.
* `description` - (Opcional) Descrição do webhook para documentação.

### Tipos de Eventos Suportados

Os seguintes tipos de eventos são suportados:

**Eventos de Usuário:**
* `user.created` - Usuário criado
* `user.updated` - Usuário atualizado
* `user.deleted` - Usuário excluído
* `user.login.success` - Login bem-sucedido
* `user.login.failed` - Tentativa de login falhou
* `user.admin.updated` - Permissões administrativas alteradas

**Eventos de Sistema:**
* `system.created` - Sistema adicionado
* `system.updated` - Sistema atualizado
* `system.deleted` - Sistema removido

**Eventos de Organização:**
* `organization.created` - Organização criada
* `organization.updated` - Organização atualizada
* `organization.deleted` - Organização excluída

**Eventos de API Key:**
* `api_key.created` - Chave de API criada
* `api_key.updated` - Chave de API atualizada
* `api_key.deleted` - Chave de API excluída

**Eventos de Webhook:**
* `webhook.created` - Webhook criado
* `webhook.updated` - Webhook atualizado
* `webhook.deleted` - Webhook excluído

**Eventos de Segurança:**
* `security.alert` - Alerta de segurança gerado
* `mfa.enabled` - MFA habilitado
* `mfa.disabled` - MFA desabilitado

**Eventos de Política:**
* `policy.applied` - Política aplicada
* `policy.removed` - Política removida

**Eventos de Aplicação:**
* `application.access.granted` - Acesso concedido à aplicação
* `application.access.revoked` - Acesso revogado da aplicação

## Atributos Exportados

Além dos argumentos acima, os seguintes atributos são exportados:

* `id` - ID único do webhook.
* `created` - Data de criação do webhook no formato ISO 8601.
* `updated` - Data da última atualização do webhook no formato ISO 8601.

## Importação

Webhooks podem ser importados usando seu ID:

```shell
terraform import jumpcloud_webhook.security_monitoring j1_webhook_1234567890
```

## Notas de Uso

### Segurança

1. Sempre use HTTPS para a URL do webhook.
2. Configure uma chave secreta forte para validar as requisições.
3. Implemente validação da assinatura no endpoint que recebe os eventos.

### Boas Práticas

1. Agrupe eventos relacionados em webhooks separados para melhor organização.
2. Use descrições claras para documentar o propósito de cada webhook.
3. Monitore a performance do seu endpoint para garantir que pode processar o volume de eventos.
4. Implemente retry logic no seu endpoint para eventos importantes.

### Exemplo de Validação de Assinatura

```python
import hmac
import hashlib

def verify_signature(secret, payload, signature):
    expected = hmac.new(
        secret.encode('utf-8'),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(signature, expected)
```

### Exemplo de Endpoint com Retry

```python
from flask import Flask, request
from functools import wraps
import time

app = Flask(__name__)

def retry_on_failure(max_retries=3, delay=1):
    def decorator(f):
        @wraps(f)
        def wrapper(*args, **kwargs):
            retries = 0
            while retries < max_retries:
                try:
                    return f(*args, **kwargs)
                except Exception as e:
                    retries += 1
                    if retries == max_retries:
                        raise e
                    time.sleep(delay)
            return None
        return wrapper
    return decorator

@app.route('/jumpcloud-events', methods=['POST'])
@retry_on_failure()
def handle_webhook():
    # Implementar lógica de processamento do evento
    return '', 200
``` 