# Exemplo de Importação de Sistemas JumpCloud Existentes
# Demonstração de como trazer sistemas existentes para o gerenciamento via Terraform

provider "jumpcloud" {
  # Recomendado usar variáveis de ambiente para credenciais
  # JUMPCLOUD_API_KEY e JUMPCLOUD_ORG_ID
}

###############################################
# DEFINIÇÃO DO RECURSO PARA IMPORTAÇÃO
###############################################

# Para importar um sistema existente, primeiro defina um recurso vazio ou parcial
# com as configurações que você pretende gerenciar
resource "jumpcloud_system" "imported_server" {
  # É necessário definir o atributo display_name, outras configurações
  # serão preenchidas pelo estado do Terraform após a importação
  display_name = "servidor-legacy-01"
  
  # Você pode adicionar outras configurações que deseja gerenciar,
  # ou deixar o recurso mínimo e completar após a importação
  
  # IMPORTANTE: Após a importação, compare com o estado real do sistema
  # antes de aplicar qualquer mudança
}

###############################################
# COMANDOS DE IMPORTAÇÃO
###############################################

# Após definir o recurso, execute o comando de importação:
#
# terraform import jumpcloud_system.imported_server SYSTEM_ID
#
# Onde SYSTEM_ID é o ID do sistema na JumpCloud que você deseja importar.
# Você pode obter esse ID através do console da JumpCloud ou da API.

###############################################
# EXEMPLO DE FLUXO DE TRABALHO DE IMPORTAÇÃO
###############################################

# 1. Execute "terraform init" para inicializar o provider
# 2. Crie este arquivo com a definição do recurso a ser importado
# 3. Execute o comando de importação:
#    terraform import jumpcloud_system.imported_server 5f8a7b6c5d4e3f2a1b0c9d8e
# 4. Execute "terraform state show jumpcloud_system.imported_server" para ver os atributos
# 5. Complete o arquivo de configuração com os valores desejados
# 6. Execute "terraform plan" para validar as alterações
# 7. Execute "terraform apply" para aplicar as alterações

###############################################
# EXEMPLO DE IMPORTAÇÃO EM MASSA
###############################################

# Para importar múltiplos sistemas, você pode usar um script como este:
#
# #!/bin/bash
# # Script para importar múltiplos sistemas JumpCloud para Terraform
#
# # Lista de sistemas no formato "NOME_RECURSO:ID_SISTEMA"
# SYSTEMS=(
#   "web_server_1:5f8a7b6c5d4e3f2a1b0c9d8e"
#   "db_server_1:6f7a8b9c0d1e2f3a4b5c6d7e"
#   "app_server_1:7f8a9b0c1d2e3f4a5b6c7d8e"
# )
#
# # Importar cada sistema
# for system in "${SYSTEMS[@]}"; do
#   RESOURCE_NAME=$(echo $system | cut -d':' -f1)
#   SYSTEM_ID=$(echo $system | cut -d':' -f2)
#   echo "Importando $RESOURCE_NAME com ID $SYSTEM_ID..."
#   terraform import "jumpcloud_system.$RESOURCE_NAME" "$SYSTEM_ID"
# done
#
# echo "Importação concluída. Verifique o estado com 'terraform state list'"

###############################################
# EXEMPLO DE RECURSO APÓS IMPORTAÇÃO COMPLETA
###############################################

# Após importar e verificar, o recurso ficaria assim (exemplo):

# resource "jumpcloud_system" "imported_server" {
#   display_name                      = "servidor-legacy-01"
#   allow_ssh_root_login              = true
#   allow_ssh_password_authentication = true
#   allow_multi_factor_authentication = false
#   description                       = "Servidor legado importado para Terraform"
#   
#   tags = [
#     "legacy",
#     "imported"
#   ]
#   
#   attributes = {
#     environment = "legacy"
#     region      = "us-west"
#     role        = "application"
#     status      = "to-be-upgraded"
#   }
# } 