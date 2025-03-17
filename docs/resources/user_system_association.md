# jumpcloud_user_system_association Resource

Gerencia associações entre usuários e sistemas no JumpCloud. Este recurso permite criar e excluir vínculos entre usuários e sistemas, controlando quais usuários têm acesso a quais sistemas.

## Referência da API JumpCloud

Para mais detalhes sobre a API subjacente, consulte:
- [API JumpCloud - Associações de Usuários a Sistemas](https://docs.jumpcloud.com/api/2.0/index.html#tag/user-associations)

## Considerações de Segurança

- Utilize associações para implementar o princípio de menor privilégio, garantindo que os usuários tenham acesso apenas aos sistemas necessários para suas funções.
- Audite regularmente as associações para identificar e remover acessos desnecessários ou obsoletos.
- Considere usar grupos de usuários para gerenciar associações em escala, em vez de associar usuários individualmente.
- As associações criadas por este recurso serão refletidas nas permissões de login do usuário para o sistema.

## Exemplos de Uso

### Associação Básica entre Usuário e Sistema

```hcl
resource "jumpcloud_user_system_association" "basic_association" {
  user_id   = jumpcloud_user.example.id
  system_id = jumpcloud_system.web_server.id
}
```

### Gerenciamento de Acesso para uma Equipe

```hcl
# Criar usuários para a equipe
resource "jumpcloud_user" "team_lead" {
  username  = "team.lead"
  email     = "team.lead@example.com"
  firstname = "Team"
  lastname  = "Lead"
  password  = "SecurePassword123!"
}

resource "jumpcloud_user" "team_member" {
  username  = "team.member"
  email     = "team.member@example.com"
  firstname = "Team"
  lastname  = "Member"
  password  = "SecurePassword456!"
}

# Configurar sistemas
resource "jumpcloud_system" "production_server" {
  display_name = "production-server"
  # ... outras configurações
}

resource "jumpcloud_system" "development_server" {
  display_name = "development-server"
  # ... outras configurações
}

# Associar usuários a sistemas
resource "jumpcloud_user_system_association" "lead_prod_access" {
  user_id   = jumpcloud_user.team_lead.id
  system_id = jumpcloud_system.production_server.id
}

resource "jumpcloud_user_system_association" "lead_dev_access" {
  user_id   = jumpcloud_user.team_lead.id
  system_id = jumpcloud_system.development_server.id
}

resource "jumpcloud_user_system_association" "member_dev_access" {
  user_id   = jumpcloud_user.team_member.id
  system_id = jumpcloud_system.development_server.id
}
```

### Uso com Dados de Usuários e Sistemas Existentes

```hcl
# Buscar um usuário existente
data "jumpcloud_user" "existing_user" {
  email = "existing.user@example.com"
}

# Buscar um sistema existente
data "jumpcloud_system" "existing_system" {
  display_name = "existing-server"
}

# Associar o usuário ao sistema
resource "jumpcloud_user_system_association" "existing_association" {
  user_id   = data.jumpcloud_user.existing_user.id
  system_id = data.jumpcloud_system.existing_system.id
}
```

## Referência de Argumentos

Os seguintes argumentos são suportados:

* `user_id` - (Obrigatório) O ID do usuário JumpCloud a ser associado. Este valor não pode ser alterado após a criação.
* `system_id` - (Obrigatório) O ID do sistema JumpCloud a ser associado. Este valor não pode ser alterado após a criação.

## Referência de Atributos

Além dos argumentos acima, os seguintes atributos são exportados:

* `id` - Um identificador composto no formato `user_id:system_id` que representa esta associação.

## Importação

Associações existentes entre usuário e sistema podem ser importadas usando um ID composto no formato `user_id:system_id`, por exemplo:

```bash
terraform import jumpcloud_user_system_association.example 5f0c1b2c3d4e5f6g7h8i9j0k:6a7b8c9d0e1f2g3h4i5j6k7l
```

## Melhores Práticas

1. **Gerenciamento de Acesso**:
   - Documente claramente por que cada associação existe para facilitar futuras auditorias.
   - Implemente um processo de revisão regular para garantir que as associações permaneçam necessárias.

2. **Escalabilidade**:
   - Para grandes equipes, considere usar grupos de usuários e associações de grupo-sistema em vez de associações individuais.
   - Utilize módulos Terraform para gerenciar conjuntos comuns de associações.

3. **Segurança**:
   - Combine com configurações do sistema que exigem MFA para login em sistemas críticos.
   - Implemente automação para revogar acesso quando usuários mudam de função ou saem da organização.

4. **Monitoramento**:
   - Configure alertas para criação ou remoção não autorizada de associações críticas.
   - Mantenha registros históricos de mudanças nas associações para fins de auditoria. 