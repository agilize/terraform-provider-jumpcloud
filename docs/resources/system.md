# jumpcloud_system Resource

Gerencia sistemas (dispositivos) no JumpCloud. Este recurso permite criar, atualizar e excluir configurações de sistemas no JumpCloud, controlando configurações de segurança, tags e atributos.

> **Nota:** O recurso `jumpcloud_system` gerencia a configuração de sistemas, mas não a criação do sistema em si. Os sistemas são registrados no JumpCloud quando o agente JumpCloud é instalado e se conecta ao servidor.

## Referência da API JumpCloud

Para mais detalhes sobre a API subjacente, consulte:
- [API JumpCloud - Sistemas](https://docs.jumpcloud.com/api/1.0/index.html#tag/systems)
- [Documentação do Agente JumpCloud](https://support.jumpcloud.com/s/article/jumpcloud-agent-deployment1)

## Considerações de Segurança

- Evite expor informações sensíveis diretamente na configuração do Terraform.
- Ao configurar a autenticação SSH, utilize políticas restritivas por padrão.
- Se possível, sempre habilite a autenticação multifator para aumentar a segurança.
- O gerenciamento centralizado de sistemas com Terraform facilita a aplicação de configurações de segurança consistentes.
- Para sistemas críticos, combine com políticas de JumpCloud para aplicar controles adicionais de segurança.
- Considere o uso de grupos de sistemas para aplicar políticas de segurança uniformes.
- Implemente a rotação regular de credenciais e chaves SSH nos sistemas gerenciados.

## Exemplos de Uso

### Configuração Básica de Sistema

```hcl
resource "jumpcloud_system" "basic_server" {
  display_name                      = "srv-app-prod-01"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = true
  allow_multi_factor_authentication = true
  description                       = "Servidor de aplicação gerenciado pelo Terraform"
  
  tags = [
    "production",
    "application"
  ]
}
```

### Sistema de Produção com Configurações Avançadas

```hcl
resource "jumpcloud_system" "prod_server" {
  display_name                      = "db-server-prod"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = false  # Apenas autenticação por chave
  allow_multi_factor_authentication = true
  description                       = "Servidor de banco de dados de produção com alta segurança"
  
  tags = [
    "production",
    "database",
    "critical"
  ]
  
  attributes = {
    environment    = "production"
    region         = "us-east"
    backup_enabled = "true"
    compliance     = "pci-dss,hipaa"
    owner          = "database-team"
  }
  
  # Configurações específicas do sistema
  ssh_root_enabled = false
  agent_bound      = true
}
```

### Múltiplos Sistemas com Diferentes Configurações

```hcl
resource "jumpcloud_system" "web_servers" {
  count = 3
  
  display_name                      = "web-server-${count.index + 1}"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = true
  allow_multi_factor_authentication = true
  description                       = "Servidor web ${count.index + 1} do cluster de produção"
  
  tags = [
    "production",
    "web-cluster",
    "nginx"
  ]
  
  attributes = {
    environment = "production"
    role        = "web-server"
    cluster     = "primary"
  }
}

resource "jumpcloud_system" "db_server" {
  display_name                      = "db-server-primary"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = false
  allow_multi_factor_authentication = true
  description                       = "Servidor de banco de dados principal"
  
  tags = [
    "production",
    "database",
    "mysql"
  ]
  
  attributes = {
    environment = "production"
    role        = "database"
    backup      = "enabled"
  }
}
```

### Gerenciamento Baseado em Ambientes

```hcl
# Definir variáveis para diferentes ambientes
locals {
  environments = {
    production = {
      allow_ssh_root_login              = false
      allow_ssh_password_authentication = false
      allow_multi_factor_authentication = true
      tags                              = ["production", "managed-by-terraform", "high-security"]
      security_level                    = "strict"
    }
    staging = {
      allow_ssh_root_login              = false
      allow_ssh_password_authentication = true
      allow_multi_factor_authentication = true
      tags                              = ["staging", "managed-by-terraform", "medium-security"]
      security_level                    = "normal"
    }
    development = {
      allow_ssh_root_login              = false
      allow_ssh_password_authentication = true
      allow_multi_factor_authentication = false
      tags                              = ["development", "managed-by-terraform", "low-security"]
      security_level                    = "relaxed"
    }
  }
  
  # Selecionar configuração baseada na variável de ambiente
  env = var.environment
}

# Criar sistema com configuração baseada no ambiente
resource "jumpcloud_system" "environment_based_server" {
  display_name                      = "app-server-${local.env}"
  allow_ssh_root_login              = local.environments[local.env].allow_ssh_root_login
  allow_ssh_password_authentication = local.environments[local.env].allow_ssh_password_authentication
  allow_multi_factor_authentication = local.environments[local.env].allow_multi_factor_authentication
  description                       = "Servidor de aplicação para ambiente ${local.env}"
  
  tags = local.environments[local.env].tags
  
  attributes = {
    environment    = local.env
    security_level = local.environments[local.env].security_level
    managed_by     = "terraform"
    team           = "devops"
  }
}
```

### Integração com Gerenciamento de Usuários

```hcl
# Gerenciar um sistema
resource "jumpcloud_system" "secure_system" {
  display_name                      = "secure-server-01"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = false
  allow_multi_factor_authentication = true
  description                       = "Servidor seguro para dados sensíveis"
  
  tags = ["secure", "restricted-access"]
  
  attributes = {
    data_classification = "confidential"
    access_control     = "strict"
  }
}

# Gerenciar um usuário administrador
resource "jumpcloud_user" "admin_user" {
  username  = "admin.user"
  email     = "admin@example.com"
  firstname = "Admin"
  lastname  = "User"
  
  attributes = {
    role = "system-administrator"
  }
}

# Associar o usuário ao sistema (requer recurso de associação)
resource "jumpcloud_user_system_association" "admin_access" {
  user_id   = jumpcloud_user.admin_user.id
  system_id = jumpcloud_system.secure_system.id
}
```

### Configuração de Classificações de Segurança

```hcl
# Definir classificações de segurança
locals {
  security_classifications = {
    tier_1 = {
      description = "Sistemas críticos com dados altamente sensíveis"
      config = {
        allow_ssh_root_login              = false
        allow_ssh_password_authentication = false
        allow_multi_factor_authentication = true
        ssh_root_enabled                  = false
        agent_bound                       = true
      }
      tags = ["tier-1", "critical", "high-security"]
      attributes = {
        security_level       = "maximum"
        patch_frequency      = "weekly"
        backup_frequency     = "daily"
        audit_logging        = "verbose"
        compliance           = "pci-dss,hipaa,soc2"
        data_classification  = "restricted"
      }
    }
    tier_2 = {
      description = "Sistemas importantes com dados sensíveis"
      config = {
        allow_ssh_root_login              = false
        allow_ssh_password_authentication = false
        allow_multi_factor_authentication = true
        ssh_root_enabled                  = false
        agent_bound                       = true
      }
      tags = ["tier-2", "important", "medium-security"]
      attributes = {
        security_level       = "high"
        patch_frequency      = "bi-weekly"
        backup_frequency     = "daily"
        audit_logging        = "standard"
        compliance           = "soc2"
        data_classification  = "confidential"
      }
    }
    tier_3 = {
      description = "Sistemas padrão com dados internos"
      config = {
        allow_ssh_root_login              = false
        allow_ssh_password_authentication = true
        allow_multi_factor_authentication = true
        ssh_root_enabled                  = false
        agent_bound                       = true
      }
      tags = ["tier-3", "standard", "basic-security"]
      attributes = {
        security_level       = "standard"
        patch_frequency      = "monthly"
        backup_frequency     = "weekly"
        audit_logging        = "basic"
        compliance           = "internal"
        data_classification  = "internal"
      }
    }
  }
}

# Aplicar classificação a um sistema
resource "jumpcloud_system" "classified_system" {
  display_name = "finance-app-server"
  description  = "Servidor de aplicações financeiras - ${local.security_classifications.tier_1.description}"
  
  # Aplicar configurações da classificação de segurança
  allow_ssh_root_login              = local.security_classifications.tier_1.config.allow_ssh_root_login
  allow_ssh_password_authentication = local.security_classifications.tier_1.config.allow_ssh_password_authentication
  allow_multi_factor_authentication = local.security_classifications.tier_1.config.allow_multi_factor_authentication
  ssh_root_enabled                  = local.security_classifications.tier_1.config.ssh_root_enabled
  agent_bound                       = local.security_classifications.tier_1.config.agent_bound
  
  # Aplicar tags da classificação
  tags = concat(local.security_classifications.tier_1.tags, ["finance", "erp"])
  
  # Combinar atributos base com atributos específicos da aplicação
  attributes = merge(local.security_classifications.tier_1.attributes, {
    application = "financial-erp"
    owner       = "finance-department"
    region      = "us-east"
    environment = "production"
  })
}
```

## Ciclo de Vida Completo com Automação

Para implementar um ciclo de vida completo para seus sistemas com JumpCloud e Terraform, considere o seguinte fluxo:

1. **Provisionamento da Infraestrutura**: Use Terraform com outros providers (AWS, Azure, etc.) para criar a infraestrutura base.

```hcl
# Exemplo de provisionamento EC2 + configuração JumpCloud
resource "aws_instance" "web_server" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
  
  # Script de user data para instalar o agente JumpCloud
  user_data = <<-EOF
    #!/bin/bash
    curl --tlsv1.2 --silent --show-error --header 'x-connect-key: ${var.jumpcloud_connect_key}' https://kickstart.jumpcloud.com/Kickstart | sudo bash
  EOF
  
  tags = {
    Name = "web-server-01"
  }
}

# Configuração do sistema no JumpCloud (após instalação do agente)
resource "jumpcloud_system" "web_server" {
  display_name                      = "web-server-01"
  allow_ssh_root_login              = false
  allow_ssh_password_authentication = false
  allow_multi_factor_authentication = true
  
  tags = [
    "production",
    "web",
    "managed-by-terraform"
  ]
  
  # Aguarda a criação da instância EC2 antes de configurar no JumpCloud
  depends_on = [aws_instance.web_server]
}
```

2. **Integração com Ferramentas de Configuração**: Combine Terraform com ferramentas como Ansible ou Chef para configuração interna.

```hcl
# Após configurar no JumpCloud, execute Ansible para configuração
resource "null_resource" "configure_web_server" {
  # Gatilho para executar apenas quando o sistema JumpCloud for modificado
  triggers = {
    jumpcloud_system_id = jumpcloud_system.web_server.id
  }
  
  provisioner "local-exec" {
    command = "ansible-playbook -i '${aws_instance.web_server.public_ip},' webserver-config.yml"
  }
  
  depends_on = [jumpcloud_system.web_server]
}
```

3. **Integração com CI/CD para Deploy Contínuo**:

```hcl
# Configure um webhook para integração com CI/CD
resource "jumpcloud_webhook" "deployment_hook" {
  name        = "deployment-webhook"
  description = "Webhook para updates de sistemas via CI/CD"
  url         = "https://jenkins.example.com/webhook/jumpcloud"
  enabled     = true
  
  # Tipos de eventos para acionar o webhook
  events = [
    "system.created",
    "system.updated"
  ]
}
```

## Referência de Argumentos

Os seguintes argumentos são suportados:

* `display_name` - (Obrigatório) O nome de exibição para o sistema no console JumpCloud.
* `allow_ssh_root_login` - (Opcional) Define se o login SSH como root é permitido. O padrão é `false`.
* `allow_ssh_password_authentication` - (Opcional) Define se a autenticação por senha no SSH é permitida. O padrão é `true`. Defina como `false` para permitir apenas autenticação por chave SSH.
* `allow_multi_factor_authentication` - (Opcional) Define se a autenticação multifator é permitida. O padrão é `false`. Recomenda-se habilitar para maior segurança.
* `tags` - (Opcional) Uma lista de tags para o sistema. As tags ajudam na organização e podem ser usadas para definir grupos.
* `description` - (Opcional) Uma descrição detalhada do sistema e seu propósito.
* `attributes` - (Opcional) Um mapa de atributos personalizados para o sistema. Útil para armazenar metadados customizados.
* `agent_bound` - (Opcional) Define se o sistema está vinculado a um agente JumpCloud. O padrão é `false`.
* `ssh_root_enabled` - (Opcional) Define se o login SSH como root está habilitado para este sistema específico. O padrão é `false`.
* `organization_id` - (Opcional) O ID da organização à qual o sistema pertence. Útil em ambientes multi-tenant.

## Referência de Atributos

Além de todos os argumentos acima, os seguintes atributos são exportados:

* `id` - O ID único do sistema no JumpCloud.
* `system_type` - O tipo do sistema (ex: "linux", "windows", "mac").
* `os` - O sistema operacional do sistema.
* `version` - A versão do sistema operacional.
* `agent_version` - A versão do agente JumpCloud instalado no sistema.
* `created` - A data de criação do registro do sistema no JumpCloud.
* `updated` - A data da última atualização do sistema.
* `hostname` - O nome do host do sistema.
* `fde_enabled` - Indica se a criptografia de disco completo (FDE) está ativada.
* `remote_ip` - O endereço IP remoto do sistema.
* `active` - Indica se o sistema está ativo no JumpCloud.
* `last_contact` - A data e hora do último contato do sistema com o JumpCloud.

## Importação

Sistemas existentes podem ser importados usando o ID, por exemplo:

```bash
terraform import jumpcloud_system.example 5f0c1b2c3d4e5f6g7h8i9j0k
```

Para importação em massa, considere usar scripts auxiliares:

```bash
#!/bin/bash
# Script para importar múltiplos sistemas JumpCloud

# Arquivo contendo IDs de sistemas no formato:
# RESOURCE_NAME,SYSTEM_ID
IMPORT_FILE="systems_to_import.csv"

while IFS=, read -r resource_name system_id
do
  echo "Importando $resource_name com ID $system_id..."
  terraform import "jumpcloud_system.$resource_name" "$system_id"
done < "$IMPORT_FILE"

echo "Importação concluída!"
```

## Gerenciamento de Estado e Ciclo de Vida

Ao gerenciar sistemas com Terraform, é crucial entender como o estado do Terraform interage com o estado real dos sistemas na JumpCloud:

1. **Criação vs. Configuração**: Lembre-se que o recurso `jumpcloud_system` configura um sistema já existente (registrado pelo agente) e não cria o próprio sistema físico.

2. **Cuidado com Deleções**: Remover um sistema da configuração Terraform fará com que a associação com o JumpCloud seja removida, mas o sistema físico continuará existindo.

3. **Alterações Externas**: Mudanças feitas fora do Terraform (via console JumpCloud) podem criar discrepâncias com o estado do Terraform.

4. **Recomendação**: Execute `terraform plan` regularmente para detectar e corrigir drift na configuração.

## Boas Práticas Avançadas

1. **Segurança e Padronização**:
   - Utilize variáveis Terraform para definir padrões de configuração consistentes.
   - Sempre desative o login SSH como root quando possível.
   - Habilite a autenticação multifator para aumentar a segurança.
   - Crie módulos Terraform para padronizar configurações de sistemas por tipo ou função.
   - Implemente verificação de compliance como parte do seu pipeline de CI/CD.

2. **Organização**:
   - Use tags de forma consistente para facilitar a gestão e a aplicação de políticas.
   - Documente o propósito de cada sistema no campo `description`.
   - Agrupe sistemas relacionados usando a mesma nomenclatura e sistema de tags.
   - Utilize workspaces do Terraform para separar ambientes (dev, staging, prod).

3. **Automação Completa**:
   - Utilize este recurso em conjunto com outras ferramentas de infraestrutura como código para gerenciar todo o ciclo de vida do sistema.
   - Considere integrar com ferramentas de configuração como Ansible ou Chef para gerenciar a configuração interna dos sistemas.
   - Implemente hooks de validação pré-commit para garantir qualidade da configuração.
   - Use backends remotos para o estado do Terraform com bloqueio de estado.

4. **Monitoramento**:
   - A gestão de sistemas com Terraform facilita a criação de dashboards padronizados para monitoramento.
   - Considere integrar com ferramentas de logging centralizadas para rastrear atividades em todos os sistemas.
   - Desenvolva alertas baseados em atributos e tags configurados via Terraform.
   - Implemente auditoria automatizada da configuração de segurança.

5. **Governança**:
   - Defina políticas claras sobre quem pode modificar configurações de sistemas.
   - Documente todas as exceções às políticas de segurança padrão.
   - Realize revisões periódicas das configurações de sistemas críticos.
   - Considere implementar sentinel ou outras ferramentas de policy-as-code. 