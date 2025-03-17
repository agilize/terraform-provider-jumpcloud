# jumpcloud_application_user_mapping Resource

Este recurso permite gerenciar mapeamentos entre usuários e aplicações no JumpCloud, concedendo acesso de usuários individuais a aplicações SSO.

## Exemplo de Uso

```hcl
# Mapeamento básico de usuário para aplicação
resource "jumpcloud_application_user_mapping" "admin_salesforce" {
  application_id = jumpcloud_application.salesforce.id
  user_id        = jumpcloud_user.admin.id
}

# Mapeamento com atributos personalizados
resource "jumpcloud_application_user_mapping" "dev_jira" {
  application_id = jumpcloud_application.jira.id
  user_id        = jumpcloud_user.developer.id
  
  attributes = {
    "role"     = "developer"
    "projects" = "alpha,beta,gamma"
    "admin"    = "false"
  }
}

# Mapeamento usando data sources para recursos existentes
resource "jumpcloud_application_user_mapping" "existing_mapping" {
  application_id = data.jumpcloud_application.existing_app.id
  user_id        = data.jumpcloud_user.existing_user.id
  
  attributes = {
    "access_level" = "standard"
    "department"   = "marketing"
  }
}
```

## Argument Reference

Os seguintes argumentos são suportados:

* `application_id` - (Obrigatório) ID da aplicação JumpCloud.
* `user_id` - (Obrigatório) ID do usuário JumpCloud.
* `attributes` - (Opcional) Mapa de atributos personalizados para o mapeamento. Estes atributos são específicos para cada tipo de aplicação e podem ser usados para definir funções, permissões ou outras configurações específicas da aplicação para o usuário.

## Import

Mapeamentos de usuário-aplicação JumpCloud podem ser importados usando uma string separada por dois pontos no formato:

```
terraform import jumpcloud_application_user_mapping.example {application_id}:{user_id}
``` 