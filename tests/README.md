# Guia de Testes do Provider Terraform JumpCloud

Este documento fornece informações sobre a estratégia de testes e como escrever novos testes para o provider Terraform JumpCloud.

## Estrutura de Testes

O provider JumpCloud utiliza vários tipos de testes:

1. **Testes Unitários**: Testam funções específicas isoladamente
2. **Testes de Integração**: Testam a interação entre componentes
3. **Testes de Aceitação**: Testam o comportamento real do provider contra a API JumpCloud
4. **Testes de Segurança**: Verificam se dados sensíveis são tratados adequadamente
5. **Testes de Performance**: Avaliam o desempenho de operações críticas

## Tipos de Arquivos de Teste

Para cada recurso e data source, existem arquivos de teste correspondentes:

- `resource_*_test.go`: Testes para recursos (CRUD)
- `data_source_*_test.go`: Testes para data sources (Read)
- `provider_*_test.go`: Testes gerais do provider

## Ferramentas de Teste

O provider utiliza as seguintes ferramentas de teste:

- **testify**: Para assertions e mocking
- **terraform-plugin-sdk/v2/helper/resource**: Para testes de aceitação
- **terraform-plugin-sdk/v2/helper/schema**: Para testes de schema

## Como Executar os Testes

### Testes Unitários

```bash
# Executar todos os testes unitários
go test ./internal/provider -v

# Executar testes específicos
go test ./internal/provider -v -run "TestResourceUser"
```

### Testes de Aceitação

Testes de aceitação exigem credenciais reais do JumpCloud:

```bash
# Configurar variáveis de ambiente
export TF_ACC=1
export JUMPCLOUD_API_KEY="sua-chave-api"
export JUMPCLOUD_ORG_ID="seu-org-id"

# Executar testes de aceitação
go test ./internal/provider -v -run "TestAcc"
```

## Padrões de Teste

### Teste de Recursos

Para cada recurso, temos testes para todas as operações CRUD:

1. **Create**: Testa a criação de um novo recurso
2. **Read**: Testa a leitura de um recurso existente
3. **Update**: Testa a atualização de um recurso
4. **Delete**: Testa a exclusão de um recurso

### Teste de Data Sources

Para cada data source, temos testes para:

1. **Busca por ID**: Testa a recuperação de dados por ID
2. **Busca por nome/identificador**: Testa a recuperação de dados por nome ou outro identificador
3. **Erros e validações**: Testa o comportamento com parâmetros incorretos ou ausentes

## Validação de Parâmetros

### Diretrizes para Validação

Para todos os recursos e data sources, seguimos estas diretrizes de validação:

1. **Validação Prévia**: Sempre validar parâmetros antes de fazer chamadas à API
2. **Mensagens Claras**: Fornecer mensagens de erro claras e específicas
3. **Validação de Campos Obrigatórios**: Verificar se todos os campos obrigatórios estão presentes
4. **Validação de Formato**: Verificar se os campos estão no formato correto
5. **Validação de Valores**: Verificar se os valores estão dentro dos limites esperados

### Exemplo de Validação

```go
// Validação de parâmetros antes de fazer chamadas à API
if userID == "" {
    return diag.FromErr(fmt.Errorf("user_id não pode ser vazio"))
}

if systemID == "" {
    return diag.FromErr(fmt.Errorf("system_id não pode ser vazio"))
}

// Validação de formato
if !isValidUUID(id) {
    return diag.FromErr(fmt.Errorf("id não é um UUID válido: %s", id))
}

// Validação de valores
if len(password) < 8 {
    return diag.FromErr(fmt.Errorf("a senha deve ter pelo menos 8 caracteres"))
}
```

### Testes de Validação

Para cada validação, criamos testes específicos:

```go
// TestResourceExample_EmptyName testa a validação de nome vazio
func TestResourceExample_EmptyName(t *testing.T) {
    mockClient := new(MockClient)
    
    // Configuração do resource com nome vazio
    d := schema.TestResourceDataRaw(t, resourceExample().Schema, nil)
    
    // Executar função
    diags := resourceExampleCreate(context.Background(), d, mockClient)
    
    // Verificar resultados
    assert.True(t, diags.HasError())
    assert.Contains(t, diags[0].Summary, "nome não pode ser vazio")
    mockClient.AssertExpectations(t) // Garantir que nenhuma chamada API foi feita
}
```

## Mock Client

O provider utiliza um cliente mock para simular as chamadas à API JumpCloud nos testes unitários. O mock client é configurado para responder a diferentes chamadas de API com respostas predefinidas:

```go
mockClient := new(MockClient)
mockClient.On("DoRequest", http.MethodGet, "/api/v2/users/test-id", nil).Return([]byte(`{"_id":"test-id"}`), nil)
```

## Funções Auxiliares para Testes

Várias funções auxiliares estão disponíveis:

- `testAccPreCheck`: Verifica se as variáveis de ambiente estão configuradas
- `testAccProvider`: Retorna uma instância do provider para testes
- `testAccProviderFactories`: Retorna as factories do provider para testes
- `testAccCheckResourceDestroy`: Verifica se um recurso foi destruído
- `testAccCheckResourceExists`: Verifica se um recurso existe

## Escrevendo Novos Testes

### 1. Testes Unitários

```go
func TestResourceExampleCreate(t *testing.T) {
    mockClient := new(MockClient)
    
    // Configurar mock response
    mockResponse := map[string]interface{}{
        "_id": "test-id",
        "name": "test-name",
    }
    responseBytes, _ := json.Marshal(mockResponse)
    mockClient.On("DoRequest", http.MethodPost, "/api/path", mock.Anything).Return(responseBytes, nil)
    
    // Configurar resource data
    d := schema.TestResourceDataRaw(t, resourceExample().Schema, nil)
    d.Set("name", "test-name")
    
    // Executar função
    diags := resourceExampleCreate(context.Background(), d, mockClient)
    
    // Verificar resultados
    assert.False(t, diags.HasError())
    assert.Equal(t, "test-id", d.Id())
    mockClient.AssertExpectations(t)
}
```

### 2. Testes de Aceitação

```go
func TestAccJumpCloudExample_basic(t *testing.T) {
    if !testAccPreCheck(t) {
        return
    }
    
    resource.Test(t, resource.TestCase{
        PreCheck:          func() { testAccPreCheck(t) },
        ProviderFactories: testAccProviderFactories(),
        CheckDestroy:      testAccCheckJumpCloudExampleDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccJumpCloudExampleConfig(),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckJumpCloudExampleExists("jumpcloud_example.test"),
                    resource.TestCheckResourceAttr("jumpcloud_example.test", "name", "test-name"),
                ),
            },
        },
    })
}
```

## Boas Práticas

1. **Mantenha os testes independentes**: Cada teste deve ser independente dos outros
2. **Use nomes descritivos**: Nomeie os testes para refletir claramente o que está sendo testado
3. **Verifique todos os cenários**: Teste tanto os casos de sucesso quanto os de erro
4. **Minimize o uso de APIs reais**: Use mocks quando possível para testes unitários
5. **Limpe após os testes**: Garanta que os recursos criados em testes sejam removidos
6. **Mantenha a cobertura de código**: Todos os recursos e data sources devem ter testes

## Verificando a Cobertura de Testes

```bash
go test ./internal/provider -cover
```

Para um relatório mais detalhado:

```bash
go test ./internal/provider -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Recursos Adicionais

- [Documentação de Testes do Go](https://golang.org/pkg/testing/)
- [Documentação do Testify](https://pkg.go.dev/github.com/stretchr/testify)
- [Testes de Aceitação do Terraform](https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests)

## Recomendações para Implementação

Para finalizar as implementações solicitadas, você pode seguir o padrão demonstrado nos exemplos acima para os demais recursos. Aqui estão algumas considerações importantes:

1. **Reutilização de código**: Para implementações similares como `jumpcloud_application_group_mapping`, você pode adaptar o código do `jumpcloud_application_user_mapping` alterando apenas os endpoints e estruturas de dados.

2. **Testes**: Crie testes unitários para cada recurso, seguindo o padrão já estabelecido no projeto. Isso garante que as implementações funcionem corretamente.

3. **Documentação**: Para cada recurso e data source, crie documentação detalhada com exemplos de uso, explicação dos argumentos e instruções de importação.

4. **Validação de API**: Antes de implementar cada recurso, verifique a documentação da API JumpCloud para confirmar o formato correto dos endpoints e payload.

## Plano de Implementação em Fases

Para implementar todos os recursos solicitados de forma organizada, recomendo seguir este cronograma:

### Semana 1: Recursos SSO
- Implementar `jumpcloud_application` e data source correspondente
- Implementar `jumpcloud_application_user_mapping`
- Implementar `jumpcloud_application_group_mapping`

### Semana 2: Recursos RADIUS e MFA
- Implementar `jumpcloud_radius_server` e data source correspondente
- Implementar `jumpcloud_mfa_settings`
- Implementar `jumpcloud_mfa_policy`

### Semana 3: Recursos de Integração
- Implementar `jumpcloud_webhook` e data source correspondente
- Implementar `jumpcloud_api_key` e data source correspondente

### Semana 4: Recursos Avançados
- Implementar `jumpcloud_organization` e data source correspondente
- Implementar `jumpcloud_device_trust`
- Implementar `jumpcloud_network_source` e data source correspondente

Cada implementação deve incluir:
1. Arquivo de código Go com implementação completa
2. Testes unitários
3. Documentação detalhada
4. Registro no provider.go

Após a conclusão de cada recurso, recomendo executar testes detalhados contra a API real do JumpCloud para garantir que todas as operações CRUD funcionem conforme esperado. 