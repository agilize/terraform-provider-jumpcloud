#!/bin/bash

# Script para executar o tfproviderlint com configurações personalizadas
# Permite ignorar temporariamente certos tipos de erros enquanto são corrigidos em fases

# Cores para saída
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Executando verificações lint personalizadas...${NC}"

# Caminho para o binário do tfproviderlint
LINTER_BIN="$HOME/go/bin/tfproviderlint"

# Verificar se o binário existe
if [ ! -f "$LINTER_BIN" ]; then
    echo -e "${RED}tfproviderlint não encontrado. Instalando...${NC}"
    go install github.com/bflad/tfproviderlint/cmd/tfproviderlint@latest
fi

# Verificar R014 (erros que já corrigimos)
echo -e "${YELLOW}Verificando regra R014 (interface{} deve ser chamado de 'meta')...${NC}"
$LINTER_BIN -R014=true ./...

# Se desejar verificar outros erros específicos, descomente e execute
# Por exemplo, para verificar apenas R001:
# echo -e "${YELLOW}Verificando regra R001...${NC}"
# $LINTER_BIN -R001=true -AT=false -R=false -S=false -V=false ./...

# Lista de erros que estamos ignorando temporariamente
echo -e "\n${YELLOW}Erros temporariamente ignorados (a serem corrigidos em fases):${NC}"
echo -e "- ${YELLOW}AT001${NC}: missing CheckDestroy"
echo -e "- ${YELLOW}AT005${NC}: acceptance test function name should begin with TestAcc"
echo -e "- ${YELLOW}AT012${NC}: file contains multiple acceptance test name prefixes"
echo -e "- ${YELLOW}R001${NC}: ResourceData.Set() key argument should be string literal"
echo -e "- ${YELLOW}R017${NC}: schema attributes should be stable across Terraform runs"
echo -e "- ${YELLOW}R019${NC}: d.HasChanges() has many arguments, consider d.HasChangesExcept()"
echo -e "- ${YELLOW}V013${NC}: custom SchemaValidateFunc should be replaced with validation.StringInSlice()"

echo -e "\n${GREEN}Para verificar um erro específico, execute:${NC}"
echo -e "  $LINTER_BIN -AT=false -R=false -S=false -V=false -<REGRA>=true ./..."
echo -e "  Exemplo: $LINTER_BIN -AT=false -R=false -S=false -V=false -R001=true ./..."

echo -e "\n${GREEN}Para corrigir todos os erros, execute:${NC}"
echo -e "  $LINTER_BIN ./..."

# Exemplo de próximos passos para correção
echo -e "\n${YELLOW}Plano sugerido para correção em fases:${NC}"
echo -e "1. Corrigir R001 (ResourceData.Set com string literal)"
echo -e "2. Corrigir R019 (HasChanges → HasChangesExcept)"
echo -e "3. Corrigir V013 (SchemaValidateFunc → validation.StringInSlice)"
echo -e "4. Corrigir R017 (Schema attributes should be stable)"
echo -e "5. Corrigir AT* (problemas em testes de aceitação)" 