# jumpcloud_policy_association Resource

Este recurso permite associar políticas do JumpCloud a grupos de usuários ou sistemas, aplicando as configurações de segurança e conformidade definidas nas políticas aos membros do grupo.

## Exemplo de Uso

```hcl
# Criar uma política de complexidade de senha
resource "jumpcloud_policy" "password_complexity" {
  name        = "Secure Password Policy"
  description = "Política de complexidade de senha para o departamento financeiro"
  type        = "password_complexity"
  active      = true
  
  configurations = {
    min_length             = "12"
    requires_uppercase     = "true"
    requires_lowercase     = "true"
    requires_number        = "true"
    requires_special_char  = "true"
  }
}

# Criar um grupo de usuários
resource "jumpcloud_user_group" "finance" {
  name        = "Finance Department"
  description = "Grupo de usuários do departamento financeiro"
}

# Associar a política ao grupo de usuários
resource "jumpcloud_policy_association" "finance_password_policy" {
  policy_id = jumpcloud_policy.password_complexity.id
  group_id  = jumpcloud_user_group.finance.id
  type      = "user_group"
}

# Associar política a um grupo de sistemas
resource "jumpcloud_system_group" "servers" {
  name        = "Production Servers"
  description = "Grupo de servidores de produção"
}

resource "jumpcloud_policy" "system_updates" {
  name        = "System Updates Policy"
  description = "Política para controle de atualizações de sistema"
  type        = "system_updates"
  active      = true
  
  configurations = {
    auto_update_enabled = "true"
    auto_update_time    = "02:00"
  }
}

resource "jumpcloud_policy_association" "servers_update_policy" {
  policy_id = jumpcloud_policy.system_updates.id
  group_id  = jumpcloud_system_group.servers.id
  type      = "system_group"
}
```

## Referência de Argumentos

Os seguintes argumentos são suportados:

* `policy_id` - (Obrigatório) O ID da política a ser associada.
* `group_id` - (Obrigatório) O ID do grupo (de usuários ou sistemas) ao qual a política será associada.
* `type` - (Obrigatório) O tipo de grupo. Valores válidos: `user_group` ou `system_group`.

## Referência de Atributos

Além dos argumentos acima, os seguintes atributos são exportados:

* `id` - O ID da associação entre política e grupo, no formato `policy_id:group_id:type`.

## Importação

Associações de políticas podem ser importadas usando o ID composto no formato `policy_id:group_id:type`:

```
terraform import jumpcloud_policy_association.example 5f0c1b2c3d4e5f6g7h8i9j0k:6a7b8c9d0e1f2g3h4i5j6k7l:user_group
```

## Cenários de Uso Comuns

### Aplicação de Múltiplas Políticas

```hcl
# Política de MFA para todos os colaboradores
resource "jumpcloud_policy" "mfa_policy" {
  name        = "Required MFA Policy"
  description = "Política global de MFA para todos os usuários"
  type        = "mfa"
  active      = true
  
  configurations = {
    allow_totp_enrollment      = "true"
    require_mfa_for_all_users  = "true"
  }
}

# Associar política de MFA a vários grupos
resource "jumpcloud_user_group" "it" {
  name = "IT Department"
}

resource "jumpcloud_user_group" "executives" {
  name = "Executive Team"
}

resource "jumpcloud_policy_association" "it_mfa" {
  policy_id = jumpcloud_policy.mfa_policy.id
  group_id  = jumpcloud_user_group.it.id
  type      = "user_group"
}

resource "jumpcloud_policy_association" "executives_mfa" {
  policy_id = jumpcloud_policy.mfa_policy.id
  group_id  = jumpcloud_user_group.executives.id
  type      = "user_group"
}
```

### Gerenciamento Condicional

```hcl
# Verificar se a política já está associada ao grupo
data "jumpcloud_user_group" "existing_group" {
  name = "Finance Department"
}

data "jumpcloud_policy" "existing_policy" {
  name = "Secure Password Policy"
}

# Buscar todas as políticas associadas (exemplo fictício)
data "jumpcloud_policy_associations" "existing_associations" {
  group_id = data.jumpcloud_user_group.existing_group.id
  type     = "user_group"
}

locals {
  # Verificar se a política já está associada
  policy_already_associated = contains(data.jumpcloud_policy_associations.existing_associations.policy_ids, data.jumpcloud_policy.existing_policy.id)
}

# Criar associação apenas se não existir
resource "jumpcloud_policy_association" "conditional_association" {
  count = local.policy_already_associated ? 0 : 1
  
  policy_id = data.jumpcloud_policy.existing_policy.id
  group_id  = data.jumpcloud_user_group.existing_group.id
  type      = "user_group"
}
```

## Considerações de Segurança

* Associe políticas de segurança críticas a grupos específicos para garantir que apenas os usuários apropriados sejam afetados.
* Ao associar políticas de MFA ou de senha complexa, considere o impacto na experiência do usuário e prepare comunicações adequadas.
* Para políticas que afetam sistemas, verifique a compatibilidade antes de aplicá-las em ambientes de produção.
* Considere implementar políticas em fases, começando com grupos menores antes de expandir para toda a organização.
* Implemente um processo de revisão regular das associações de políticas para garantir que permaneçam adequadas às necessidades de segurança. 