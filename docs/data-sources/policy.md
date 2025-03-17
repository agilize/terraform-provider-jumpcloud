# jumpcloud_policy Data Source

Use este data source para obter informações sobre uma política específica existente no JumpCloud. Este data source permite recuperar detalhes sobre políticas como complexidade de senha, MFA e outras configurações de segurança.

## Exemplo de Uso

```hcl
# Buscar política por nome
data "jumpcloud_policy" "password_policy" {
  name = "Secure Password Policy"
}

# Buscar política por ID
data "jumpcloud_policy" "mfa_policy" {
  id = "5f8b0e1b9d81b81b33c92a1c" # Exemplo de ID (substitua pelo ID real)
}

# Verificar se política está ativa
output "password_policy_status" {
  value = "${data.jumpcloud_policy.password_policy.name} está ${data.jumpcloud_policy.password_policy.active ? "ativa" : "inativa"}"
}

# Verificar configurações da política
output "password_policy_min_length" {
  value = lookup(data.jumpcloud_policy.password_policy.configurations, "min_length", "não especificado")
}

# Exibir tipo e template da política
output "mfa_policy_details" {
  value = "Política: ${data.jumpcloud_policy.mfa_policy.name}, Tipo: ${data.jumpcloud_policy.mfa_policy.type}, Criada em: ${data.jumpcloud_policy.mfa_policy.created}"
}
```

## Uso Condicional

Este data source é útil para criar recursos condicionalmente com base em políticas existentes:

```hcl
# Verificar se uma política de MFA já existe
data "jumpcloud_policy" "existing_mfa" {
  name = "Required MFA Policy"
}

# Criar a política apenas se não existir
resource "jumpcloud_policy" "conditional_mfa" {
  count       = data.jumpcloud_policy.existing_mfa.id != "" ? 0 : 1
  name        = "Required MFA Policy"
  description = "Política criada condicionalmente pelo Terraform"
  type        = "mfa"
  active      = true
  
  configurations = {
    require_mfa_for_all_users = "true"
  }
}
```

## Referência de Argumentos

Os seguintes argumentos são suportados:

* `id` - (Opcional) O ID da política no JumpCloud.
* `name` - (Opcional) O nome da política no JumpCloud.

> **Nota:** É necessário especificar exatamente um desses argumentos.

## Referência de Atributos

Os seguintes atributos são exportados:

* `id` - O ID da política.
* `name` - O nome da política.
* `description` - A descrição da política.
* `type` - O tipo da política. Possíveis valores: `password_complexity`, `samba_ad_password_sync`, `password_expiration`, `custom`, `password_reused`, `password_failed_attempts`, `account_lockout_timeout`, `mfa`, `system_updates`.
* `template` - O template usado pela política.
* `active` - Indica se a política está ativa.
* `created` - A data de criação da política.
* `configurations` - Um mapa das configurações específicas da política.
* `organization_id` - O ID da organização à qual a política pertence.

## Associação com Outros Recursos

As informações de uma política existente podem ser usadas para associá-la a grupos:

```hcl
data "jumpcloud_policy" "existing_password_policy" {
  name = "Secure Password Policy"
}

data "jumpcloud_user_group" "finance" {
  name = "Finance Department"
}

resource "jumpcloud_policy_association" "finance_password_policy" {
  policy_id = data.jumpcloud_policy.existing_password_policy.id
  group_id  = data.jumpcloud_user_group.finance.id
  type      = "user_group"
}
```

## Casos de Uso Comuns

1. **Validação de Políticas**: Verificar se políticas específicas existem e estão ativas antes de criar novas associações.
2. **Automação Condicional**: Criar recursos apenas se determinadas políticas não existirem.
3. **Reporting**: Gerar relatórios sobre as configurações de políticas em uso.
4. **Validação de Configuração**: Verificar se as políticas existentes estão configuradas conforme os requisitos de segurança.
5. **Integração com Outros Sistemas**: Usar informações de políticas para integrar com outros sistemas ou ferramentas de automação. 