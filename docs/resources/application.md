# jumpcloud_application Resource

Este recurso permite gerenciar aplicações no JumpCloud para fornecer Single Sign-On (SSO) para seus usuários. O JumpCloud suporta aplicações SAML, OAuth e OIDC.

## Exemplo de Uso

### Aplicação SAML básica

```hcl
resource "jumpcloud_application" "salesforce" {
  name         = "Salesforce"
  display_name = "Salesforce SSO"
  description  = "Acesso SSO ao Salesforce para todos os funcionários"
  type         = "saml"
  sso_url      = "https://login.salesforce.com"
  
  config = {
    idp_entity_id  = "https://sso.jumpcloud.com/saml2/salesforce"
    sp_entity_id   = "https://login.salesforce.com"
    acs_url        = "https://login.salesforce.com/services/saml2/acs"
    constant_attribute_name_format = "true"
  }
}
```

### Aplicação SAML com logo e atributos

```hcl
resource "jumpcloud_application" "jira" {
  name         = "Jira Cloud"
  display_name = "Jira"
  description  = "Gerenciamento de projetos e tarefas"
  type         = "saml"
  logo         = "https://example.com/jira-logo.png"
  sso_url      = "https://yourdomain.atlassian.net"
  
  config = {
    idp_entity_id  = "https://sso.jumpcloud.com/saml2/jira"
    sp_entity_id   = "https://yourdomain.atlassian.net"
    acs_url        = "https://yourdomain.atlassian.net/plugins/servlet/saml/acs"
    idp_initiated_login = "true"
    name_id_format = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"
    attribute_statements = jsonencode({
      "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier" = "user.email"
      "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/displayname" = "user.firstname user.lastname"
      "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/groups" = "user.groups"
    })
  }
}
```

### Aplicação OAuth

```hcl
resource "jumpcloud_application" "custom_oauth" {
  name         = "API Portal"
  display_name = "Portal de APIs"
  description  = "Portal de acesso às APIs da empresa"
  type         = "oauth"
  active       = true
  
  config = {
    client_id       = "api-portal-client"
    client_secret   = "s3cr3t-n0t-1n-t3rr4f0rm"
    redirect_uri    = "https://api.example.com/callback"
    grant_types     = "authorization_code,refresh_token"
    allowed_scopes  = "read:api,write:api"
  }
}
```

## Argument Reference

Os seguintes argumentos são suportados:

* `name` - (Obrigatório) Nome da aplicação.
* `type` - (Obrigatório, ForceNew) Tipo da aplicação. Valores válidos: `saml`, `oidc`, `oauth`.
* `display_name` - (Opcional) Nome de exibição da aplicação que será mostrado na interface do JumpCloud.
* `description` - (Opcional) Descrição da aplicação.
* `sso_url` - (Opcional) URL de SSO para a aplicação. Geralmente usado para aplicações SAML.
* `saml_metadata` - (Opcional) Metadados SAML para a aplicação. Pode ser um XML no formato de metadados SAML.
* `logo` - (Opcional) URL ou string base64 da imagem do logo da aplicação.
* `active` - (Opcional) Define se a aplicação está ativa. Padrão: `true`.
* `config` - (Opcional) Mapa de configurações específicas para o tipo de aplicação. Os valores dependerão do tipo de aplicação:
  
  **Para SAML**:
  * `idp_entity_id` - ID da entidade do provedor de identidade (JumpCloud).
  * `sp_entity_id` - ID da entidade do provedor de serviço (aplicação).
  * `acs_url` - URL do serviço de consumo de asserção.
  * `idp_initiated_login` - "true" ou "false" para habilitar login iniciado pelo IdP.
  * `name_id_format` - Formato do NameID.
  * `attribute_statements` - JSON codificado com mapeamentos de atributos.
  
  **Para OAuth/OIDC**:
  * `client_id` - ID do cliente.
  * `client_secret` - Segredo do cliente (sensível).
  * `redirect_uri` - URI de redirecionamento.
  * `grant_types` - Tipos de concessão separados por vírgula.
  * `allowed_scopes` - Escopos permitidos separados por vírgula.

## Attribute Reference

Além dos argumentos listados acima, os seguintes atributos são exportados:

* `id` - ID da aplicação.
* `created` - Data de criação da aplicação.
* `updated` - Data da última atualização da aplicação.

## Import

Aplicações JumpCloud podem ser importadas usando o ID da aplicação:

```
terraform import jumpcloud_application.example {application_id}
``` 