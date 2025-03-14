# Exemplos do Recurso jumpcloud_system

Este diretório contém exemplos de uso do recurso `jumpcloud_system` do Provider Terraform para JumpCloud.

## Visão Geral

O recurso `jumpcloud_system` permite gerenciar a configuração de sistemas (dispositivos) no JumpCloud através do Terraform. Use estes exemplos como ponto de partida para suas próprias configurações.

## Exemplos Disponíveis

### resource.tf

Exemplo básico de configuração de sistemas JumpCloud, incluindo:
- Configuração do provider JumpCloud
- Configuração de um servidor web
- Configuração de múltiplos servidores web em cluster
- Configuração de um servidor de banco de dados
- Outputs para os IDs dos sistemas

### security_focused.tf

Exemplo avançado de configuração com ênfase em segurança, incluindo:
- Servidor com segurança aprimorada para aplicações financeiras
- Configuração de um host bastion para acesso seguro
- Uso detalhado de atributos para documentar configurações de segurança
- Outputs com flag sensitive para dados sensíveis

### import.tf

Exemplo de como importar sistemas JumpCloud existentes para o Terraform:
- Definição de recursos para importação
- Comandos de exemplo para importação
- Fluxo de trabalho de importação passo a passo
- Script de exemplo para importação em massa

## Scripts Auxiliares

### scripts/bootstrap_aws_instance.sh

Script para provisionar o agente JumpCloud em instâncias AWS:
- Detecta automaticamente metadados da instância
- Instala o agente JumpCloud
- Configura nome de host baseado no ID da instância
- Gera logs detalhados do processo de instalação

## Como Usar Estes Exemplos

1. **Prepare seu ambiente:**
   ```bash
   # Configure as variáveis de ambiente para autenticação
   export JUMPCLOUD_API_KEY="sua-api-key"
   export JUMPCLOUD_ORG_ID="seu-org-id"
   ```

2. **Inicialize o Terraform:**
   ```bash
   terraform init
   ```

3. **Adapte os exemplos** às suas necessidades, modificando variáveis, tags, atributos e outras configurações.

4. **Valide suas configurações:**
   ```bash
   terraform validate
   terraform plan
   ```

5. **Aplique a configuração:**
   ```bash
   terraform apply
   ```

## Importante

- O recurso `jumpcloud_system` configura sistemas existentes, não cria sistemas físicos ou instala o agente JumpCloud.
- O agente JumpCloud deve estar instalado e registrado antes que o Terraform possa gerenciar o sistema.
- Use o script `bootstrap_aws_instance.sh` ou similar para provisionar o agente em novas instâncias.

## Ciclo de Vida Típico

1. Provisione infraestrutura (ex: instância EC2, VM Azure, etc.)
2. Instale o agente JumpCloud na instância (usando scripts de bootstrap ou ferramentas de configuração)
3. Use o Terraform com o recurso `jumpcloud_system` para configurar o sistema no JumpCloud
4. Gerencie usuários, grupos e políticas do JumpCloud usando outros recursos do provider

## Recursos Relacionados

- Documentação do recurso: `terraform-provider-jumpcloud/docs/resources/system.md`
- API JumpCloud: [https://docs.jumpcloud.com/api/](https://docs.jumpcloud.com/api/)
- Documentação do JumpCloud: [https://support.jumpcloud.com/](https://support.jumpcloud.com/) 