#!/usr/bin/env bash
set -e

# Cores para melhor legibilidade
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Verificando requisitos para publicação no Terraform Registry...${NC}"

# Função para verificar um arquivo ou diretório
check_file() {
  if [ -f "$1" ] || [ -d "$1" ]; then
    echo -e "  [${GREEN}✓${NC}] $2 encontrado: $1"
    return 0
  else
    echo -e "  [${RED}✗${NC}] $2 não encontrado: $1"
    return 1
  }
}

# Verificar estrutura básica do repositório
echo -e "\n${YELLOW}1. Verificando estrutura básica do repositório:${NC}"
ERRORS=0

check_file "terraform-registry-manifest.json" "Arquivo de manifesto" || ((ERRORS++))
check_file "docs/index.md" "Documentação principal" || ((ERRORS++))
check_file "docs/resources" "Diretório de documentação de recursos" || ((ERRORS++))
check_file "docs/data-sources" "Diretório de documentação de data sources" || ((ERRORS++))
check_file ".github/workflows/release.yml" "Workflow de release" || ((ERRORS++))
check_file ".goreleaser.yml" "Configuração do GoReleaser" || ((ERRORS++))

# Verificar manifesto do registry
echo -e "\n${YELLOW}2. Verificando conteúdo do manifesto:${NC}"
if [ -f "terraform-registry-manifest.json" ]; then
  if grep -q '"protocol_versions": \[\("5.0"\|"6.0"\)' terraform-registry-manifest.json; then
    echo -e "  [${GREEN}✓${NC}] Versões de protocolo definidas corretamente"
  else
    echo -e "  [${RED}✗${NC}] Versões de protocolo não definidas corretamente"
    ((ERRORS++))
  fi
fi

# Verificar namespace no README e exemplos
echo -e "\n${YELLOW}3. Verificando namespace nos exemplos:${NC}"
NAMESPACE="registry.terraform.io/agilize/jumpcloud"
FILES_WITH_INCORRECT_NAMESPACE=0

# Verificar no README
if grep -q "source *= *\"$NAMESPACE\"" README.md; then
  echo -e "  [${GREEN}✓${NC}] Namespace correto no README.md"
else
  echo -e "  [${RED}✗${NC}] Namespace incorreto no README.md"
  ((ERRORS++))
fi

# Verificar nos exemplos
if [ -d "examples" ]; then
  EXAMPLE_FILES=$(find examples -name "*.tf" | wc -l)
  CORRECT_FILES=$(grep -l "source *= *\"$NAMESPACE\"" $(find examples -name "*.tf") | wc -l)
  
  if [ "$EXAMPLE_FILES" -eq 0 ]; then
    echo -e "  [${YELLOW}!${NC}] Nenhum arquivo de exemplo encontrado"
  elif [ "$CORRECT_FILES" -eq "$EXAMPLE_FILES" ]; then
    echo -e "  [${GREEN}✓${NC}] Namespace correto em todos os $EXAMPLE_FILES arquivos de exemplo"
  else
    FILES_WITH_INCORRECT_NAMESPACE=$((EXAMPLE_FILES - CORRECT_FILES))
    echo -e "  [${RED}✗${NC}] $FILES_WITH_INCORRECT_NAMESPACE de $EXAMPLE_FILES arquivos de exemplo não usam o namespace correto"
    ((ERRORS++))
  fi
fi

# Verificar configuração do GoReleaser para publicação
echo -e "\n${YELLOW}4. Verificando configuração do GoReleaser:${NC}"
if [ -f ".goreleaser.yml" ]; then
  if grep -q "signs:" .goreleaser.yml; then
    echo -e "  [${GREEN}✓${NC}] Configuração de assinatura encontrada no GoReleaser"
  else
    echo -e "  [${RED}✗${NC}] Configuração de assinatura não encontrada no GoReleaser"
    ((ERRORS++))
  fi
  
  if grep -q "terraform-registry" .goreleaser.yml; then
    echo -e "  [${GREEN}✓${NC}] Configuração para Terraform Registry encontrada"
  else
    echo -e "  [${RED}✗${NC}] Configuração para Terraform Registry não encontrada"
    ((ERRORS++))
  fi
fi

# Verificar workflow do GitHub Actions
echo -e "\n${YELLOW}5. Verificando workflow do GitHub Actions:${NC}"
if [ -f ".github/workflows/release.yml" ]; then
  if grep -q "GPG_PRIVATE_KEY" .github/workflows/release.yml; then
    echo -e "  [${GREEN}✓${NC}] Configuração da chave GPG encontrada no workflow"
  else
    echo -e "  [${RED}✗${NC}] Configuração da chave GPG não encontrada no workflow"
    ((ERRORS++))
  fi
fi

# Resumo
echo -e "\n${YELLOW}Resumo da verificação:${NC}"
if [ $ERRORS -eq 0 ]; then
  echo -e "${GREEN}✅ O provider está pronto para publicação no Terraform Registry!${NC}"
  echo -e "   Continue com os seguintes passos:"
  echo -e "   1. Certifique-se de ter adicionado sua chave GPG pública no Terraform Registry"
  echo -e "   2. Crie uma tag no formato 'vX.Y.Z' para iniciar o processo de release"
  echo -e "   3. Após a conclusão da release, publique o provider no Terraform Registry"
  exit 0
else
  echo -e "${RED}❌ Foram encontrados $ERRORS problemas que precisam ser corrigidos antes da publicação.${NC}"
  exit 1
fi 