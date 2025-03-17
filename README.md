# Terraform Provider for JumpCloud

[![Go Report Card](https://goreportcard.com/badge/github.com/ferreirafav/terraform-provider-jumpcloud)](https://goreportcard.com/report/github.com/ferreirafav/terraform-provider-jumpcloud)

Este provider Terraform permite gerenciar recursos do JumpCloud através do Terraform, possibilitando a automação de processos como:

- Gerenciamento de usuários JumpCloud (criação, atualização, deleção)
- Gerenciamento de sistemas JumpCloud (configuração, atualização)
- Gerenciamento de grupos de usuários (criação, atualização, deleção)
- Gerenciamento de associações entre usuários e sistemas (criação, deleção)
- Gerenciamento de políticas e associações de políticas a grupos
- Recuperação de informações sobre usuários, sistemas, grupos e associações para integração com outros recursos 

O provider utiliza a API oficial do JumpCloud para realizar todas as operações, garantindo consistência e segurança na gestão dos recursos.

## Funcionalidades

### Recursos (Resources)

| Nome | Descrição |
|------|-----------|
| `jumpcloud_user` | Gerencia usuários no JumpCloud, permitindo definir atributos, configurações de MFA e senhas |
| `jumpcloud_system` | Gerencia sistemas no JumpCloud, permitindo configurar tags, descrições e configurações de segurança |
| `jumpcloud_user_group` | Gerencia grupos de usuários no JumpCloud |
| `jumpcloud_system_group` | Gerencia grupos de sistemas no JumpCloud |
| `jumpcloud_user_group_membership` | Gerencia a associação de usuários a grupos de usuários |
| `jumpcloud_system_group_membership` | Gerencia a associação de sistemas a grupos de sistemas |
| `jumpcloud_user_system_association` | Gerencia associações entre usuários e sistemas, controlando acesso direto |
| `jumpcloud_command` | Gerencia comandos no JumpCloud, permitindo executar scripts nos sistemas |
| `jumpcloud_command_association` | Gerencia associações entre comandos e sistemas ou grupos |
| `jumpcloud_policy` | Gerencia políticas no JumpCloud, como políticas de senha, MFA e bloqueio de conta |
| `jumpcloud_policy_association` | Gerencia associações entre políticas e grupos de usuários ou sistemas |

### Fontes de Dados (Data Sources)

| Nome | Descrição |
|------|-----------|
| `jumpcloud_user` | Obtém informações sobre usuários existentes no JumpCloud |
| `jumpcloud_system` | Obtém informações sobre sistemas existentes no JumpCloud |
| `jumpcloud_user_group` | Obtém informações sobre grupos de usuários existentes no JumpCloud |
| `jumpcloud_system_group` | Obtém informações sobre grupos de sistemas existentes no JumpCloud |
| `jumpcloud_user_system_association` | Verifica a associação entre um usuário e um sistema |
| `jumpcloud_command` | Obtém informações sobre comandos existentes no JumpCloud |
| `jumpcloud_policy` | Obtém informações sobre políticas existentes no JumpCloud |

## Requisitos

- [Terraform](https://www.terraform.io/downloads.html) 0.13.x ou superior
- Go 1.20 ou superior (para desenvolvimento)
- Conta JumpCloud com permissões administrativas
- API Key do JumpCloud

## Instalação

### Terraform 0.13+

Para usar o provider, adicione a seguinte configuração no seu arquivo Terraform:

```hcl
terraform {
  required_providers {
    jumpcloud = {
      source = "ferreirafav/jumpcloud"
      version = "0.1.0"
    }
  }
}

provider "jumpcloud" {
  api_key = var.jumpcloud_api_key  # Ou defina a variável de ambiente JUMPCLOUD_API_KEY
  org_id  = var.jumpcloud_org_id   # Ou defina a variável de ambiente JUMPCLOUD_ORG_ID
}
```

## Exemplos de Uso

### Gerenciamento de Usuários

```hcl
resource "jumpcloud_user" "example" {
  username    = "example.user"
  email       = "example.user@example.com"
  firstname   = "Example"
  lastname    = "User"
  password    = "securePassword123!"
  description = "Created by Terraform"
  
  attributes = {
    department = "IT"
    location   = "Remote"
  }
  
  mfa_enabled          = true
  password_never_expires = false
}
```

### Gerenciamento de Sistemas

```hcl
resource "jumpcloud_system" "web_server" {
  display_name     = "web-server-01"
  description      = "Web server for production environment"
  
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = false
  allow_multi_factor_authentication = true
  
  tags = ["web", "production", "managed-by-terraform"]
}
```

### Associação entre Usuário e Sistema

```hcl
resource "jumpcloud_user_system_association" "admin_access" {
  user_id   = jumpcloud_user.admin.id
  system_id = jumpcloud_system.web_server.id
}
```

### Gerenciamento de Políticas

```hcl
resource "jumpcloud_policy" "password_complexity" {
  name        = "Secure Password Policy"
  description = "Política de complexidade de senha segura para todos os usuários"
  type        = "password_complexity"
  active      = true
  
  configurations = {
    min_length             = "12"
    requires_uppercase     = "true"
    requires_lowercase     = "true"
    requires_number        = "true"
    requires_special_char  = "true"
    password_expires_days  = "90"
    enable_password_expiry = "true"
  }
}

resource "jumpcloud_user_group" "finance" {
  name        = "Finance Department"
  description = "Grupo de usuários do departamento financeiro"
}

resource "jumpcloud_policy_association" "finance_password_policy" {
  policy_id = jumpcloud_policy.password_complexity.id
  group_id  = jumpcloud_user_group.finance.id
  type      = "user_group"
}
```

### Verificação de Associações e Criação Condicional

```hcl
# Verificar se um usuário tem acesso a um sistema
data "jumpcloud_user_system_association" "check_access" {
  user_id   = data.jumpcloud_user.existing_user.id
  system_id = data.jumpcloud_system.existing_system.id
}

# Criar associação apenas se ainda não existir
resource "jumpcloud_user_system_association" "conditional_access" {
  count = data.jumpcloud_user_system_association.check_access.associated ? 0 : 1
  
  user_id   = data.jumpcloud_user.existing_user.id
  system_id = data.jumpcloud_system.existing_system.id
}

# Verificar se uma política de MFA já existe
data "jumpcloud_policy" "existing_mfa" {
  name = "Required MFA Policy"
}

# Criar a política apenas se não existir
resource "jumpcloud_policy" "conditional_mfa" {
  count = data.jumpcloud_policy.existing_mfa.id != "" ? 0 : 1
  
  name        = "Required MFA Policy"
  description = "Política criada condicionalmente pelo Terraform"
  type        = "mfa"
  active      = true
  
  configurations = {
    require_mfa_for_all_users = "true"
  }
}
```

## Desenvolvimento

### Construção Local

1. Clone o repositório
2. Execute `go build` para compilar o provider
3. Execute o script `scripts/build.sh` para instalar o provider localmente

### Execução de Testes 

```
go test ./...
```

### Teste de Aceitação

```
TF_ACC=1 JUMPCLOUD_API_KEY=sua-api-key go test ./... -v
```

## Contribuição

Contribuições são bem-vindas! Por favor, siga estas etapas:

1. Fork o repositório
2. Crie um branch para sua contribuição (`git checkout -b feature/nome-da-feature`)
3. Faça as alterações necessárias
4. Execute os testes (`go test ./...`)
5. Envie um Pull Request

## Licença

Este projeto está licenciado sob a licença MIT - veja o arquivo LICENSE para mais detalhes.