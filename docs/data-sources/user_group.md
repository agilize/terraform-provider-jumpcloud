# jumpcloud_user_group Data Source

Use este data source para obter informações sobre um grupo de usuários existente no JumpCloud.

## Exemplo de Uso

```hcl
# Buscar um grupo de usuários por nome
data "jumpcloud_user_group" "by_name" {
  name = "developers"
}

# Buscar um grupo de usuários por ID
data "jumpcloud_user_group" "by_id" {
  id = "5f0c1b2c3d4e5f6g7h8i9j0k"
}

# Uso das informações do grupo
output "group_details" {
  value = {
    id          = data.jumpcloud_user_group.by_name.id
    name        = data.jumpcloud_user_group.by_name.name
    description = data.jumpcloud_user_group.by_name.description
    attributes  = data.jumpcloud_user_group.by_name.attributes
  }
}

# Exemplo de uso em conjunto com outro recurso
resource "jumpcloud_user" "new_member" {
  username  = "new.developer"
  email     = "new.developer@example.com"
  firstname = "New"
  lastname  = "Developer"
  password  = "SecurePassword123!"
  
  attributes = {
    department = "Engineering"
    group_id   = data.jumpcloud_user_group.by_name.id
  }
}
```

## Casos de Uso Comuns

1. **Referência para Associações**: Use para associar usuários a grupos existentes.
2. **Auditoria e Relatórios**: Recupere informações sobre grupos para fins de auditoria.
3. **Automação**: Use em scripts de automação para gerenciar permissões baseadas em grupo.
4. **Integrações**: Integre com outros sistemas que precisam de informações sobre grupos JumpCloud.

## Referência de Argumentos

Os seguintes argumentos são suportados:

* `id` - (Opcional) O ID do grupo de usuários a ser recuperado.
* `name` - (Opcional) O nome do grupo de usuários a ser recuperado.

**Nota:** Você deve especificar exatamente um dos argumentos acima. O data source retornará um erro se nenhum ou ambos forem fornecidos.

## Referência de Atributos

Além dos argumentos acima, os seguintes atributos são exportados:

* `id` - O ID único do grupo de usuários no JumpCloud.
* `name` - O nome do grupo de usuários.
* `description` - A descrição do grupo de usuários.
* `type` - O tipo do grupo.
* `attributes` - Um mapa de atributos personalizados associados ao grupo.
* `created` - A data de criação do grupo.

## Melhores Práticas

1. **Identifique Grupos Consistentemente**:
   - Utilize nomes de grupo para referência quando a legibilidade for importante.
   - Use IDs de grupo para referência quando a estabilidade for crítica (os nomes podem mudar, os IDs não).

2. **Uso Eficiente**:
   - Reutilize o mesmo data source quando precisar referenciar o mesmo grupo em vários lugares.
   - Evite consultas excessivas à API especificando exatamente apenas o grupo que você precisa.

3. **Tratamento de Erros**:
   - Verifique se o grupo existe antes de tentar referenciá-lo em outros recursos.
   - Use a saída `id` como verificação de existência em suas configurações condicionais.

4. **Documentação**:
   - Adicione comentários explicando por que um grupo específico é referenciado em sua configuração.
   - Documente quaisquer dependências em outros recursos que usam informações do grupo. 