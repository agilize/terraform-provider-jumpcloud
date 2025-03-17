# jumpcloud_webhook Data Source

Use este data source para recuperar informações sobre um webhook específico existente no JumpCloud.

## Exemplo de Uso

```hcl
# Obter um webhook por ID
data "jumpcloud_webhook" "by_id" {
  id = "5f1b1bb2c9e9a9b7e8d6c5a4"
}

# Obter um webhook por nome
data "jumpcloud_webhook" "by_name" {
  name = "User Events Monitor"
}

# Usar o webhook encontrado em outra configuração
resource "jumpcloud_webhook_subscription" "additional_events" {
  webhook_id   = data.jumpcloud_webhook.by_name.id
  event_type   = "user.password_expired"
  description  = "Adicionar notificação de expiração de senha ao webhook existente"
}

# Output com informações do webhook
output "webhook_details" {
  value = {
    id          = data.jumpcloud_webhook.by_name.id
    url         = data.jumpcloud_webhook.by_name.url
    enabled     = data.jumpcloud_webhook.by_name.enabled
    event_types = data.jumpcloud_webhook.by_name.event_types
    created     = data.jumpcloud_webhook.by_name.created
  }
}
```

## Argument Reference

Os seguintes argumentos são suportados. **Nota:** Exatamente um desses argumentos deve ser especificado:

* `id` - (Opcional) O ID do webhook a ser recuperado.
* `name` - (Opcional) O nome do webhook a ser recuperado.

## Attribute Reference

Além de todos os argumentos acima, os seguintes atributos são exportados:

* `url` - A URL de destino para o webhook.
* `enabled` - Se o webhook está ativado ou não.
* `event_types` - Lista de tipos de eventos que disparam o webhook.
* `description` - A descrição do webhook.
* `created` - A data de criação do webhook.
* `updated` - A data da última atualização do webhook. 