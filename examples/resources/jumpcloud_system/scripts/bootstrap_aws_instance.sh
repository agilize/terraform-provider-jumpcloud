#!/bin/bash
#
# Script de bootstrap para instalar o agente JumpCloud em instâncias AWS
# Este script pode ser usado como user-data para AWS EC2 ou em instâncias existentes
#
# Uso:
#   ./bootstrap_aws_instance.sh <JUMPCLOUD_CONNECT_KEY> [SISTEMA_TAGS]
#
# Exemplo:
#   ./bootstrap_aws_instance.sh a1b2c3d4e5f6g7h8i9j0 "production,web-server,nginx"

# Verificar se a chave de conexão foi fornecida
if [ -z "$1" ]; then
    echo "Erro: JUMPCLOUD_CONNECT_KEY é obrigatório"
    echo "Uso: $0 <JUMPCLOUD_CONNECT_KEY> [SISTEMA_TAGS]"
    exit 1
fi

# Obter parâmetros
CONNECT_KEY="$1"
TAGS="${2:-}"

# Gravar log de instalação
LOGFILE="/var/log/jumpcloud_install.log"
exec > >(tee -a ${LOGFILE}) 2>&1

echo "$(date '+%Y-%m-%d %H:%M:%S') Iniciando instalação do agente JumpCloud..."

# Identificar informações do sistema
INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
REGION=$(curl -s http://169.254.169.254/latest/meta-data/placement/region)
HOSTNAME=$(hostname)

echo "$(date '+%Y-%m-%d %H:%M:%S') Informações do sistema:"
echo "  - Instance ID: $INSTANCE_ID"
echo "  - Region: $REGION"
echo "  - Hostname: $HOSTNAME"

# Detectar o sistema operacional
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=${ID}
    VERSION=${VERSION_ID}
    echo "$(date '+%Y-%m-%d %H:%M:%S') Sistema operacional: $OS $VERSION"
else
    echo "$(date '+%Y-%m-%d %H:%M:%S') Sistema operacional não identificado. Tentando instalar assim mesmo."
    OS="unknown"
    VERSION="unknown"
fi

# Configurar o nome do host se estiver em AWS
if [[ ! -z "$INSTANCE_ID" ]]; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') Configurando nome do host baseado no Instance ID..."
    hostnamectl set-hostname "aws-$INSTANCE_ID" || echo "Falha ao configurar hostname"
fi

# Instalar o agente JumpCloud
echo "$(date '+%Y-%m-%d %H:%M:%S') Instalando agente JumpCloud..."
echo "$(date '+%Y-%m-%d %H:%M:%S') Usando chave de conexão: ${CONNECT_KEY:0:5}*****"

# Baixar e executar o instalador JumpCloud
echo "$(date '+%Y-%m-%d %H:%M:%S') Baixando e executando o instalador JumpCloud..."
curl --tlsv1.2 --silent --show-error --header "x-connect-key: $CONNECT_KEY" https://kickstart.jumpcloud.com/Kickstart | sudo bash

# Verificar status da instalação
if [ $? -eq 0 ]; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') Agente JumpCloud instalado com sucesso!"
else
    echo "$(date '+%Y-%m-%d %H:%M:%S') Falha na instalação do agente JumpCloud. Verifique o log para mais detalhes."
    exit 1
fi

# Aguardar o agente iniciar e se registrar
echo "$(date '+%Y-%m-%d %H:%M:%S') Aguardando o agente iniciar e se registrar na JumpCloud (30s)..."
sleep 30

# Verificar o status do serviço
if command -v systemctl &> /dev/null; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') Status do serviço JumpCloud:"
    systemctl status jcagent.service
fi

# Definir metadados personalizados para o sistema
echo "$(date '+%Y-%m-%d %H:%M:%S') Configurando metadados para o sistema..."

# Definir tags no formato JumpCloud usando o utilitário jq se disponível
if [ ! -z "$TAGS" ] && command -v jq &> /dev/null; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') Configurando tags: $TAGS"
    
    # Este é um exemplo, na prática isso será feito através do Terraform
    # e da API do JumpCloud, não através do agente local
    echo "Tags serão configuradas através do Terraform"
fi

# Informações adicionais para AWS
if [[ ! -z "$INSTANCE_ID" ]]; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') Adicionando metadados AWS..."
    
    # Nota: Na prática, esses metadados seriam configurados através do
    # recurso jumpcloud_system no Terraform, não localmente
    echo "Metadados AWS serão configurados através do Terraform"
fi

echo "$(date '+%Y-%m-%d %H:%M:%S') Instalação completa do agente JumpCloud!"
echo "$(date '+%Y-%m-%d %H:%M:%S') Agora você pode gerenciar este sistema na console JumpCloud"
echo "$(date '+%Y-%m-%d %H:%M:%S') ou usando o provider Terraform para JumpCloud."
echo ""
echo "Log de instalação disponível em: $LOGFILE"
exit 0 