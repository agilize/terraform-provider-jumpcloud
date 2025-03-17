# Estratégia de Versionamento

Este repositório adota o [Versionamento Semântico](https://semver.org/lang/pt-BR/) com algumas adaptações para melhor gerenciamento de branches e releases.

## Formato da Versão

Todas as versões seguem o formato:

```
MAJOR.MINOR.PATCH[-sufixo]
```

Onde:
- **MAJOR**: Incrementado quando há mudanças incompatíveis com versões anteriores
- **MINOR**: Incrementado quando há adição de funcionalidades de forma compatível
- **PATCH**: Incrementado quando há correções de bugs compatíveis
- **sufixo** (opcional): Identifica versões especiais (beta, rc, etc.)

## Branches e Ciclo de Vida

O repositório segue o seguinte fluxo de trabalho:

### Branch `develop`

- Todas as funcionalidades em desenvolvimento devem ser enviadas para a branch `develop`
- Versões construídas a partir desta branch recebem o sufixo `-beta`
- O número de versão é incrementado automaticamente (PATCH)
- Exemplo: `v0.1.0-beta`

### Branch `main`

- Contém o código estável para produção
- Versões construídas a partir desta branch não têm sufixo
- Versões são promovidas a partir da branch `develop`, removendo o sufixo `-beta`
- Exemplo: `v0.1.0`

### Pull Requests

- Pull Requests para a branch `develop` recebem um número de versão temporário com o sufixo `-prX` (onde X é o número do PR)
- Exemplo: `v0.1.0-pr42`

## Incremento Automático de Versão

O sistema incrementa automaticamente o número da versão PATCH cada vez que um novo código é enviado para as branches principais:

1. A versão base é determinada pelo último tag no repositório
2. O número PATCH é incrementado em 1
3. Sufixos são adicionados conforme a branch alvo

## Publicação de Pacotes

Todos os pacotes são publicados no GitHub Packages:

- **Pacotes Beta**: Para testes e validação, compilados da branch `develop`
- **Pacotes Estáveis**: Para uso em produção, compilados da branch `main`

Para usar uma versão específica, configure seu Terraform:

```hcl
terraform {
  required_providers {
    jumpcloud = {
      source  = "github.com/ferreirafav/jumpcloud"
      version = "0.1.0" # ou "0.1.0-beta" para versões beta
    }
  }
}
```

## Changelog

O CHANGELOG.md é atualizado automaticamente para versões estáveis (não-beta). As entradas são organizadas por:

- Features
- Bug Fixes
- Documentation
- Other Changes

## Convenções de Commit

Para facilitar a geração automática do changelog, seguimos as convenções de commit:

- `feat:` Novas funcionalidades
- `fix:` Correções de bugs
- `docs:` Alterações em documentação
- `chore:` Manutenção do repositório, sem alterações no código principal
- `test:` Adição ou modificação de testes
- `refactor:` Refatoração de código sem alteração de comportamento
- `style:` Alterações que não afetam o comportamento do código (formatação, espaços em branco, etc.)
- `perf:` Melhorias de performance
- `ci:` Alterações nos scripts de CI/CD

## Notas Adicionais

- Para alterar manualmente a versão MAJOR ou MINOR, crie uma tag manualmente e envie-a para o repositório.
- O sistema sempre respeita tags existentes e constrói a próxima versão a partir delas. 