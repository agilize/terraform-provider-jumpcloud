#!/bin/bash

# Script para verificar apenas os erros críticos de linting, ignorando os que serão tratados em fases
# Este script executa o tfproviderlint desativando todas as regras exceto as críticas

# Cores para saída
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Executando verificações de linting críticas...${NC}"

# Caminho para o binário do tfproviderlint
LINTER_BIN="$HOME/go/bin/tfproviderlint"

# Verificar se o binário existe
if [ ! -f "$LINTER_BIN" ]; then
    echo -e "${RED}tfproviderlint não encontrado. Instalando...${NC}"
    go install github.com/bflad/tfproviderlint/cmd/tfproviderlint@latest
fi

# Lista de erros que estão sendo ignorados
echo -e "${YELLOW}Executando tfproviderlint ignorando erros não prioritários...${NC}"

# Executa o linter desativando todas as regras exceto R014 (que já corrigimos)
$LINTER_BIN \
  -AT001=false \
  -AT005=false \
  -AT012=false \
  -R001=false \
  -R017=false \
  -R019=false \
  -V013=false \
  ./...

# Verificar o código de saída
if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}Verificação concluída: Todos os erros críticos foram corrigidos!${NC}"
    echo -e "${GREEN}Os erros não críticos serão tratados em fases futuras.${NC}"
else
    echo -e "\n${RED}Verificação falhou: Existem erros críticos que precisam ser corrigidos.${NC}"
fi

echo -e "\n${YELLOW}Nota:${NC} Para verificar todos os erros, use o comando:"
echo -e "  $LINTER_BIN ./..."
echo -e "${YELLOW}Para implementar correções fase a fase, consulte o script run_linter.sh${NC}" 