# jumpcloud_application Data Source

Este data source permite obter informações sobre uma aplicação específica configurada no JumpCloud. Pode ser útil para referenciar aplicações existentes sem precisar recriá-las em seu código Terraform.

## Exemplo de Uso

### Buscar por ID

```hcl
data "jumpcloud_application" "salesforce" {
  id = "5f8d3d0d9d1d8b6a8c4c1d9e"
}

output "salesforce_sso_url" {
  value = data.jumpcloud_application.salesforce.sso_url
}
```

### Buscar por Nome

```hcl
data "jumpcloud_application" "jira" {
  name = "Jira Cloud"
}

# Usar o ID em outro recurso
resource "jumpcloud_application_user_mapping" "jira_user" {
  application_id = data.jumpcloud_application.jira.id
  user_id        = jumpcloud_user.dev_user.id
}
```

## Argument Reference

Os seguintes argumentos são suportados:

* `id` - (Opcional) ID da aplicação no JumpCloud. Conflita com `name`.
* `name` - (Opcional) Nome da aplicação no JumpCloud. Conflita com `id`.

**Nota**: Exatamente um de `id` ou `name` deve ser especificado.

## Attribute Reference

Os seguintes atributos são exportados:

* `display_name` - Nome de exibição da aplicação que é mostrado na interface do JumpCloud.
* `description` - Descrição da aplicação.
* `sso_url` - URL de SSO para a aplicação. Geralmente usado para aplicações SAML.
* `saml_metadata` - Metadados SAML da aplicação.
* `type` - Tipo da aplicação. Pode ser `saml`, `oidc` ou `oauth`.
* `config` - Mapa de configurações específicas para o tipo de aplicação.
* `logo` - URL ou string base64 da imagem do logo da aplicação.
* `active` - Indica se a aplicação está ativa.
* `created` - Data de criação da aplicação.
* `updated` - Data da última atualização da aplicação. 