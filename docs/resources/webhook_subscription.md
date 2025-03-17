# jumpcloud_webhook_subscription

Gerencia assinaturas de eventos para webhooks no JumpCloud. Este recurso permite especificar quais eventos específicos um webhook deve monitorar, permitindo um controle granular sobre as notificações.

## Exemplo de Uso

### Monitoramento de Segurança
```hcl
# Criar webhook para monitoramento de segurança
resource "jumpcloud_webhook" "security_monitor" {
  name        = "Security Events Monitor"
  url         = "https://security.example.com/jumpcloud-events"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook para monitoramento de eventos de segurança"
}

# Assinar eventos de falha de login
resource "jumpcloud_webhook_subscription" "failed_logins" {
  webhook_id   = jumpcloud_webhook.security_monitor.id
  event_type   = "user.login.failed"
  description  = "Monitorar tentativas de login mal sucedidas"
}

# Assinar eventos de alteração de MFA
resource "jumpcloud_webhook_subscription" "mfa_changes" {
  webhook_id   = jumpcloud_webhook.security_monitor.id
  event_type   = "mfa.disabled"
  description  = "Monitorar desativação de MFA"
}

# Assinar alertas de segurança
resource "jumpcloud_webhook_subscription" "security_alerts" {
  webhook_id   = jumpcloud_webhook.security_monitor.id
  event_type   = "security.alert"
  description  = "Monitorar alertas de segurança"
}
```

### Automação de Usuários
```hcl
# Criar webhook para automação de usuários
resource "jumpcloud_webhook" "user_automation" {
  name        = "User Management Automation"
  url         = "https://automation.example.com/users"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook para automação de gerenciamento de usuários"
}

# Assinar eventos de criação de usuário
resource "jumpcloud_webhook_subscription" "user_created" {
  webhook_id   = jumpcloud_webhook.user_automation.id
  event_type   = "user.created"
  description  = "Notificar quando novos usuários são criados"
}

# Assinar eventos de atualização de usuário
resource "jumpcloud_webhook_subscription" "user_updated" {
  webhook_id   = jumpcloud_webhook.user_automation.id
  event_type   = "user.updated"
  description  = "Notificar quando usuários são atualizados"
}

# Assinar eventos de exclusão de usuário
resource "jumpcloud_webhook_subscription" "user_deleted" {
  webhook_id   = jumpcloud_webhook.user_automation.id
  event_type   = "user.deleted"
  description  = "Notificar quando usuários são excluídos"
}
```

### Monitoramento de Sistemas
```hcl
# Criar webhook para monitoramento de sistemas
resource "jumpcloud_webhook" "system_monitor" {
  name        = "System Events Monitor"
  url         = "https://monitoring.example.com/systems"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook para monitoramento de eventos de sistemas"
}

# Assinar eventos de criação de sistema
resource "jumpcloud_webhook_subscription" "system_created" {
  webhook_id   = jumpcloud_webhook.system_monitor.id
  event_type   = "system.created"
  description  = "Notificar quando novos sistemas são adicionados"
}

# Assinar eventos de atualização de sistema
resource "jumpcloud_webhook_subscription" "system_updated" {
  webhook_id   = jumpcloud_webhook.system_monitor.id
  event_type   = "system.updated"
  description  = "Notificar quando sistemas são atualizados"
}

# Assinar eventos de remoção de sistema
resource "jumpcloud_webhook_subscription" "system_deleted" {
  webhook_id   = jumpcloud_webhook.system_monitor.id
  event_type   = "system.deleted"
  description  = "Notificar quando sistemas são removidos"
}
```

### Auditoria de Aplicações
```hcl
# Criar webhook para auditoria de aplicações
resource "jumpcloud_webhook" "application_audit" {
  name        = "Application Access Audit"
  url         = "https://audit.example.com/applications"
  secret      = var.webhook_secret
  enabled     = true
  description = "Webhook para auditoria de acesso a aplicações"
}

# Assinar eventos de concessão de acesso
resource "jumpcloud_webhook_subscription" "access_granted" {
  webhook_id   = jumpcloud_webhook.application_audit.id
  event_type   = "application.access.granted"
  description  = "Notificar quando acesso é concedido a aplicações"
}

# Assinar eventos de revogação de acesso
resource "jumpcloud_webhook_subscription" "access_revoked" {
  webhook_id   = jumpcloud_webhook.application_audit.id
  event_type   = "application.access.revoked"
  description  = "Notificar quando acesso é revogado de aplicações"
}
```

## Argumentos

Os seguintes argumentos são suportados:

* `webhook_id` - (Obrigatório) ID do webhook ao qual esta assinatura pertence.
* `event_type` - (Obrigatório) Tipo de evento que será monitorado. Veja a lista completa de eventos suportados na documentação do recurso `jumpcloud_webhook`.
* `description` - (Opcional) Descrição do propósito desta assinatura de evento.

## Atributos Exportados

Além dos argumentos acima, os seguintes atributos são exportados:

* `id` - ID único da assinatura do webhook.
* `created` - Data de criação da assinatura no formato ISO 8601.
* `updated` - Data da última atualização da assinatura no formato ISO 8601.

## Importação

Assinaturas de webhook podem ser importadas usando seu ID:

```shell
terraform import jumpcloud_webhook_subscription.failed_logins j1_webhook_sub_1234567890
```

## Notas de Uso

### Boas Práticas

1. Use descrições claras e específicas para cada assinatura, facilitando o entendimento do propósito.
2. Agrupe assinaturas relacionadas com o mesmo webhook para melhor organização.
3. Considere o volume de eventos ao assinar múltiplos tipos de eventos no mesmo webhook.
4. Documente o propósito e o fluxo de processamento de cada tipo de evento assinado.

### Exemplo de Processamento de Eventos

```python
from flask import Flask, request
import json

app = Flask(__name__)

def process_login_failed(event_data):
    user = event_data.get('user')
    ip = event_data.get('source_ip')
    # Implementar lógica de alerta para falhas de login
    
def process_mfa_disabled(event_data):
    user = event_data.get('user')
    admin = event_data.get('admin')
    # Implementar lógica de auditoria para desativação de MFA

def process_system_created(event_data):
    system = event_data.get('system')
    # Implementar lógica de inventário para novos sistemas

event_handlers = {
    'user.login.failed': process_login_failed,
    'mfa.disabled': process_mfa_disabled,
    'system.created': process_system_created
}

@app.route('/jumpcloud-events', methods=['POST'])
def handle_webhook():
    event = request.json
    event_type = event.get('type')
    
    if event_type in event_handlers:
        handler = event_handlers[event_type]
        handler(event.get('data', {}))
        
    return '', 200
``` 