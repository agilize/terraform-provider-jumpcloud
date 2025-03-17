# jumpcloud_application_group_mapping Resource

Este recurso permite gerenciar mapeamentos entre grupos e aplicações no JumpCloud, concedendo acesso de grupos de usuários ou grupos de sistemas a aplicações SSO.

## Exemplo de Uso

```hcl
# Mapeamento de grupo de usuários para aplicação
resource "jumpcloud_application_group_mapping" "marketing_team" {
  application_id = jumpcloud_application.salesforce.id
  group_id       = jumpcloud_user_group.marketing.id
  type           = "user_group"  # Padrão
  
  attributes = {
    "access_level" = "standard"
    "department"   = "Marketing"
  }
}

# Mapeamento de grupo de sistemas para aplicação
resource "jumpcloud_application_group_mapping" "production_servers" {
  application_id = jumpcloud_application.monitoring_tool.id
  group_id       = jumpcloud_system_group.production.id
  type           = "system_group"
}

# Mapeamento usando data sources para recursos existentes
resource "jumpcloud_application_group_mapping" "existing_mapping" {
  application_id = data.jumpcloud_application.existing_app.id
  group_id       = data.jumpcloud_user_group.existing_group.id
  
  attributes = {
    "role"     = "viewer"
    "region"   = "us-west"
    "enabled"  = "true"
  }
}
```

## Argument Reference

Os seguintes argumentos são suportados:

* `application_id` - (Obrigatório) ID da aplicação JumpCloud.
* `group_id` - (Obrigatório) ID do grupo JumpCloud.
* `type` - (Opcional) Tipo de grupo: `user_group` (padrão) ou `system_group`.
* `attributes` - (Opcional) Mapa de atributos personalizados para o mapeamento. Estes atributos são específicos para cada tipo de aplicação.

## Import

Mapeamentos de grupo-aplicação JumpCloud podem ser importados usando uma string separada por dois pontos no formato:

```
terraform import jumpcloud_application_group_mapping.example {application_id}:{group_type}:{group_id}
``` 