# Terraform Provider for JumpCloud

[![Go Report Card](https://goreportcard.com/badge/github.com/ferreirafav/terraform-provider-jumpcloud)](https://goreportcard.com/report/github.com/ferreirafav/terraform-provider-jumpcloud)

Este provider Terraform permite gerenciar recursos do JumpCloud através do Terraform, possibilitando a automação de processos como:

- Gerenciamento de usuários JumpCloud (criação, atualização, deleção)
- Gerenciamento de sistemas JumpCloud (configuração, atualização)
- Recuperação de informações sobre usuários e sistemas para integração com outros recursos 

O provider utiliza a API oficial do JumpCloud para realizar todas as operações, garantindo consistência e segurança na gestão dos recursos.

## Funcionalidades

### Recursos (Resources)

| Nome | Descrição |
|------|-----------|
| `jumpcloud_user` | Gerencia usuários no JumpCloud, permitindo definir atributos, configurações de MFA e senhas |
| `jumpcloud_system` | Gerencia sistemas (dispositivos) no JumpCloud, configurando atributos, tags e políticas de segurança |

### Fontes de Dados (Data Sources)

| Nome | Descrição |
|------|-----------|
| `jumpcloud_user` | Recupera informações detalhadas sobre usuários do JumpCloud |
| `jumpcloud_system` | Recupera informações sobre sistemas registrados no JumpCloud |

## Requisitos

- [Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
- [Go](https://golang.org/doc/install) >= 1.18 (apenas para desenvolvimento)
- Uma conta JumpCloud com acesso à API

## Instalação

### Usando o Terraform Registry (recomendado)

Adicione a configuração abaixo ao seu arquivo Terraform:

```hcl
terraform {
  required_providers {
    jumpcloud = {
      source  = "ferreirafav/jumpcloud"
      version = "0.1.0"
    }
  }
}
```

### Instalação manual (para desenvolvimento)

1. Clone o repositório
   ```sh
   git clone https://github.com/ferreirafav/terraform-provider-jumpcloud.git
   ```

2. Entre no diretório do repositório
   ```sh
   cd terraform-provider-jumpcloud
   ```

3. Compile o provider
   ```sh
   go build -o terraform-provider-jumpcloud
   ```

4. Instale o provider local
   ```sh
   make install
   ```

## Configuração

Para usar o provider, adicione a seguinte configuração Terraform:

```hcl
provider "jumpcloud" {
  api_key = "your_api_key"  # ou use a variável de ambiente JUMPCLOUD_API_KEY
  org_id  = "your_org_id"   # ou use a variável de ambiente JUMPCLOUD_ORG_ID
  # api_url = "https://console.jumpcloud.com/api"  # opcional, este é o valor padrão
}
```

### Variáveis de ambiente

Para maior segurança, recomendamos o uso de variáveis de ambiente:

```sh
export JUMPCLOUD_API_KEY="sua_api_key"
export JUMPCLOUD_ORG_ID="seu_org_id"
```

## Exemplos de Uso

### Gerenciando um usuário

```hcl
resource "jumpcloud_user" "devops_user" {
  username    = "devops.user"
  email       = "devops.user@exemplo.com"
  firstname   = "DevOps"
  lastname    = "User"
  password    = "P@ssw0rd_segura!123"
  description = "Usuário DevOps gerenciado pelo Terraform"
  
  attributes = {
    department = "TI"
    role       = "DevOps Engineer"
    location   = "Remoto"
  }
  
  mfa_enabled = true
  password_never_expires = false
}
```

### Consultando informações de um sistema

```hcl
data "jumpcloud_system" "web_server" {
  display_name = "web-server-prod"
}

output "sistema_info" {
  value = {
    id              = data.jumpcloud_system.web_server.id
    sistema_tipo    = data.jumpcloud_system.web_server.system_type
    sistema_os      = data.jumpcloud_system.web_server.os
    sistema_versao  = data.jumpcloud_system.web_server.version
    sistema_tags    = data.jumpcloud_system.web_server.tags
  }
}
```

## Documentação Completa

A documentação de cada recurso e fonte de dados está disponível nos seguintes arquivos:

- [Recurso de Usuário](docs/resources/user.md)
- [Recurso de Sistema](docs/resources/system.md)
- [Fonte de Dados de Usuário](docs/data-sources/user.md)
- [Fonte de Dados de Sistema](docs/data-sources/system.md)

Os exemplos completos de uso podem ser encontrados no diretório [examples](examples/).

## Desenvolvimento

### Pré-requisitos

Para contribuir com o desenvolvimento do provider, você precisará:

- [Go](http://www.golang.org) versão 1.18 ou superior
- [Terraform](https://www.terraform.io/downloads.html) versão 0.13 ou superior
- Uma conta JumpCloud para testes

### Compilação

Para compilar o provider:

```sh
go build -o terraform-provider-jumpcloud
```

### Testes

O provider inclui vários tipos de testes:

```sh
# Testes unitários
make test-unit

# Testes de integração
make test-integration

# Testes de aceitação
make test-acceptance

# Testes de segurança
make test-security

# Testes de performance
make test-performance
```

### Gerando Documentação

Para gerar ou atualizar a documentação:

```sh
make docs
```

## Links Úteis

- [Documentação da API JumpCloud](https://docs.jumpcloud.com/api)
- [Terraform Registry](https://registry.terraform.io/providers/ferreirafav/jumpcloud/latest)
- [Terraform Documentation](https://www.terraform.io/docs)

## Licença

Este provider é licenciado sob a licença MIT. 