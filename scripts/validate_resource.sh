#!/bin/bash

# Script de Validação de Recursos do Terraform Provider JumpCloud
# Uso: ./scripts/validate_resource.sh <resource_file_path>

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para imprimir com cor
print_status() {
    local status=$1
    local message=$2
    
    case $status in
        "success")
            echo -e "${GREEN}✓${NC} $message"
            ;;
        "error")
            echo -e "${RED}✗${NC} $message"
            ;;
        "warning")
            echo -e "${YELLOW}⚠${NC} $message"
            ;;
        "info")
            echo -e "${BLUE}ℹ${NC} $message"
            ;;
    esac
}

# Função para verificar se arquivo existe
check_file_exists() {
    local file=$1
    if [ ! -f "$file" ]; then
        print_status "error" "Arquivo não encontrado: $file"
        exit 1
    fi
    print_status "success" "Arquivo encontrado: $file"
}

# Função para verificar estrutura básica do recurso
check_resource_structure() {
    local file=$1
    local errors=0
    
    print_status "info" "Verificando estrutura do recurso..."
    
    # Verificar se tem CreateContext
    if grep -q "CreateContext:" "$file"; then
        print_status "success" "CreateContext encontrado"
    else
        print_status "error" "CreateContext não encontrado"
        ((errors++))
    fi
    
    # Verificar se tem ReadContext
    if grep -q "ReadContext:" "$file"; then
        print_status "success" "ReadContext encontrado"
    else
        print_status "error" "ReadContext não encontrado"
        ((errors++))
    fi
    
    # Verificar se tem UpdateContext
    if grep -q "UpdateContext:" "$file"; then
        print_status "success" "UpdateContext encontrado"
    else
        print_status "warning" "UpdateContext não encontrado (pode ser intencional)"
    fi
    
    # Verificar se tem DeleteContext
    if grep -q "DeleteContext:" "$file"; then
        print_status "success" "DeleteContext encontrado"
    else
        print_status "error" "DeleteContext não encontrado"
        ((errors++))
    fi
    
    # Verificar se tem Schema
    if grep -q "Schema:" "$file"; then
        print_status "success" "Schema encontrado"
    else
        print_status "error" "Schema não encontrado"
        ((errors++))
    fi
    
    return $errors
}

# Função para verificar uso de client
check_client_usage() {
    local file=$1
    local errors=0
    
    print_status "info" "Verificando uso do client..."
    
    # Verificar se usa GetClientFromMeta ou ConvertToClientInterface
    if grep -q "GetClientFromMeta\|ConvertToClientInterface" "$file"; then
        print_status "success" "Client obtido corretamente"
    else
        print_status "error" "Client não está sendo obtido corretamente"
        ((errors++))
    fi
    
    # Verificar se usa DoRequest
    if grep -q "DoRequest" "$file"; then
        print_status "success" "DoRequest encontrado"
    else
        print_status "error" "DoRequest não encontrado"
        ((errors++))
    fi
    
    return $errors
}

# Função para verificar tratamento de erros
check_error_handling() {
    local file=$1
    local errors=0
    
    print_status "info" "Verificando tratamento de erros..."
    
    # Verificar se usa IsNotFoundError
    if grep -q "IsNotFoundError" "$file"; then
        print_status "success" "IsNotFoundError encontrado"
    else
        print_status "warning" "IsNotFoundError não encontrado (recomendado no Read)"
    fi
    
    # Verificar se remove do state quando não encontrado
    if grep -q 'd.SetId("")' "$file"; then
        print_status "success" "Remoção do state implementada"
    else
        print_status "warning" "Remoção do state não encontrada"
    fi
    
    # Verificar se usa diag.FromErr
    if grep -q "diag.FromErr" "$file"; then
        print_status "success" "diag.FromErr encontrado"
    else
        print_status "error" "diag.FromErr não encontrado"
        ((errors++))
    fi
    
    return $errors
}

# Função para verificar logging
check_logging() {
    local file=$1
    
    print_status "info" "Verificando logging..."
    
    # Verificar se usa tflog
    if grep -q "tflog.Debug\|tflog.Warn\|tflog.Error" "$file"; then
        print_status "success" "Logging implementado"
    else
        print_status "warning" "Logging não encontrado (recomendado)"
    fi
}

# Função para verificar campos sensíveis
check_sensitive_fields() {
    local file=$1
    
    print_status "info" "Verificando campos sensíveis..."
    
    # Procurar por campos que deveriam ser sensíveis
    local sensitive_keywords=("password" "secret" "token" "api_key" "private_key" "certificate")
    local found_sensitive=false
    
    for keyword in "${sensitive_keywords[@]}"; do
        if grep -qi "\"$keyword\"" "$file"; then
            found_sensitive=true
            # Verificar se está marcado como Sensitive
            if grep -A 5 "\"$keyword\"" "$file" | grep -q "Sensitive.*true"; then
                print_status "success" "Campo '$keyword' marcado como sensível"
            else
                print_status "warning" "Campo '$keyword' pode precisar ser marcado como sensível"
            fi
        fi
    done
    
    if [ "$found_sensitive" = false ]; then
        print_status "info" "Nenhum campo sensível óbvio encontrado"
    fi
}

# Função para verificar validações
check_validations() {
    local file=$1
    
    print_status "info" "Verificando validações..."
    
    # Verificar se usa ValidateFunc
    if grep -q "ValidateFunc:" "$file"; then
        print_status "success" "ValidateFunc encontrado"
    else
        print_status "info" "ValidateFunc não encontrado (pode não ser necessário)"
    fi
    
    # Verificar se usa ConflictsWith
    if grep -q "ConflictsWith:" "$file"; then
        print_status "success" "ConflictsWith encontrado"
    else
        print_status "info" "ConflictsWith não encontrado (pode não ser necessário)"
    fi
}

# Função para verificar descrições
check_descriptions() {
    local file=$1
    local errors=0
    
    print_status "info" "Verificando descrições..."
    
    # Contar campos sem descrição
    local fields_without_desc=$(grep -c "Type:.*schema.Type" "$file" || true)
    local fields_with_desc=$(grep -c "Description:" "$file" || true)
    
    if [ $fields_with_desc -gt 0 ]; then
        print_status "success" "Descrições encontradas ($fields_with_desc campos)"
    else
        print_status "warning" "Nenhuma descrição encontrada"
    fi
}

# Função para verificar se existe teste
check_tests() {
    local file=$1
    local test_file="${file%%.go}_test.go"
    
    print_status "info" "Verificando testes..."
    
    if [ -f "$test_file" ]; then
        print_status "success" "Arquivo de teste encontrado: $test_file"
        
        # Verificar se tem testes básicos
        if grep -q "func Test" "$test_file"; then
            local test_count=$(grep -c "func Test" "$test_file")
            print_status "success" "Testes encontrados: $test_count"
        else
            print_status "warning" "Nenhum teste encontrado no arquivo"
        fi
    else
        print_status "warning" "Arquivo de teste não encontrado: $test_file"
    fi
}

# Função para verificar documentação
check_documentation() {
    local file=$1
    local resource_name=$(basename "$file" | sed 's/resource_//' | sed 's/data_source_//' | sed 's/.go$//')
    
    print_status "info" "Verificando documentação..."
    
    # Verificar se existe exemplo
    local example_dir="examples/resources"
    if echo "$file" | grep -q "data_source"; then
        example_dir="examples/data-sources"
    fi
    
    if [ -d "$example_dir" ]; then
        print_status "info" "Diretório de exemplos existe: $example_dir"
    else
        print_status "warning" "Diretório de exemplos não encontrado: $example_dir"
    fi
}

# Função principal
main() {
    local file=$1
    
    if [ -z "$file" ]; then
        echo "Uso: $0 <resource_file_path>"
        echo "Exemplo: $0 jumpcloud/users/users_directory/resource_user.go"
        exit 1
    fi
    
    echo ""
    echo "========================================="
    echo "  Validação de Recurso Terraform"
    echo "========================================="
    echo ""
    
    check_file_exists "$file"
    echo ""
    
    local total_errors=0
    
    check_resource_structure "$file"
    total_errors=$((total_errors + $?))
    echo ""
    
    check_client_usage "$file"
    total_errors=$((total_errors + $?))
    echo ""
    
    check_error_handling "$file"
    total_errors=$((total_errors + $?))
    echo ""
    
    check_logging "$file"
    echo ""
    
    check_sensitive_fields "$file"
    echo ""
    
    check_validations "$file"
    echo ""
    
    check_descriptions "$file"
    echo ""
    
    check_tests "$file"
    echo ""
    
    check_documentation "$file"
    echo ""
    
    echo "========================================="
    if [ $total_errors -eq 0 ]; then
        print_status "success" "Validação concluída sem erros críticos!"
    else
        print_status "error" "Validação concluída com $total_errors erro(s) crítico(s)"
        exit 1
    fi
    echo "========================================="
    echo ""
}

# Executar
main "$@"

