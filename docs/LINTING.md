# Linting no Provider JumpCloud

Este documento descreve os erros de linting identificados no projeto e o plano para corrigi-los gradualmente.

## Visão Geral

Estamos usando a ferramenta `tfproviderlint` para verificar a conformidade do código com as melhores práticas para providers Terraform. Alguns erros foram identificados e serão corrigidos em fases.

## Erros Corrigidos

- ✅ **R014**: Parâmetros do tipo `interface{}` devem ser chamados de `meta`. Este erro foi corrigido em todos os arquivos.

## Erros Pendentes

Os seguintes erros de linting serão corrigidos em fases futuras:

### Erros em Testes de Aceitação

- **AT001**: Missing CheckDestroy - testes de aceitação devem incluir uma verificação de destruição para garantir que os recursos sejam limpos corretamente.
- **AT005**: Nomes de funções de teste de aceitação devem começar com `TestAcc`.
- **AT012**: Arquivo contém múltiplos prefixos de nomes para testes de aceitação, o que pode causar confusão.

### Erros em Recursos

- **R001**: O argumento de chave para `ResourceData.Set()` deve ser uma string literal, não uma variável.
- **R017**: Atributos de schema devem ser estáveis entre execuções do Terraform para evitar problemas de estado.
- **R019**: `d.HasChanges()` tem muitos argumentos, considere usar `d.HasChangesExcept()`.

### Erros de Validação

- **V013**: Funções de validação customizadas devem ser substituídas por `validation.StringInSlice()` ou `validation.StringNotInSlice()`.

## Plano de Correção

Para facilitar o processo de correção, seguiremos a seguinte ordem:

1. **Fase 1**: Corrigir R001 (ResourceData.Set com string literal)
2. **Fase 2**: Corrigir R019 (HasChanges → HasChangesExcept)
3. **Fase 3**: Corrigir V013 (SchemaValidateFunc → validation.StringInSlice)
4. **Fase 4**: Corrigir R017 (Schema attributes should be stable)
5. **Fase 5**: Corrigir AT* (problemas em testes de aceitação)

## Configuração do CI/CD

Para evitar que os erros de linting bloqueiem o desenvolvimento enquanto são corrigidos gradualmente, implementamos as seguintes soluções:

### GitHub Actions

O workflow de verificação de pull requests (`.github/workflows/pr-check.yml`) foi configurado para:

1. Executar a verificação do tfproviderlint apenas para erros críticos e R014 (já corrigido)
2. Ignorar temporariamente os erros que serão corrigidos em fases
3. Listar os erros pendentes para referência

À medida que cada fase de correção for concluída, o workflow será atualizado para habilitar a verificação das regras correspondentes.

### Scripts Locais

Fornecemos dois scripts para ajudar no processo de verificação local:

- `check_required_lint.sh`: Verifica apenas os erros críticos, ignorando os que serão tratados em fases.
- `run_linter.sh`: Fornece opções para verificar erros específicos e informações sobre como executar o linter.

## Como Contribuir para Correções

Se você deseja contribuir para corrigir erros de linting, siga estas etapas:

1. Escolha uma fase para trabalhar com base no plano de correção.
2. Execute o lint específico para a regra que está corrigindo:
   ```
   $HOME/go/bin/tfproviderlint -AT=false -R=false -S=false -V=false -<REGRA>=true ./...
   ```
3. Faça as correções necessárias nos arquivos indicados.
4. Execute os testes para garantir que suas alterações não causaram regressões.
5. Envie um PR com uma descrição clara das correções realizadas.

## Detalhes dos Erros de Linting

### R001: ResourceData.Set() com string literal

```go
// Incorreto
key := "attribute_name"
d.Set(key, value)

// Correto
d.Set("attribute_name", value)
```

### R019: HasChanges → HasChangesExcept

```go
// Incorreto
if d.HasChanges("attr1", "attr2", "attr3", "attr4", "attr5") {
    // ...
}

// Correto
if d.HasChangesExcept("attr6", "attr7") {
    // ...
}
```

### V013: SchemaValidateFunc → validation.StringInSlice

```go
// Incorreto
ValidateFunc: func(v interface{}, k string) (warns []string, errs []error) {
    value := v.(string)
    validValues := []string{"one", "two", "three"}
    valid := false
    for _, val := range validValues {
        if value == val {
            valid = true
            break
        }
    }
    if !valid {
        errs = append(errs, fmt.Errorf("%s must be one of %v, got: %s", k, validValues, value))
    }
    return
},

// Correto
ValidateFunc: validation.StringInSlice([]string{"one", "two", "three"}, false),
``` 