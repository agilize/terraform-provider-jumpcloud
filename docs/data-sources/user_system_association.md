# jumpcloud_user_system_association Data Source

Use este data source para verificar se existe uma associação entre um usuário e um sistema específicos no JumpCloud.

## Exemplo de Uso

```hcl
# Verificar se um usuário está associado a um sistema
data "jumpcloud_user_system_association" "check_access" {
  user_id   = "5f0c1b2c3d4e5f6g7h8i9j0k"  # ID do usuário
  system_id = "6a7b8c9d0e1f2g3h4i5j6k7l"  # ID do sistema
}

# Uso com dados de usuários e sistemas existentes
data "jumpcloud_user" "existing_user" {
  email = "existing.user@example.com"
}

data "jumpcloud_system" "existing_system" {
  display_name = "existing-server"
}

data "jumpcloud_user_system_association" "check_specific_access" {
  user_id   = data.jumpcloud_user.existing_user.id
  system_id = data.jumpcloud_system.existing_system.id
}

# Verificação condicional baseada na associação
output "user_access_status" {
  value = data.jumpcloud_user_system_association.check_access.associated ? "Usuário tem acesso ao sistema" : "Usuário NÃO tem acesso ao sistema"
}

# Uso em lógica condicional
locals {
  needs_access_grant = !data.jumpcloud_user_system_association.check_specific_access.associated
}

# Criar a associação apenas se ela não existir
resource "jumpcloud_user_system_association" "conditional_association" {
  count     = local.needs_access_grant ? 1 : 0
  user_id   = data.jumpcloud_user.existing_user.id
  system_id = data.jumpcloud_system.existing_system.id
}
```

## Casos de Uso Comuns

1. **Verificação de Acesso**: Verifique se um usuário específico tem acesso a um sistema antes de realizar outras operações.
2. **Auditoria de Segurança**: Use para gerar relatórios sobre quais usuários têm acesso a sistemas críticos.
3. **Automação Condicional**: Implemente lógica condicional para criar ou remover associações com base no estado atual.
4. **Validação de Configuração**: Verifique se as associações esperadas existem em um ambiente gerenciado manualmente ou por múltiplas ferramentas.

## Referência de Argumentos

Os seguintes argumentos são suportados:

* `user_id` - (Obrigatório) O ID do usuário JumpCloud para verificar a associação.
* `system_id` - (Obrigatório) O ID do sistema JumpCloud para verificar a associação.

## Referência de Atributos

Além dos argumentos acima, os seguintes atributos são exportados:

* `id` - Um identificador composto no formato `user_id:system_id` para referência.
* `associated` - Um valor booleano que indica se o usuário está associado ao sistema (true) ou não (false).

## Melhores Práticas

1. **Uso para Lógica Condicional**:
   - Use este data source para implementar lógica condicional em suas configurações Terraform.
   - Evite criar associações duplicadas verificando primeiro se elas já existem.

2. **Combinação com Recursos**:
   - Use em conjunto com o recurso `jumpcloud_user_system_association` para implementar uma abordagem idempotente para o gerenciamento de associações.

3. **Eficiência**:
   - Minimize o número de consultas à API agrupando verificações relacionadas.
   - Use o data source apenas quando necessário, especialmente em configurações grandes.

4. **Tratamento de Erros**:
   - Implemente verificações adicionais para garantir que os IDs de usuário e sistema sejam válidos antes de verificar as associações.
   - Considere a utilização de blocos count ou for_each para lidar com múltiplas verificações. 